// Package room 提供 B站 直播间相关的 API 调用
// 本包不导入父包 bilibili，通过自有的 HttpClient 接口实现依赖反转
package room

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/api"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/wbi"
)

// Service 提供 B站 直播间相关的 API 调用，包括直播间信息、弹幕、排行榜、大航海、禁言管理等
type Service struct {
	client HttpClient
}

// NewService 创建直播间服务实例
// client 需实现 HttpClient 接口（Get/Post/PostForm），由父包 *bilibili.Client 自动满足
func NewService(client HttpClient) *Service {
	return &Service{client: client}
}

// =========================================================================
// 直播间基本信息
// =========================================================================

// RealRoomInfo 直播间详细信息
type RealRoomInfo struct {
	UID        int64  `json:"uid"`         // 主播 UID
	RoomID     int64  `json:"room_id"`     // 真实房间号（长 ID）
	Title      string `json:"title"`       // 直播间标题
	LiveStatus int    `json:"live_status"` // 直播状态：0=未开播, 1=直播中, 2=轮播中
	Online     int    `json:"online"`      // 在线观众数（人气值，非真实人数）
	Attention  int    `json:"attention"`   // 关注数
	LiveTime   string `json:"live_time"`   // 开播时间，格式如 "2025-01-01 12:00:00"
	Keyframe   string `json:"keyframe"`    // 直播间封面图 URL
}

type realRoomInfoResponse struct {
	api.Response
	Data RealRoomInfo `json:"data"`
}

// GetRealRoomID 获取直播间真实房间号，短 ID 会被解析为长 ID
//
// 调用 B站 /room/v1/Room/get_info 接口
// 无需 Cookie 登录态
//
// 参数 roomID 可以是短房间号（如 6 位数字），返回的是真实的完整房间号
// 如果传入的已是真实房间号或 API 返回的 RoomID 为 0，则原样返回
func (s *Service) GetRealRoomID(ctx context.Context, roomID int64) (int64, error) {
	path := fmt.Sprintf(api.EndpointRoomInfo, roomID)
	var resp realRoomInfoResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return 0, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return 0, err
	}
	if resp.Data.RoomID != 0 {
		return resp.Data.RoomID, nil
	}
	return roomID, nil
}

