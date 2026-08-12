package live

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_sign_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	robotconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// Service 管理 B站 直播 WebSocket 监听器的完整生命周期
//
// 与项目中其他 Service（admin/role/menu）不同，本 Service 是**有状态**的：
//   - 持有长生命周期的 *bilibili.Client（管理 Cookie/Session）
//   - 持有 *live.Listener（管理 WebSocket 后台 goroutine）
//   - 使用 sync.Mutex 保护所有可变状态
//
// 典型用法（在其他 Service 中通过 DI 获取）：
//
//	s.liveSvc.UpdateRoom(ctx, 22384516)
//	s.liveSvc.StartListener(ctx)
//	status, _, _ := s.liveSvc.GetListenerStatus()
type Service struct {
	mu        sync.Mutex
	client    *bilibili.Client // B站 API 客户端（长生命周期，管理 Cookie）
	listener  *live.Listener   // WebSocket 监听器（nil = 未启动）
	roomID    int64            // 目标房间号
	stateFile string           // Cookie 持久化文件路径
	// listener 生命周期控制
	listenerCtx    context.Context
	listenerCancel context.CancelFunc
	// 消息统计（由后台 goroutine 更新，mu 保护）
	stats listenerStats
	// 原始消息日志
	rawLogger *logger.RawMessageLogger
	// procDone 在消息处理 goroutine 退出后关闭，供 StopListener 等待
	procDone chan struct{}
	// 弹幕发送
	roomSvc *room.Service // 弹幕 API 服务
	queue   *live.Queue   // 弹幕发送优先级队列（nil = 未创建）
	// 直播间状态缓存
	roomState *RoomState
	// 自动广告定时发送器
	robotConfigSvc   *robotconfigsvc.Service // 机器人配置服务
	autoSenderCancel context.CancelFunc      // 自动发送器取消函数
	autoSenderDone   chan struct{}           // 自动发送器退出信号
	// 前端 WebSocket 推送
	hub *Hub // 前端消息推送中心
	// 消息业务处理器分发器
	dispatcher *messageDispatcher
	// 数据访问
	liveDanmuRepo         live_danmu.Repository
	liveGiftRepo          live_gift.Repository
	liveSessionRepo       live_session.Repository
	LiveUserBlacklistRepo live_user_blacklist.Repository
}

// New 创建直播服务
//
// bilibili.Client 在此创建并持久化（整个服务生命周期内复用）
// WithStateFile 会在启动时自动恢复之前保存的登录态
func New(cfg config.LiveConfig, robotConfigSvc *robotconfigsvc.Service, configCache *robotconfig.Cache, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveSessionRepo live_session.Repository, liveUserRepo live_user.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, liveUserSignLogRepo live_user_sign_log.Repository, LiveUserBlacklistRepo live_user_blacklist.Repository) *Service {
	client := bilibili.NewClient(
		bilibili.WithStateFile(cfg.StateFile),
	)
	hub := newHub()
	go hub.Run(context.Background())
	roomState := &RoomState{}
	var enqueueFn func(msg string, priority int)
	dispatcher := newMessageDispatcher(
		newLiveStatusProcessor(liveSessionRepo, liveDanmuRepo, liveGiftRepo, roomState),
		newLiveEndProcessor(liveSessionRepo, liveDanmuRepo, liveGiftRepo, roomState),
		newGiftProcessor(liveGiftRepo),
		newInteractProcessor(liveUserRepo),
		newDanmuProcessor(liveUserRepo, liveDanmuRepo, liveGiftRepo, liveUserCreditLogRepo, liveUserSignLogRepo, LiveUserBlacklistRepo, roomState, configCache, client, func() int64 {
			if sess := client.Session(); sess != nil {
				return sess.UID
			}
			return 0
		}, func(msg string, priority int) {
			if enqueueFn != nil {
				enqueueFn(msg, priority)
			}
		}),
		newPkProcessor(),
	)
	// 从配置中恢复上次监听的房间号
	defaultRoomID := robotConfigSvc.GetRoomID()
	s := &Service{
		client:          client,
		stateFile:       cfg.StateFile,
		roomID:          defaultRoomID,
		roomState:       roomState,
		rawLogger:       logger.NewRawMessageLogger(),
		roomSvc:         room.NewService(client),
		hub:             hub,
		dispatcher:      dispatcher,
		liveDanmuRepo:   liveDanmuRepo,
		liveGiftRepo:    liveGiftRepo,
		liveSessionRepo: liveSessionRepo,
		robotConfigSvc:  robotConfigSvc,
	}
	enqueueFn = s.EnqueueDanmu
	return s
}

