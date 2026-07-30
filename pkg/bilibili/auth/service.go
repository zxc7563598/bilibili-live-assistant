// Package auth 提供 B站 登录/鉴权相关的 API 调用
// 本包不导入父包 bilibili，通过自有的 HttpClient 接口实现依赖反转
package auth

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/api"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/wbi"
)

// Service 提供 B站 认证相关的 API 调用，包括扫码登录、用户信息获取、WBI 密钥和设备指纹
type Service struct {
	client HttpClient
}

// NewService 创建认证服务实例
// client 需实现 HttpClient 接口（Get/Post），由父包 *bilibili.Client 自动满足
func NewService(client HttpClient) *Service {
	return &Service{client: client}
}

// =========================================================================
// 登录二维码
// =========================================================================

// QRCode 登录二维码信息
type QRCode struct {
	URL       string `json:"url"`        // 二维码图片 URL，可用于生成二维码供用户扫描
	QrcodeKey string `json:"qrcode_key"` // 二维码唯一标识 key，用于后续轮询扫码状态
}

type qrCodeResponse struct {
	api.Response
	Data QRCode `json:"data"`
}

// GetQRCode 获取登录二维码
//
// 调用 B站 /x/passport-login/web/qrcode/generate 接口
// 无需登录态
//
// 返回的 QRCode 包含：
//   - URL: 二维码图片 URL，可生成二维码供用户扫描
//   - QrcodeKey: 二维码标识 key，用于后续调用 PollQRCode 轮询扫码状态
func (s *Service) GetQRCode(ctx context.Context) (*QRCode, error) {
	var resp qrCodeResponse
	if err := s.client.Get(ctx, api.EndpointQRCodeGenerate, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// =========================================================================
// 扫码状态轮询
// =========================================================================

// QRCodeStatus 二维码扫码状态
type QRCodeStatus struct {
	Code         int    `json:"code"`          // 扫码状态码：86101=未扫码, 86090=已扫码未确认, 0=已确认登录, 86038=已过期
	Message      string `json:"message"`       // 状态描述信息
	RedirectURL  string `json:"url"`           // 登录确认后的跳转 URL，用于种 Cookie
	RefreshToken string `json:"refresh_token"` // 刷新令牌，扫码成功时附带
}

type qrCodePollResponse struct {
	api.Response
	Data QRCodeStatus `json:"data"`
}

// PollQRCode 轮询扫码状态
//
// 调用 B站 /x/passport-login/web/qrcode/poll 接口
// qrcodeKey 为 GetQRCode 返回的 QrcodeKey
//
// 返回的 QRCodeStatus.Code 含义：
//   - 86101：等待扫码（需继续轮询）
//   - 86090：已扫码，等待用户确认（需继续轮询）
//   - 0：已确认登录成功，RedirectURL 可用于种 Cookie
//   - 86038：二维码已过期，需重新调用 GetQRCode
func (s *Service) PollQRCode(ctx context.Context, qrcodeKey string) (*QRCodeStatus, error) {
	path := api.EndpointQRCodePoll + qrcodeKey
	var resp qrCodePollResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// =========================================================================
// 当前登录用户信息
// =========================================================================

// UserInfo 当前登录用户基本信息（从 nav 接口获取）
type UserInfo struct {
	IsLogin bool   `json:"isLogin"` // 是否已登录
	UID     int64  `json:"mid"`     // 用户 UID
	UName   string `json:"uname"`   // 用户名
	Face    string `json:"face"`    // 头像 URL
}

type userInfoResponse struct {
	api.Response
	Data UserInfo `json:"data"`
}

// GetUserInfo 获取当前登录用户的 nav 信息
//
// 调用 B站 /x/web-interface/nav 接口
// 需要 Cookie 登录态（CookieJar 中需有 SESSDATA 等有效 Cookie）
//
// 返回的 UserInfo.IsLogin 为 true 表示登录态有效
func (s *Service) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	var resp userInfoResponse
	if err := s.client.Get(ctx, api.EndpointNav, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// =========================================================================
// WBI 签名密钥
// =========================================================================

// WbiKeys WBI 签名密钥对，从 nav 接口的 wbi_img 字段中提取
// imgKey 和 subKey 用于 wbi.Sign() 计算接口签名
type WbiKeys struct {
	ImgKey string // 图片密钥，从 wbi_img.img_url 文件名中提取（不含扩展名）
	SubKey string // 子密钥，从 wbi_img.sub_url 文件名中提取（不含扩展名）
}

type wbiNavResponse struct {
	api.Response
	Data struct {
		WbiImg struct {
			ImgURL string `json:"img_url"` // WBI 图片 URL，如 https://i0.hdslb.com/bfs/wbi/xxx.png
			SubURL string `json:"sub_url"` // WBI 子 URL，如 https://i0.hdslb.com/bfs/wbi/yyy.png
		} `json:"wbi_img"`
	} `json:"data"`
}

// GetWbiKeys 获取 WBI 签名密钥对
//
// 调用 B站 /x/web-interface/nav 接口，从 wbi_img 字段提取 imgKey 和 subKey
// 返回的密钥对用于 wbi.Sign() 计算需要 WBI 签名的接口参数
//
// 使用 wbi.ExtractKeys 从 wbi_img URL 中提取文件名（不含扩展名）作为 key
func (s *Service) GetWbiKeys(ctx context.Context) (*WbiKeys, error) {
	var resp wbiNavResponse
	if err := s.client.Get(ctx, api.EndpointNav, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	imgKey, subKey := wbi.ExtractKeys(resp.Data.WbiImg.ImgURL, resp.Data.WbiImg.SubURL)
	return &WbiKeys{ImgKey: imgKey, SubKey: subKey}, nil
}

// =========================================================================
// 设备指纹（Buvid）
// =========================================================================

// BuvidInfo 设备指纹信息，用于唯一标识客户端设备
type BuvidInfo struct {
	Buvid3 string // Buvid3 设备标识，前端常用
	Buvid4 string // Buvid4 设备标识，部分接口使用
}

type buvidResponse struct {
	api.Response
	Data struct {
		B3 string `json:"b_3"` // Buvid3，如 "XX-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
		B4 string `json:"b_4"` // Buvid4，Base64 编码的设备指纹
	} `json:"data"`
}

// GetBuvid 获取设备指纹 Buvid3/Buvid4
//
// 调用 B站 /x/frontend/finger/spi 接口
// 无需 Cookie 登录态，可用于首次请求建立设备标识
//
// Buvid3 建议写入 CookieJar 中的 buvid3 Cookie，以便后续请求自动携带
func (s *Service) GetBuvid(ctx context.Context) (*BuvidInfo, error) {
	var resp buvidResponse
	if err := s.client.Get(ctx, api.EndpointFingerSpy, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &BuvidInfo{Buvid3: resp.Data.B3, Buvid4: resp.Data.B4}, nil
}
