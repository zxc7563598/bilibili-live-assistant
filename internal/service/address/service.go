package address

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_address"
	"gorm.io/gorm"
)

type Service struct {
	db                  *gorm.DB
	liveUserAddressRepo live_user_address.Repository
}

func New(db *gorm.DB, liveUserAddressRepo live_user_address.Repository) *Service {
	return &Service{
		db:                  db,
		liveUserAddressRepo: liveUserAddressRepo,
	}
}

// GetDefaultAddress 获取用户指定类型的默认收货地址
func (s *Service) GetDefaultAddress(ctx context.Context, userID int64, addressType int) (AddressItem, int, error) {
	t := enum.AddressType(addressType)
	if !t.IsValid() {
		return AddressItem{}, 11303, nil
	}
	entity, err := s.liveUserAddressRepo.GetDefaultByUserID(ctx, nil, userID, t)
	if err != nil {
		return AddressItem{}, 61301, err
	}
	if entity == nil {
		return AddressItem{}, 0, nil
	}
	return toAddressItem(*entity), 0, nil
}

// GetAddressList 获取用户收货地址列表；addressType 为 nil 时返回全部类型，否则仅返回该类型
func (s *Service) GetAddressList(ctx context.Context, userID int64, addressType *int) ([]AddressItem, int, error) {
	var t *enum.AddressType
	if addressType != nil {
		v := enum.AddressType(*addressType)
		if !v.IsValid() {
			return []AddressItem{}, 11303, nil
		}
		t = &v
	}
	list, err := s.liveUserAddressRepo.ListByUserID(ctx, nil, userID, t)
	if err != nil {
		return []AddressItem{}, 61301, err
	}
	items := make([]AddressItem, 0, len(list))
	for _, entity := range list {
		items = append(items, toAddressItem(entity))
	}
	return items, 0, nil
}

// GetAddressByID 根据收货地址ID获取收货地址信息
func (s *Service) GetAddressByID(ctx context.Context, userID, id int64) (AddressItem, int, error) {
	entity, err := s.liveUserAddressRepo.GetByID(ctx, nil, id)
	if err != nil {
		return AddressItem{}, 61301, err
	}
	if entity == nil || entity.UserID != userID {
		return AddressItem{}, 51301, nil
	}
	return toAddressItem(*entity), 0, nil
}

// SaveAddress 添加/变更收货地址信息，返回地址 ID
func (s *Service) SaveAddress(ctx context.Context, userID int64, req AddressReq) (int64, int, error) {
	isCreate := req.ID == nil || *req.ID <= 0
	// type / is_default 未提供时：创建走默认（虚拟 / 非默认），更新则保留原值
	var t enum.AddressType
	if req.Type == nil {
		if isCreate {
			t = enum.AddressTypeVirtual
		}
	} else if v := enum.AddressType(*req.Type); v.IsValid() {
		t = v
	} else {
		return 0, 11303, nil
	}
	var isDefault enum.YesNo
	if req.IsDefault == nil {
		if isCreate {
			isDefault = enum.No
		}
	} else if v := enum.YesNo(*req.IsDefault); v.IsValid() {
		isDefault = v
	} else {
		return 0, 11301, nil
	}
	// 事务外完成校验与实体组装
	var entity model.LiveUserAddress
	if isCreate {
		e, code, err := s.buildCreateEntity(userID, req, t, isDefault)
		if code != 0 {
			return 0, code, err
		}
		entity = *e
	} else {
		e, code, err := s.buildUpdateEntity(ctx, userID, req)
		if code != 0 {
			return 0, code, err
		}
		entity = *e
	}
	// 写入 + 置默认
	var addressID int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if isCreate {
			created, err := s.liveUserAddressRepo.Create(ctx, tx, &entity)
			if err != nil {
				return err
			}
			addressID = created.ID
		} else {
			if err := s.liveUserAddressRepo.Update(ctx, tx, &entity); err != nil {
				return err
			}
			addressID = entity.ID
		}
		if isDefault == enum.Yes {
			if err := s.liveUserAddressRepo.SetDefault(ctx, tx, addressID, userID, t); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, 61301, err
	}
	return addressID, 0, nil
}

// DeleteAddress 删除用户的收货地址（软删除）；仅允许删除归属当前用户且存在的地址
func (s *Service) DeleteAddress(ctx context.Context, userID, id int64) (int, error) {
	entity, err := s.liveUserAddressRepo.GetByID(ctx, nil, id)
	if err != nil {
		return 61301, err
	}
	if entity == nil || entity.UserID != userID {
		return 51301, nil
	}
	if err := s.liveUserAddressRepo.Delete(ctx, nil, id); err != nil {
		return 61301, err
	}
	return 0, nil
}