// Client 返回内部的 *bilibili.Client，供其他 Service 进行高级操作
func (s *Service) Client() *bilibili.Client {
	return s.client
}

// GetQRCode 获取 B站 扫码登录二维码
func (s *Service) GetQRCode(ctx context.Context) (*QRCodeResp, int, error) {
	qr, err := s.client.Auth.GetQRCode(ctx)
	if err != nil {
		return nil, 60401, err
	}
	return &QRCodeResp{
		URL:       qr.URL,
		QrcodeKey: qr.QrcodeKey,
	}, 0, nil
}

// PollQRCode 轮询扫码状态
//
// 当 status.Code == 0（登录成功）时，自动完成：
//  1. 请求 redirectURL 种 Cookie 到 CookieJar
//  2. 获取用户信息（UID、用户名）
//  3. 获取 Buvid 设备指纹
//  4. 更新 client.Session
//  5. 持久化 state 到磁盘
func (s *Service) PollQRCode(ctx context.Context, qrcodeKey string) (*PollQRCodeResp, int, error) {
	status, err := s.client.Auth.PollQRCode(ctx, qrcodeKey)
	if err != nil {
		return nil, 60401, err
	}
	resp := &PollQRCodeResp{
		Status:    status.Code,
		Message:   status.Message,
		IsScanned: status.Code == 86090,
		IsSuccess: status.Code == 0,
		IsExpired: status.Code == 86038,
	}
	// 登录成功，完成后续流程
	if status.Code == 0 {
		// 请求 redirectURL 以通过 http.Client.Jar 自动捕获 Set-Cookie
		if err := s.client.Get(ctx, status.RedirectURL, nil); err != nil {
			return nil, 60401, fmt.Errorf("无法跳转到指定的重定向链接: %w”", err)
		}
		// 获取用户信息
		userInfo, err := s.client.Auth.GetUserInfo(ctx)
		if err != nil {
			return nil, 60401, fmt.Errorf("登录成功后无法获取用户资料: %w", err)
		}
		// 获取设备指纹
		buvidInfo, err := s.client.Auth.GetBuvid(ctx)
		if err != nil {
			return nil, 60401, fmt.Errorf("登录成功后无法获取 buvid: %w", err)
		}
		// 更新 session
		s.client.SetSession(&bilibili.Session{
			UID:      userInfo.UID,
			Username: userInfo.UName,
			Face:     userInfo.Face,
			Buvid:    buvidInfo.Buvid3,
		})
		// 持久化到磁盘
		if s.stateFile != "" {
			if err := s.client.SaveState(s.stateFile); err != nil {
				return nil, 60401, fmt.Errorf("save state after login: %w", err)
			}
		}
	}
	return resp, 0, nil
}

// GetLoginStatus 返回当前登录状态
func (s *Service) GetLoginStatus(ctx context.Context) (*LoginStatusResp, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := s.client.Session()
	if session == nil {
		return &LoginStatusResp{IsLoggedIn: false}, 0, nil
	}
	return &LoginStatusResp{
		IsLoggedIn: true,
		UID:        session.UID,
		Username:   session.Username,
		Face:       session.Face,
		Buvid:      session.Buvid,
	}, 0, nil
}

