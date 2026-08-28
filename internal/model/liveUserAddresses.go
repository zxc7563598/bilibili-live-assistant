package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveUserAddress struct {
	ID         int64            `gorm:"primaryKey"`
	UserID     int64            `gorm:"not null;comment:用户ID"`
	Name       string           `gorm:"type:varchar(100);comment:名称"`
	Phone      string           `gorm:"type:varchar(100);comment:手机号"`
	RegionCode string           `gorm:"type:varchar(100);comment:地区code,['370000', '370100', '370116']"`
	Region     string           `gorm:"type:varchar(255);comment:地区文字描述，空格隔开"`
	Detail     string           `gorm:"type:varchar(255);comment:详细地址"`
	Email      string           `gorm:"type:varchar(255);comment:邮箱地址"`
	Type       enum.AddressType `gorm:"type:smallint;not null;default:0;comment:地址类型"`
	IsDefault  enum.YesNo       `gorm:"type:smallint;not null;default:0;comment:默认地址"`
	BaseModel
}

func (LiveUserAddress) TableName() string {
	return "live_user_addresses"
}
