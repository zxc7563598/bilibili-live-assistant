package live

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// startAutoSender 读取广告配置，若启用则启动后台 ticker goroutine 定时发送广告弹幕
func (s *Service) startAutoSender(ctx context.Context) {
	s.mu.Lock()
	if s.autoSenderCancel != nil {
		s.mu.Unlock()
		return
	}
	// 创建子 context 用于控制本次自动发送会话
	autoCtx, autoCancel := context.WithCancel(ctx)
	s.autoSenderCancel = autoCancel
	s.autoSenderDone = make(chan struct{})
	done := s.autoSenderDone
	s.mu.Unlock()
	// 读取配置
	cfg, _, err := s.robotConfigSvc.GetAdConfig(context.Background())
	if err != nil {
		log.Printf("[live.AutoSender] 读取广告配置失败: %v, 自动发送器未启动", err)
		s.cleanupAutoSender(autoCancel, done)
		return
	}
	if cfg.Enabled != "1" {
		log.Printf("[live.AutoSender] 广告配置未启用(enabled=%s), 自动发送器未启动", cfg.Enabled)
		s.cleanupAutoSender(autoCancel, done)
		return
	}
	if len(cfg.Content) == 0 {
		log.Printf("[live.AutoSender] 广告内容为空, 自动发送器未启动")
		s.cleanupAutoSender(autoCancel, done)
		return
	}
	// 解析发送间隔
	intervalSec, err := strconv.Atoi(cfg.Interval)
	if err != nil || intervalSec <= 0 {
		intervalSec = 60
	}
	interval := time.Duration(intervalSec) * time.Second
	log.Printf("[live.AutoSender] 自动广告发送器已启动，间隔=%v, 模式=%s, 场景=%s, 内容条数=%d",
		interval, cfg.SendMode, cfg.Scene, len(cfg.Content))

	go s.runAutoSender(autoCtx, done, cfg.Scene, cfg.SendMode, cfg.Content, interval)
}

// stopAutoSender 停止自动广告发送器并等待 goroutine 退出
func (s *Service) stopAutoSender() {
	s.mu.Lock()
	cancel := s.autoSenderCancel
	done := s.autoSenderDone
	s.autoSenderCancel = nil
	s.autoSenderDone = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
		log.Printf("[live.AutoSender] 自动广告发送器已停止")
	}
}

// cleanupAutoSender 清理 autoSender 状态（配置不满足启动条件时调用）
func (s *Service) cleanupAutoSender(cancel context.CancelFunc, done chan struct{}) {
	cancel()
	close(done)
	s.mu.Lock()
	s.autoSenderCancel = nil
	s.autoSenderDone = nil
	s.mu.Unlock()
}

// RestartAutoSender 重启自动广告发送器（stop → 重新读配置 → start）
func (s *Service) RestartAutoSender() {
	s.mu.Lock()
	isRunning := s.listener != nil && s.listener.IsRunning()
	listenerCtx := s.listenerCtx
	s.mu.Unlock()

	if !isRunning || listenerCtx == nil {
		return
	}
	s.stopAutoSender()
	s.startAutoSender(listenerCtx)
}

// runAutoSender 自动广告发送器的主循环，在独立 goroutine 中运行
func (s *Service) runAutoSender(ctx context.Context, done chan struct{}, scene, sendMode string, content []string, interval time.Duration) {
	defer close(done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var seqIndex int
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.checkAdSend(scene) {
				continue
			}
			msg := pickAdMessage(sendMode, content, &seqIndex)
			if msg != "" {
				// 定时任务, priority = 50
				s.EnqueueDanmu(msg, 50)
			}
		}
	}
}

// checkAdScene 根据配置检查当前是否满足发送场景
func (s *Service) checkAdSend(scene string) bool {
	sceneValue := ptr.ParseEnumInt[enum.Scene](scene)
	switch sceneValue {
	case enum.SceneLive, enum.SceneNotLive:
		liveStatus := s.roomState.LiveStatus()
		if liveStatus == -1 {
			s.mu.Lock()
			roomID := s.roomID
			s.mu.Unlock()
			roomInfo, err := s.client.Room.GetRealRoomInfo(context.Background(), roomID)
			if err != nil {
				log.Printf("[live.AutoSender] 获取直播间状态失败: %v", err)
				return false
			}
			s.roomState.Update(roomInfo)
			liveStatus = roomInfo.LiveStatus
		}
		isLive := enum.Scene(liveStatus) == enum.SceneLive
		if sceneValue == enum.SceneLive {
			return isLive
		}
		return !isLive
	default:
		return true
	}
}

// pickAdMessage 根据发送模式从内容列表中选取一条消息
func pickAdMessage(sendMode string, content []string, seqIndex *int) string {
	if len(content) == 0 {
		return ""
	}
	sendModeValue := ptr.ParseEnumInt[enum.SendMode](sendMode)
	switch sendModeValue {
	case enum.SendModeSequential:
		idx := *seqIndex % len(content)
		*seqIndex++
		return content[idx]
	case enum.SendModeRandom:
		return content[rand.Intn(len(content))]
	default:
		return ""
	}
}
