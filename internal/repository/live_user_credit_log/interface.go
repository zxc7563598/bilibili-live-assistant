package live_user_credit_log

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// AddCreditLogParams 添加积分/星光变动日志的参数
type AddCreditLogParams struct {
	UserID       int64             // 用户ID
	ChangeType   enum.ChangeType   // 变动类型（增加/减少）
	ChangeAmount int64             // 变动数值
	BizType      string            // 业务类型
	Remark       string            // 备注
	OperatorType enum.OperatorType // 操作方
	OperatorID   int64             // 操作人标识ID
}

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserCreditLog]

	// AddStarsLog 添加星光变动记录，并同步更新用户星光
	AddStarsLog(ctx context.Context, tx *gorm.DB, params AddCreditLogParams) (*model.LiveUserCreditLog, error)
	// AddPointsLog 添加积分变动记录，并同步更新用户积分
	AddPointsLog(ctx context.Context, tx *gorm.DB, params AddCreditLogParams) (*model.LiveUserCreditLog, error)
}

// AddStarsLog 添加星光变动记录
func (r *gormRepo) AddStarsLog(ctx context.Context, tx *gorm.DB, params AddCreditLogParams) (*model.LiveUserCreditLog, error) {
	return r.addCreditLog(ctx, tx, enum.CreditTypeStars, params)
}

// AddPointsLog 添加积分变动记录
func (r *gormRepo) AddPointsLog(ctx context.Context, tx *gorm.DB, params AddCreditLogParams) (*model.LiveUserCreditLog, error) {
	return r.addCreditLog(ctx, tx, enum.CreditTypePoints, params)
}

// addCreditLog 核心方法：事务内查询用户 → 计算变动前/后数值 → 创建日志 → 更新用户积分/星光
func (r *gormRepo) addCreditLog(ctx context.Context, tx *gorm.DB, creditType enum.CreditType, params AddCreditLogParams) (*model.LiveUserCreditLog, error) {
	db := r.getDB(ctx, tx)
	var log model.LiveUserCreditLog
	err := db.Transaction(func(tx *gorm.DB) error {
		// 查询当前用户
		var user model.LiveUser
		if err := tx.Where("uid = ?", params.UserID).First(&user).Error; err != nil {
			return fmt.Errorf("查询用户失败: %w", err)
		}
		// 获取变动前数值
		var beforeValue int64
		var fieldName string
		if creditType == enum.CreditTypeStars {
			beforeValue = user.Stars
			fieldName = "stars"
		} else {
			beforeValue = user.Points
			fieldName = "points"
		}
		// 计算变动后数值
		var afterValue int64
		switch params.ChangeType {
		case enum.ChangeTypeIncrease:
			afterValue = beforeValue + params.ChangeAmount
		case enum.ChangeTypeReduce:
			if beforeValue < params.ChangeAmount {
				return fmt.Errorf("余额不足: 当前值为 %d，无法减少 %d", beforeValue, params.ChangeAmount)
			}
			afterValue = beforeValue - params.ChangeAmount
		default:
			return fmt.Errorf("未知的变动类型: %v", params.ChangeType)
		}
		// 创建日志记录
		log = model.LiveUserCreditLog{
			UserID:       params.UserID,
			CreditType:   creditType,
			ChangeType:   params.ChangeType,
			ChangeAmount: params.ChangeAmount,
			BeforeValue:  beforeValue,
			AfterValue:   afterValue,
			BizType:      params.BizType,
			Remark:       params.Remark,
			OperatorType: params.OperatorType,
			OperatorID:   params.OperatorID,
		}
		if err := tx.Create(&log).Error; err != nil {
			return fmt.Errorf("创建日志失败: %w", err)
		}
		// 更新用户积分/星光
		if err := tx.Model(&user).Where("uid = ?", params.UserID).Update(fieldName, afterValue).Error; err != nil {
			return fmt.Errorf("更新用户%s失败: %w", fieldName, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &log, nil
}
