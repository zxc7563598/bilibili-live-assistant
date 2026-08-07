package model

type RobotConfig struct {
	ID          int64  `gorm:"primaryKey"`
	GroupName   string `gorm:"type:varchar(100);not null;uniqueIndex:uk_group_key;comment:分组名称"`
	ConfigKey   string `gorm:"type:varchar(100);not null;uniqueIndex:uk_group_key;comment:配置键"`
	ConfigValue string `gorm:"type:text;not null;comment:配置值"`
	Remark      string `gorm:"type:varchar(255);not null;default:'';comment:备注说明"`
	BaseModel
}

func (RobotConfig) TableName() string {
	return "robot_configs"
}