// Logout 清除 B站 登录态
//
// 注意：这仅清除本地 Cookie/Session，不会使 B站 服务端的 session 失效
// 同时删除 state 持久化文件
//
// 如果当前正在监听直播间，先停止监听
func (s *Service) Logout(ctx context.Context) (int, error) {
	// 如果正在监听，先停止
	if code, err := s.StopListener(); err != nil {
		return code, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// 清空 CookieJar 和 Session
	s.client = bilibili.NewClient()
	// 删除持久化文件
	if s.stateFile != "" {
		if err := os.Remove(s.stateFile); err != nil && !os.IsNotExist(err) {
			return 60401, fmt.Errorf("remove state file: %w", err)
		}
	}
	return 0, nil
}

// UpdateRoom 设置监听目标房间号
//
// 房间号的最终校验在 StartListener 中（通过 GetRealRoomID），此处仅做基本校验
//
// 如果当前正在监听直播间且房间号发生变化：
//  1. 停止当前监听
//  2. 更新房间号
//  3. 启动新房间的监听
func (s *Service) UpdateRoom(ctx context.Context, roomID int64) (int, error) {
	if roomID <= 0 {
		return 40404, nil
	}
	s.mu.Lock()
	oldRoomID := s.roomID
	isRunning := s.listener != nil && s.listener.IsRunning()
	s.mu.Unlock()
	// 相同房间号，无需操作
	if oldRoomID == roomID {
		return 0, nil
	}
	// 未在监听：直接更新房间号即可
	if !isRunning {
		s.mu.Lock()
		s.roomID = roomID
		s.mu.Unlock()
		// 持久化房间号
		if _, err := s.robotConfigSvc.SetRoomID(context.Background(), roomID); err != nil {
			log.Printf("[live.Service] 持久化房间号到配置失败: %v", err)
		}
		return 0, nil
	}
	// 正在监听且房间号变更：停旧 → 换号 → 启新
	if code, err := s.StopListener(); code != 0 {
		return code, err
	}
	s.mu.Lock()
	s.roomID = roomID
	s.mu.Unlock()
	// 持久化房间号
	if _, err := s.robotConfigSvc.SetRoomID(context.Background(), roomID); err != nil {
		log.Printf("[live.Service] 持久化房间号到配置失败: %v", err)
	}
	return s.StartListener(ctx)
}

// StartListener 创建并启动 WebSocket 监听器
//
// 前置条件：
//   - 已登录 B站（client.Session() != nil）
//   - 已设置有效房间号
//   - 监听器未在运行
//
// 启动流程：
//  1. 调用 live.NewListenerFromClient 自动获取连接信息
//  2. 调用 listener.Connect 建立 WebSocket
//  3. 启动后台 goroutine 处理消息并更新统计
func (s *Service) StartListener(ctx context.Context) (int, error) {
	s.mu.Lock()
	// 前置检查
	if s.listener != nil && s.listener.IsRunning() {
		s.mu.Unlock()
		return 40402, nil
	}
	if s.client.Session() == nil {
		s.mu.Unlock()
		return 40401, nil
	}
	if s.roomID <= 0 {
		s.mu.Unlock()
		return 40404, nil
	}
	roomID := s.roomID
	s.mu.Unlock()
	// 使用独立的 background context 创建 listener（不受 HTTP 请求 context 超时影响）
	listener, err := live.NewListenerFromClient(
		context.Background(),
		s.client,
		roomID,
		live.WithPingInterval(30*time.Second),
		live.WithMsgChannelSize(512),
	)
	if err != nil {
		return 40406, fmt.Errorf("创建监听器失败: %w", err)
	}
	// 创建独立 context 用于控制本次监听会话
	listenerCtx, listenerCancel := context.WithCancel(context.Background())
	if err := listener.Connect(listenerCtx); err != nil {
		listenerCancel()
		return 40406, fmt.Errorf("连接监听器失败: %w", err)
	}
	// 更新内部状态
	s.mu.Lock()
	s.listener = listener
	s.listenerCtx = listenerCtx
	s.listenerCancel = listenerCancel
	// 重置统计
	s.stats = listenerStats{
		startTime: time.Now(),
	}
	// 启动消息处理 goroutine
	s.procDone = make(chan struct{})
	procDone := s.procDone
	s.mu.Unlock()
	// 同步会话状态：清理非监听房间记录 + 根据实际直播状态修正数据
	s.syncSessionsOnStart(ctx, roomID)
	// 创建并启动弹幕发送队列
	s.queue = live.NewQueue(s.listener.RoomID(), &danmuSender{roomSvc: s.roomSvc, client: s.client, danmuLogger: logger.NewDanmuSendLogger()})
	s.queue.Start(listenerCtx)
	// 启动自动广告定时发送器
	s.startAutoSender(listenerCtx)
	go s.processMessages(listenerCtx, listener, procDone, s.rawLogger)
	return 0, nil
}

// StopListener 停止 WebSocket 监听器
//
// 停止流程：
//  1. 取消 listener context（触发 processMessages + listener.run 退出）
//  2. 调用 listener.Stop()（关闭 WebSocket，等待内部 goroutine）
//  3. 等待消息处理 goroutine 退出
func (s *Service) StopListener() (int, error) {
	s.mu.Lock()
	if s.listener == nil || !s.listener.IsRunning() {
		s.mu.Unlock()
		return 0, nil
	}
	cancel := s.listenerCancel
	listener := s.listener
	procDone := s.procDone
	queue := s.queue
	s.mu.Unlock()
	// 停止自动广告发送器
	s.stopAutoSender()
	// 停止弹幕发送队列并清空待发送消息
	if queue != nil {
		queue.Stop()
		queue.Clear()
	}
	// 取消 context — 触发所有 goroutine 退出
	cancel()
	// 停止 listener — 关闭 WebSocket，等待 run() goroutine 退出
	if err := listener.Stop(); err != nil {
		return 40407, err
	}
	// 等待消息处理 goroutine 退出
	<-procDone
	// 清理状态
	s.mu.Lock()
	s.listener = nil
	s.listenerCtx = nil
	s.listenerCancel = nil
	s.procDone = nil
	s.queue = nil
	s.autoSenderCancel = nil
	s.autoSenderDone = nil
	s.mu.Unlock()
	return 0, nil
}

// GetListenerStatus 返回监听器状态与消息统计
func (s *Service) GetListenerStatus(ctx context.Context) (*ListenerStatusResp, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	isRunning := s.listener != nil && s.listener.IsRunning()
	resp := &ListenerStatusResp{
		IsRunning: isRunning,
		RoomID:    s.roomID,
	}
	if s.roomID > 0 {
		roomInfo, err := s.client.Room.GetRealRoomInfo(ctx, s.roomID)
		if err != nil {
			return nil, 60401, fmt.Errorf("无法在线获取直播间信息: %w", err)
		}
		s.roomState.Update(roomInfo)
		resp.UID = roomInfo.UID
		resp.Title = roomInfo.Title
		resp.LiveStatus = roomInfo.LiveStatus
		resp.Online = roomInfo.Online
		resp.Attention = roomInfo.Attention
		resp.LiveTime = roomInfo.LiveTime
	}
	if isRunning {
		resp.MsgCount = s.stats.msgCount
		resp.DanmuCount = s.stats.danmuCount
		resp.GiftCount = s.stats.giftCount
		resp.StartTime = s.stats.startTime.Format("2006-01-02 15:04:05")
		resp.Uptime = time.Since(s.stats.startTime).Truncate(time.Second).String()
	}
	return resp, 0, nil
}

// EnqueueDanmu 将弹幕加入发送队列
//
// priority 越小越优先发送，同优先级按入队顺序（FIFO）发送。
// 即使监听器未启动也可以入队，消息会在 StartListener 后依次发送。
//
// 如果监听器未在运行（queue == nil），消息会被丢弃。
func (s *Service) EnqueueDanmu(msg string, priority int) {
	s.mu.Lock()
	q := s.queue
	s.mu.Unlock()
	if q != nil {
		q.Enqueue(msg, priority)
	}
}

// Hub 返回前端 WebSocket 消息推送中心，供 Handler 层注册客户端连接
func (s *Service) Hub() *Hub {
	return s.hub
}

// Shutdown 优雅停止服务（监听器 + 持久化），供 main.go 在进程退出前调用
//
// 如果监听器正在运行，先停止；然后保存 state 到磁盘
func (s *Service) Shutdown() {
	s.mu.Lock()
	isRunning := s.listener != nil && s.listener.IsRunning()
	s.mu.Unlock()
	if isRunning {
		// 忽略错误 — 关机时无法响应客户端，尽力而为
		_, _ = s.StopListener()
	}
	// 最终持久化
	if s.stateFile != "" && s.client.Session() != nil {
		_ = s.client.SaveState(s.stateFile)
	}
	// 关闭原始消息日志
	if s.rawLogger != nil {
		s.rawLogger.Close()
	}
	// 关闭所有前端 WebSocket 连接
	if s.hub != nil {
		s.hub.closeAllClients()
	}
}
