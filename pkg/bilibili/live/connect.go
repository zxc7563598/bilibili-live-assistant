package live

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
)

// NewListenerFromClient 创建一个已配置好连接信息的 Listener
//
// 它自动完成所有连接前准备工作：
//   - 解析真实房间号（短 ID → 长 ID）
//   - 获取 WBI 签名密钥
//   - 获取弹幕 WebSocket 连接信息（Host、Token、WSS 端口）
//   - 提取用户认证信息（uid、buvid3）
//   - 构建 WebSocket 握手 Headers（Cookie、User-Agent 等）
//
// 返回的 Listener 尚未连接，调用 Connect(ctx) 建立 WebSocket 连接并开始监听
//
// 典型用法：
//
//	client := bilibili.NewClient(bilibili.WithStateFile("state.json"))
//	listener, err := live.NewListenerFromClient(ctx, client, 22384516,
//	    live.WithOnConnected(func() { fmt.Println("已连接") }),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	listener.Connect(ctx)
//	defer listener.Stop()
//
//	for msg := range listener.Messages() {
//	    switch msg.Cmd {
//	    case live.CmdDanmuMsg:
//	        // 处理弹幕
//	    }
//	}
func NewListenerFromClient(ctx context.Context, client *bilibili.Client, roomID int64, opts ...ListenerOption) (*Listener, error) {
	// 解析真实房间号
	realRoomID, err := client.Room.GetRealRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("live.NewListenerFromClient: 获取真实房间号失败: %w", err)
	}
	// 获取 WBI 签名密钥
	wbiKeys, err := client.Auth.GetWbiKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("live.NewListenerFromClient: 获取 WBI 签名密钥失败: %w", err)
	}
	// 获取弹幕 WebSocket 连接信息
	danmuInfo, err := client.Room.GetDanmuInfo(ctx, realRoomID, wbiKeys.ImgKey, wbiKeys.SubKey)
	if err != nil {
		return nil, fmt.Errorf("live.NewListenerFromClient: 获取弹幕 WebSocket 连接信息失败: %w", err)
	}
	// 提取用户认证信息
	session := client.Session()
	var uid int64
	var buvid3 string
	if session != nil {
		uid = session.UID
		buvid3 = session.Buvid
	}
	// 从 API 获取最新的 buvid（可能比 session 中的更新）
	buvidInfo, err := client.Auth.GetBuvid(ctx)
	if err == nil && buvidInfo != nil && buvidInfo.Buvid3 != "" {
		buvid3 = buvidInfo.Buvid3
		client.SetSession(&bilibili.Session{
			UID:      session.UID,
			Username: session.Username,
			Buvid:    buvid3,
		})
	}
	// 构建 WebSocket 握手 Headers
	cookieURL, _ := url.Parse("https://" + danmuInfo.Host)
	cookieStr := client.CookieJar().CookieString(cookieURL)
	headers := map[string][]string{
		"Cookie":          {cookieStr},
		"User-Agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"Origin":          {"https://live.bilibili.com"},
		"Pragma":          {"no-cache"},
		"Cache-Control":   {"no-cache"},
		"Accept-Encoding": {"gzip, deflate, br, zstd"},
		"Accept-Language": {"zh-CN,zh;q=0.9"},
	}
	// 合并选项：WithAuth 放在前面（作为默认值），用户的 opts 可覆盖
	allOpts := append([]ListenerOption{WithAuth(uid, buvid3)}, opts...)
	// 创建 Listener 并存储预获取的连接信息
	listener := NewListener(realRoomID, allOpts...)
	listener.danmuInfo = danmuInfo
	listener.wsHeaders = headers
	return listener, nil
}
