// Package user 提供 B站 用户/主播信息相关的 API 调用
// 本包不导入父包 bilibili，通过自有的 HttpClient 接口实现依赖反转
package user

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/api"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/wbi"
)

// Service 提供 B站 用户信息和主播信息相关的 API 调用
type Service struct {
	client HttpClient
}

// NewService 创建用户信息服务实例
// client 需实现 HttpClient 接口（Get/Post），由父包 *bilibili.Client 自动满足
func NewService(client HttpClient) *Service {
	return &Service{client: client}
}

// =========================================================================
// 主播基本信息（无需登录态）
// =========================================================================

// MasterInfo 主播基本信息（无需 Cookie 登录态）
type MasterInfo struct {
	UID  int64  `json:"uid"`   // 主播 UID
	Name string `json:"uname"` // 主播名称
	Face string `json:"face"`  // 主播头像 URL
}

type masterInfoResponse struct {
	api.Response
	Data struct {
		Info MasterInfo `json:"info"`
	} `json:"data"`
}

// GetMasterInfo 获取指定 UID 的主播基本信息
//
// 调用 B站 /live_user/v1/Master/info 接口
// 无需 Cookie 登录态
//
// 参数 uid 为主播的 UID（非房间号）
func (s *Service) GetMasterInfo(ctx context.Context, uid int64) (*MasterInfo, error) {
	path := fmt.Sprintf(api.EndpointMasterInfo, uid)
	var resp masterInfoResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data.Info, nil
}

// =========================================================================
// 用户空间信息（WBI 签名 + 需要登录态）
// =========================================================================

// StreamerInfo 用户空间详细信息
type StreamerInfo struct {
	MID  int64  `json:"mid"`  // 用户 MID（同 UID）
	Name string `json:"name"` // 用户昵称
	Sex  string `json:"sex"`  // 性别：男/女/保密
	Face string `json:"face"` // 头像 URL
	Sign string `json:"sign"` // 个人签名
}

type streamerInfoResponse struct {
	api.Response
	Data StreamerInfo `json:"data"`
}

// GetStreamerInfo 获取指定 UID 的用户空间信息
//
// 调用 B站 /x/space/wbi/acc/info 接口
// 需要 Cookie 登录态，且需要 WBI 签名
//
// 参数：
//   - uid: 目标用户 UID
//   - imgKey, subKey: WBI 签名密钥，通过 auth.Service.GetWbiKeys 获取
func (s *Service) GetStreamerInfo(ctx context.Context, uid int64, imgKey, subKey string) (*StreamerInfo, error) {
	params := map[string]string{"mid": fmt.Sprintf("%d", uid)}
	signedQuery := wbi.Sign(params, imgKey, subKey)

	path := fmt.Sprintf(api.EndpointStreamerInfo, signedQuery)
	var resp streamerInfoResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if err := api.CheckError(resp.Code, resp.Message); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