// GetRealRoomInfo 获取直播间详细信息
//
// 调用 B站 /room/v1/Room/get_info 接口
// 无需 Cookie 登录态
//
// 参数 roomID 为直播间真实房间号（长 ID）如果是短 ID，可先用 GetRealRoomID 解析
func (s *Service) GetRealRoomInfo(ctx context.Context, roomID int64) (*RealRoomInfo, error) {
	path := fmt.Sprintf(api.EndpointRoomInfo, roomID)
	var resp realRoomInfoResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// =========================================================================
// 弹幕 WebSocket 连接信息（WBI 签名）
// =========================================================================

// DanmuInfo 弹幕 WebSocket 连接所需信息
type DanmuInfo struct {
	Token   string `json:"token"`    // WebSocket 认证 Token
	Host    string `json:"host"`     // WebSocket 服务器主机名（优先使用 WSS）
	Port    int    `json:"port"`     // TCP 端口
	WSPort  int    `json:"ws_port"`  // WebSocket（非加密）端口
	WSSPort int    `json:"wss_port"` // WebSocket Secure（SSL 加密）端口
}

type danmuInfoHost struct {
	Host    string `json:"host"`     // 服务器主机名
	Port    int    `json:"port"`     // TCP 端口
	WSPort  int    `json:"ws_port"`  // WS 端口
	WSSPort int    `json:"wss_port"` // WSS 端口
}

type danmuInfoResponse struct {
	api.Response
	Data struct {
		Token    string          `json:"token"`     // WebSocket 连接 Token
		HostList []danmuInfoHost `json:"host_list"` // 可用服务器列表，取第一个即可
	} `json:"data"`
}

// GetDanmuInfo 获取弹幕 WebSocket 连接信息
//
// 调用 B站 /xlive/web-room/v1/index/getDanmuInfo 接口
// 需要 WBI 签名
//
// 参数：
//   - roomID: 直播间真实房间号
//   - imgKey, subKey: WBI 签名密钥，通过 auth.GetWbiKeys 获取
//
// 返回的 DanmuInfo 包含 WebSocket 连接地址和 Token，可用于连接 B站 弹幕服务器
func (s *Service) GetDanmuInfo(ctx context.Context, roomID int64, imgKey, subKey string) (*DanmuInfo, error) {
	params := map[string]string{
		"id":           strconv.FormatInt(roomID, 10),
		"type":         "0",
		"web_location": "444.8",
	}
	signedQuery := wbi.Sign(params, imgKey, subKey)

	path := fmt.Sprintf(api.EndpointDanmuInfo, signedQuery)
	var resp danmuInfoResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	if len(resp.Data.HostList) == 0 {
		return nil, fmt.Errorf("room: no available websocket host")
	}
	h := resp.Data.HostList[0]
	return &DanmuInfo{
		Token:   resp.Data.Token,
		Host:    h.Host,
		Port:    h.Port,
		WSPort:  h.WSPort,
		WSSPort: h.WSSPort,
	}, nil
}

// =========================================================================
// 弹幕发送权限与发送
// =========================================================================

// BarragePermission 用户在目标直播间的弹幕发送权限
type BarragePermission struct {
	Mode   int `json:"mode"`   // 弹幕模式（滚动弹幕等）
	Color  int `json:"color"`  // 弹幕颜色（十进制 RGB 值，如 16777215）
	Length int `json:"length"` // 弹幕最大长度（UTF-8 字符数）
	Bubble int `json:"bubble"` // 气泡类型（0 为无气泡）
}

type barragePermissionResponse struct {
	api.Response
	Data struct {
		Property struct {
			Danmu  BarragePermission `json:"danmu"`  // 弹幕权限
			Bubble int               `json:"bubble"` // 气泡权限
		} `json:"property"`
	} `json:"data"`
}

// GetBarragePermission 获取用户在目标直播间的弹幕发送权限
//
// 调用 B站 /xlive/web-room/v1/index/getInfoByUser 接口
// 需要 Cookie 登录态
//
// 返回的权限信息用于 SendDanmu 时设置弹幕颜色、模式等参数
func (s *Service) GetBarragePermission(ctx context.Context, roomID int64) (*BarragePermission, error) {
	path := fmt.Sprintf(api.EndpointBarragePermission, roomID)
	var resp barragePermissionResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data.Property.Danmu, nil
}

// SendDanmu 发送弹幕到指定直播间
//
// 调用 B站 /msg/send 接口（form-urlencoded POST）
// 需要 Cookie 登录态和 CSRF Token
//
// 参数：
//   - roomID: 直播间真实房间号
//   - message: 弹幕内容（UTF-8 字符串，最长 30 个字符）
//   - csrfToken: CSRF Token，通过 client.CSRF() 从 CookieJar 中获取
//
// 内部会先调用 GetBarragePermission 获取弹幕权限，再发送弹幕
func (s *Service) SendDanmu(ctx context.Context, roomID int64, message string, csrfToken string) error {
	perm, err := s.GetBarragePermission(ctx, roomID)
	if err != nil {
		return fmt.Errorf("room: get barrage permission: %w", err)
	}
	form := url.Values{
		"color":      {strconv.Itoa(perm.Color)},
		"fontsize":   {"25"},
		"mode":       {strconv.Itoa(perm.Mode)},
		"msg":        {message},
		"rnd":        {strconv.FormatInt(time.Now().Unix(), 10)},
		"roomid":     {strconv.FormatInt(roomID, 10)},
		"bubble":     {strconv.Itoa(perm.Bubble)},
		"csrf_token": {csrfToken},
		"csrf":       {csrfToken},
	}

	var resp api.Response
	if err := s.client.PostForm(ctx, api.EndpointSendDanmu, form, &resp); err != nil {
		return fmt.Errorf("room: send danmu: %w", err)
	}
	return api.CheckError(resp.Code, resp.Message)
}

// =========================================================================
// 在线排行榜
// =========================================================================

// OnlineRankItem 在线排行条目
type OnlineRankItem struct {
	Rank  int    `json:"userRank"` // 排名
	UID   int64  `json:"uid"`      // 用户 UID
	Name  string `json:"name"`     // 用户名称
	Score int64  `json:"score"`    // 贡献值（金瓜子数）
}

// OnlineGoldRank 在线金瓜子榜
type OnlineGoldRank struct {
	OnlineNum int              `json:"onlineNum"`      // 在线高能用户数
	Items     []OnlineRankItem `json:"OnlineRankItem"` // 排行列表
}

type onlineGoldRankResponse struct {
	api.Response
	Data OnlineGoldRank `json:"data"`
}

// GetOnlineGoldRank 获取直播间在线金瓜子榜
//
// 调用 B站 /xlive/general-interface/v1/rank/getOnlineGoldRank 接口
// 无需 Cookie 登录态
//
// 参数：
//   - anchorUID: 主播 UID
//   - roomID: 直播间真实房间号
func (s *Service) GetOnlineGoldRank(ctx context.Context, anchorUID, roomID int64) (*OnlineGoldRank, error) {
	path := fmt.Sprintf(
		api.EndpointOnlineGoldRank,
		anchorUID, roomID,
	)
	var resp onlineGoldRankResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// =========================================================================
// 大航海（舰长数）
// =========================================================================

type vipNumbersResponse struct {
	api.Response
	Data struct {
		Info struct {
			Num int `json:"num"` // 大航海总人数（舰长+提督+总督）
		} `json:"info"`
	} `json:"data"`
}

// GetVipNumbers 获取直播间大航海（舰长/提督/总督）总人数
//
// 调用 B站 /xlive/app-room/v2/guardTab/topListNew 接口
// 无需 Cookie 登录态
//
// 参数：
//   - anchorUID: 主播 UID
//   - roomID: 直播间真实房间号
func (s *Service) GetVipNumbers(ctx context.Context, anchorUID, roomID int64) (int, error) {
	path := fmt.Sprintf(
		api.EndpointGuardTopList,
		anchorUID, roomID,
	)
	var resp vipNumbersResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return -1, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return -1, err
	}
	return resp.Data.Info.Num, nil
}

// =========================================================================
// 禁言管理
// =========================================================================

// AddSilentUser 在直播间禁言指定用户
//
// 调用 B站 /xlive/web-ucenter/v1/banned/AddSilentUser 接口（form-urlencoded POST）
// 需要 Cookie 登录态和 CSRF Token操作用户需为房管或主播
//
// 参数：
//   - roomID: 直播间真实房间号
//   - targetUID: 要禁言的目标用户 UID
//   - msg: 违规弹幕内容（可为空字符串）
//   - csrfToken: CSRF Token，通过 client.CSRF() 获取
func (s *Service) AddSilentUser(ctx context.Context, roomID, targetUID int64, msg, csrfToken string) error {
	form := url.Values{
		"room_id":    {strconv.FormatInt(roomID, 10)},
		"tuid":       {strconv.FormatInt(targetUID, 10)},
		"msg":        {msg},
		"mobile_app": {"web"},
		"csrf_token": {csrfToken},
		"csrf":       {csrfToken},
	}
	var resp api.Response
	if err := s.client.PostForm(ctx, api.EndpointAddSilentUser, form, &resp); err != nil {
		return err
	}
	return api.CheckError(resp.Code, resp.Message)
}

// SilentUserItem 禁言用户条目
type SilentUserItem struct {
	ID     int64  `json:"id"`      // 禁言记录 ID（用于解除禁言）
	RoomID int64  `json:"room_id"` // 直播间房间号
	UID    int64  `json:"uid"`     // 被禁言用户 UID
	Name   string `json:"tname"`   // 被禁言用户名称
}

// SilentListResponse 禁言用户列表响应
type SilentListResponse struct {
	Total     int              `json:"total"`      // 禁言总人数
	TotalPage int              `json:"total_page"` // 总页数
	Items     []SilentUserItem `json:"data"`       // 当前页禁言记录列表
}

type silentListResponse struct {
	api.Response
	Data SilentListResponse `json:"data"`
}

// GetSilentUserList 获取直播间禁言用户列表
//
// 调用 B站 /xlive/web-ucenter/v1/banned/GetSilentUserList 接口（form-urlencoded POST）
// 需要 Cookie 登录态和 CSRF Token操作用户需为房管或主播
//
// 参数：
//   - roomID: 直播间真实房间号
//   - page: 页码，从 1 开始
//   - csrfToken: CSRF Token，通过 client.CSRF() 获取
//
// 返回的 SilentListResponse 包含：
//   - Total: 总禁言用户数
//   - TotalPage: 总页数
//   - Items: 当前页禁言记录
func (s *Service) GetSilentUserList(ctx context.Context, roomID, page int64, csrfToken string) (*SilentListResponse, error) {
	form := url.Values{
		"room_id":    {strconv.FormatInt(roomID, 10)},
		"ps":         {strconv.FormatInt(page, 10)},
		"csrf_token": {csrfToken},
		"csrf":       {csrfToken},
	}
	var resp silentListResponse
	if err := s.client.PostForm(ctx, api.EndpointGetSilentUserList, form, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// DelSilentUser 解除直播间禁言
//
// 调用 B站 /banned_service/v1/Silent/del_room_block_user 接口（form-urlencoded POST）
// 需要 Cookie 登录态和 CSRF Token操作用户需为房管或主播
//
// 参数：
//   - roomID: 直播间真实房间号
//   - blacklistID: 禁言记录 ID（来自 SilentUserItem.ID，非用户 UID）
//   - csrfToken: CSRF Token，通过 client.CSRF() 获取
func (s *Service) DelSilentUser(ctx context.Context, roomID, blacklistID int64, csrfToken string) error {
	form := url.Values{
		"roomid":     {strconv.FormatInt(roomID, 10)},
		"id":         {strconv.FormatInt(blacklistID, 10)},
		"csrf_token": {csrfToken},
		"csrf":       {csrfToken},
	}
	var resp api.Response
	if err := s.client.PostForm(ctx, api.EndpointDelSilentUser, form, &resp); err != nil {
		return err
	}
	return api.CheckError(resp.Code, resp.Message)
}
