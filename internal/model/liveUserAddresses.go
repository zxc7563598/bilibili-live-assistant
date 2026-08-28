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

// 数据库暂时需要实现的方法：
// 根据用户ID获取全部数据
// 根据ID变更Default（用户只能有一个 Default === enum.Yes 的数据，每次变更时需要先吧用户所有的数据置换为 No 然后再变更）
// 暂时没想到什么其他的定的，根据这个表的内容帮我做一些常用的查询/变更方法吧
