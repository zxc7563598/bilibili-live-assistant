// B站 扫码登录工具。
//
// 运行方式:
//
//	go run ./test/login/
//
// 流程：获取二维码 → 立即轮询 + 每 5 秒轮询扫码状态 → 登录成功后访问跳转 URL 种 Cookie → 保存状态。
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
)

const stateFile = "bilibili_state.json"

func main() {
	fmt.Println("========== B站 扫码登录 ==========")
	fmt.Println()
	// 尝试加载已有状态
	client := bilibili.NewClient(bilibili.WithStateFile(stateFile))
	if client.Session() != nil {
		fmt.Printf("检测到已有登录态: UID=%d, 用户名=%s\n", client.Session().UID, client.Session().Username)
		fmt.Print("是否重新登录？(y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("已取消，使用现有登录态。")
			return
		}
		// 创建全新客户端（不加载旧状态）
		client = bilibili.NewClient()
	}
	ctx := context.Background()
	// 获取二维码
	fmt.Println("正在获取登录二维码...")
	qr, err := client.Auth.GetQRCode(ctx)
	if err != nil {
		fmt.Printf("获取二维码失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Printf("┌──────────────────────────────────────────────────┐\n")
	fmt.Printf("│  请使用 B站 App 扫描下方二维码登录               │\n")
	fmt.Printf("│                                                  │\n")
	fmt.Printf("│  二维码链接（可复制到浏览器打开）:                │\n")
	fmt.Printf("│  %s  │\n", qr.URL)
	fmt.Printf("│                                                  │\n")
	fmt.Printf("│  轮询密钥: %s                       │\n", qr.QrcodeKey)
	fmt.Printf("└──────────────────────────────────────────────────┘\n")
	fmt.Println()
	// 轮询扫码状态，首次立即执行，之后每 5 秒一次
	fmt.Println("等待扫码中...（每 5 秒轮询一次，Ctrl+C 取消）")
	fmt.Println()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	// poll 执行单次轮询，返回 true 表示登录流程结束（成功或失败）
	poll := func() bool {
		status, err := client.Auth.PollQRCode(ctx, qr.QrcodeKey)
		if err != nil {
			fmt.Printf("  [%s] 轮询失败: %v\n", time.Now().Format("15:04:05"), err)
			return false
		}
		switch status.Code {
		case 86101:
			fmt.Printf("  [%s] 等待扫码...\n", time.Now().Format("15:04:05"))
		case 86090:
			fmt.Printf("  [%s] 已扫码，请在手机上点击确认！\n", time.Now().Format("15:04:05"))
		case 0:
			fmt.Println()
			fmt.Println("========== 扫码确认成功，正在获取登录信息... ==========")
			// 访问跳转 URL 种 Cookie
			if status.RedirectURL != "" {
				if err := visitRedirectURL(ctx, client, status.RedirectURL); err != nil {
					fmt.Printf("访问跳转 URL 失败（Cookie 可能未正确设置）: %v\n", err)
				} else {
					fmt.Println("Cookie 已获取")
				}
			}
			// 获取用户信息
			userInfo, err := client.Auth.GetUserInfo(ctx)
			if err != nil {
				fmt.Printf("获取用户信息失败: %v\n", err)
			} else {
				fmt.Printf("用户: %s (UID: %d)\n", userInfo.UName, userInfo.UID)
				buvid := ""
				if s := client.Session(); s != nil {
					buvid = s.Buvid
				}
				client.SetSession(&bilibili.Session{
					UID:      userInfo.UID,
					Username: userInfo.UName,
					Face:     userInfo.Face,
					Buvid:    buvid,
				})
			}
			// 保存状态
			if err := client.SaveState(stateFile); err != nil {
				fmt.Printf("保存状态失败: %v\n", err)
			} else {
				fmt.Printf("Cookie 已保存到: %s\n", stateFile)
			}
			// 显示保存的 Cookie 摘要
			fmt.Println()
			fmt.Println("已保存的 Cookie:")
			cj := client.CookieJar()
			for _, domain := range []string{
				"https://api.bilibili.com",
				"https://live.bilibili.com",
				"https://passport.bilibili.com",
			} {
				u, _ := url.Parse(domain)
				cookieStr := cj.CookieString(u)
				if cookieStr != "" {
					fmt.Printf("  [%s]\n    %s\n", domain, cookieStr)
				}
			}
			fmt.Println()
			fmt.Println("登录完成！下次启动将自动恢复登录态。")
			fmt.Println("====================================================")
			return true
		case 86038:
			fmt.Println()
			fmt.Println("二维码已过期，请重新运行本程序。")
			return true
		default:
			fmt.Printf("  [%s] 状态码: %d, 消息: %s\n",
				time.Now().Format("15:04:05"), status.Code, status.Message)
		}
		return false
	}
	// 立即执行首次轮询
	if poll() {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-sigCh:
			fmt.Println("\n已取消。")
			return
		case <-ticker.C:
			if poll() {
				return
			}
		}
	}
}

// visitRedirectURL 访问扫码成功后的跳转 URL，使 http.Client 的 CookieJar 自动捕获 Set-Cookie。
func visitRedirectURL(ctx context.Context, client *bilibili.Client, redirectURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
