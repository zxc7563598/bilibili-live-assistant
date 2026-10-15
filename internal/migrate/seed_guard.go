package migrate

import (
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"gorm.io/gorm"
)

// seedLiveGuardRecords 从历史礼物记录回填大航海开通台账
//
// 仅在台账为空时执行一次：表里已经有数据说明要么此前回填过、要么是导入进来的，
// 此时再重放只会把同一批开通重复记一遍。
//
// 回填出来的名单月份区间是「理想堆叠」——按每次开通的月份往后接 num 个月。
// 每日对账任务会把实际过期时间提前清掉的部分，这里看不见，而且它改的是
// live_users 的到期时间（天数口径），与名单月份无关。各等级独立记账，互不截断。
func seedLiveGuardRecords(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.LiveGuardRecord{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var gifts []model.LiveGift
	if err := db.Where("gift_type = ?", enum.GiftTypeGuard).
		Order("uid asc, send_at asc, id asc").
		Find(&gifts).Error; err != nil {
		return err
	}

	type guardKey struct {
		uid   int64
		level enum.BadgeType
	}
	// 每个 (用户, 档位) 各自维护上一段覆盖到的月份，作为下一次顺延的基准
	prevCoverTo := make(map[guardKey]int64)
	rows := make([]model.LiveGuardRecord, 0, len(gifts))
	for _, gift := range gifts {
		level, ok := enum.BadgeTypeByGuardName(gift.GiftName)
		if !ok {
			continue
		}
		key := guardKey{uid: gift.UID, level: level}
		coverFrom, coverTo := liveuser.GuardCoverWindow(prevCoverTo[key], time.Unix(gift.SendAt, 0), gift.Num)
		prevCoverTo[key] = coverTo
		rows = append(rows, model.LiveGuardRecord{
			RoomID:     gift.RoomID,
			UID:        gift.UID,
			Uname:      gift.Uname,
			GuardLevel: level,
			GiftID:     gift.GiftID,
			Price:      gift.Price,
			Num:        gift.Num,
			OpenAt:     gift.SendAt,
			CoverFrom:  coverFrom,
			CoverTo:    coverTo,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return db.CreateInBatches(&rows, 500).Error
}
