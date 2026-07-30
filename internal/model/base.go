package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt int64          `gorm:"not null;comment:创建时间"`
	UpdatedAt int64          `gorm:"not null;comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除"`
}

// BeforeCreate 自动写入时间
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().Unix()
	if m.CreatedAt == 0 {
		m.CreatedAt = now
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate 自动更新更新时间
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now().Unix()
	return nil
}
