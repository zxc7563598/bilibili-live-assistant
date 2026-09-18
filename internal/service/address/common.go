package address

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/region"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// toAddressItem 将数据模型转换为对外返回的地址项
func toAddressItem(m model.LiveUserAddress) AddressItem {
	return AddressItem{
		ID:         m.ID,
		Name:       m.Name,
		Phone:      m.Phone,
		RegionCode: m.RegionCode,
		Region:     m.Region,
		Detail:     m.Detail,
		Email:      m.Email,
		Type:       int(m.Type),
		IsDefault:  int(m.IsDefault),
	}
}

// buildCreateEntity 校验新增地址的必填项并组装实体（事务外调用）
func (s *Service) buildCreateEntity(userID int64, req AddressReq, t enum.AddressType, isDefault enum.YesNo) (*model.LiveUserAddress, int, error) {
	entity := &model.LiveUserAddress{
		UserID:    userID,
		Name:      ptr.TrimStr(req.Name),
		Phone:     ptr.TrimStr(req.Phone),
		Detail:    ptr.TrimStr(req.Detail),
		Email:     ptr.TrimStr(req.Email),
		Type:      t,
		IsDefault: isDefault,
	}
	// 地区：实体地址必填并由 code 派生文案；虚拟地址提供 code 时按非必填处理（用于清空）
	required := t == enum.AddressTypeActual
	if code, err := s.applyRegion(entity, req, required); code != 0 {
		return nil, code, err
	}
	// 按类型统一校验必填项
	if code := validateEntityFields(entity); code != 0 {
		return nil, code, nil
	}
	return entity, 0, nil
}

// buildUpdateEntity 校验归属并把请求中的字段合并到原实体（事务外调用）。
// type / is_default 与其余字段一致采用「仅覆盖请求中提供的字段」语义，
// 合并完成后按与创建路径一致的规则校验该类型必填项，防止部分更新绕过服务端校验。
func (s *Service) buildUpdateEntity(ctx context.Context, userID int64, req AddressReq) (*model.LiveUserAddress, int, error) {
	entity, err := s.liveUserAddressRepo.GetByID(ctx, nil, *req.ID)
	if err != nil {
		return nil, CodeQueryFailed, err
	}
	if entity == nil || entity.UserID != userID {
		return nil, CodeNotFound, nil
	}

	// 仅覆盖请求中提供的字段，未提供字段保留原值
	if req.Name != nil {
		entity.Name = ptr.TrimStr(req.Name)
	}
	if req.Phone != nil {
		entity.Phone = ptr.TrimStr(req.Phone)
	}
	if req.Detail != nil {
		entity.Detail = ptr.TrimStr(req.Detail)
	}
	if req.Email != nil {
		entity.Email = ptr.TrimStr(req.Email)
	}
	if req.Type != nil {
		entity.Type = enum.AddressType(*req.Type)
	}
	if req.IsDefault != nil {
		entity.IsDefault = enum.YesNo(*req.IsDefault)
	}

	// 地区字段：
	//   - 提供 region_code → 校验/清空并派生文案（以后端派生为准）；实体地址下地区必填
	//   - 仅提供 region 文案 → 直接覆盖文案
	//   - 两者都未提供 → 地区保持不变
	if req.RegionCode != nil {
		required := entity.Type == enum.AddressTypeActual
		if code, err := s.applyRegion(entity, req, required); code != 0 {
			return nil, code, err
		}
	} else if req.Region != nil {
		entity.Region = ptr.TrimStr(req.Region)
	}
	// 合并后按该类型校验必填项（含 region_code 已清空的实体地址）
	if code := validateEntityFields(entity); code != 0 {
		return nil, code, nil
	}
	return entity, 0, nil
}

// validateEntityFields 按地址类型校验实体必填项（创建与更新共用）
func validateEntityFields(e *model.LiveUserAddress) int {
	if e.Name == "" {
		return CodeNameRequired
	}
	switch e.Type {
	case enum.AddressTypeActual:
		// 实体地址：手机号 + 地区 + 详细地址
		if e.Phone == "" {
			return CodePhoneRequired
		}
		if e.RegionCode == "" {
			return CodeRegionRequired
		}
		if e.Detail == "" {
			return CodeDetailRequired
		}
	default:
		// 虚拟地址：仅需邮箱
		if e.Email == "" {
			return CodeEmailRequired
		}
	}
	return 0
}

// applyRegion 根据请求中的 region_code 设置/清空地区信息。
// RegionCode 语义为 JSON 数组字符串（如 ["370000","370100","370116"]）：
//   - nil：required 时视为地区缺失；否则地区字段保持不动
//   - 空串或 "[]"：视为清空地区（region_code 与 region 文案均置空）；required 时视为缺失
//   - 合法省市区链：归一化为紧凑 JSON 存储，并用 label 派生 region 文案
//
// region 文案始终以后端从 regions.json 派生的结果为准，避免与前端拼接不一致。
func (s *Service) applyRegion(entity *model.LiveUserAddress, req AddressReq, required bool) (int, error) {
	if req.RegionCode == nil {
		if required {
			return CodeRegionRequired, nil
		}
		return 0, nil
	}
	src := strings.TrimSpace(*req.RegionCode)
	if src == "" {
		if required {
			return CodeRegionRequired, nil
		}
		entity.RegionCode = ""
		entity.Region = ""
		return 0, nil
	}

	var codes []string
	if err := json.Unmarshal([]byte(src), &codes); err != nil {
		return CodeRegionInvalid, err
	}
	if len(codes) == 0 {
		if required {
			return CodeRegionRequired, nil
		}
		entity.RegionCode = ""
		entity.Region = ""
		return 0, nil
	}

	text, ok := region.Resolve(codes)
	if !ok {
		return CodeRegionInvalid, nil
	}
	canonical, err := json.Marshal(codes)
	if err != nil {
		return CodeRegionInvalid, err
	}
	entity.RegionCode = string(canonical)
	entity.Region = text
	return 0, nil
}
