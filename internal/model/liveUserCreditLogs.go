package model

import (
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"gorm.io/gorm"
)

type LiveUserCreditLog struct {
	ID           int64             `gorm:"primaryKey"`
	UserID       int64             `gorm:"not null;comment:用户ID"`
	CreditType   enum.CreditType   `gorm:"not null;default:0,comment:积分类型"`
	ChangeType   enum.ChangeType   `gorm:"not null;default:0,comment:变动类型"`
	ChangeAmount int64             `gorm:"not null;default:0;comment:变动数值"`
	BeforeValue  int64             `gorm:"not null;default:0;comment:变动前数值"`
	AfterValue   int64             `gorm:"not null;default:0;comment:变动后数值"`
	BizType      string            `gorm:"type:varchar(100);comment:业务类型"`
	Remark       string            `gorm:"type:varchar(255);comment:备注/原因说明"`
	OperatorType enum.OperatorType `gorm:"not null;default:0,comment:操作方"`
	OperatorID   int64             `gorm:"not null;default:0,comment:操作人标识ID"`
	BaseModel
}

func (LiveUserCreditLog) TableName() string {
	return "live_user_credit_logs"
}

// BeforeUpdate 禁止修改日志记录
func (l *LiveUserCreditLog) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("live_user_credit_logs 不允许修改")
}

// BeforeDelete 禁止删除日志记录
func (l *LiveUserCreditLog) BeforeDelete(tx *gorm.DB) error {
	return errors.New("live_user_credit_logs 不允许删除")
}
