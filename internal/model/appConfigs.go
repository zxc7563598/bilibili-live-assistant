package model

// APP配置表
type AppConfig struct {
	ID          int64  `gorm:"primaryKey"`
	ConfigKey   string `gorm:"type:varchar(100);not null;uniqueIndex:uk_group_key;comment:配置键"`
	ConfigValue string `gorm:"type:text;not null;comment:配置值"`
	Remark      string `gorm:"type:varchar(255);not null;default:'';comment:备注说明"`
	BaseModel
}

func (AppConfig) TableName() string {
	return "app_configs"
}

// 数据库暂时需要实现的方法：
// - 获取所有配置（这张表的内容是不会随意增加的，最多也就几十条，用于决定APP配置，后续是考虑启动的时候查一次塞到缓存里的）
