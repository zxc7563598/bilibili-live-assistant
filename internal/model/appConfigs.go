package model

// APP配置表
type AppConfig struct {
	ID          int64  `gorm:"primaryKey"`
	ConfigKey   string `gorm:"type:varchar(100);not null;uniqueIndex:uk_config_key;comment:配置键"`
	ConfigValue string `gorm:"type:text;not null;comment:配置值"`
	Remark      string `gorm:"type:varchar(255);not null;default:'';comment:备注说明"`
	BaseModel
}

func (AppConfig) TableName() string {
	return "app_configs"
}
