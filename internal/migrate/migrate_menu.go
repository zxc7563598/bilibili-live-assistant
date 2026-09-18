package migrate

import (
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// legacyMenuFix 一条「老库里配错了、需要就地改正」的菜单字段记录
type legacyMenuFix struct {
	// Code 菜单编码，定位用。menus.code 上有唯一索引，正常只会命中一行
	Code string
	// OldNames 历史种子数据给这条菜单用过的名称。**只有当前名称命中其中之一才会修正**，
	// 因此管理员自己在后台改过的名称不在名单里，不会被覆盖
	OldNames []string
	// NewName / NewIcon 要修正成的目标值，与 seed.go 里该菜单的当前定义保持一致
	NewName string
	NewIcon string
}

// legacyMenuFixes 需要修正的菜单清单
var legacyMenuFixes = []legacyMenuFix{
	{
		Code:     "Home",
		OldNames: []string{"房间配置"},
		NewName:  "欢迎您",
		NewIcon:  "i-fe:disc",
	},
}

// fixLegacyMenus 兼容历史数据：把老库里被改过含义的菜单名称与图标修正为当前种子数据的值
//
// 依赖 Run() 中先执行的 AutoMigrate 建好 menus 表（表不存在时直接返回）。
//
// 幂等且不误伤：
//   - 只匹配 OldNames 里列出的历史名称。管理员在后台把菜单改成别的名字后不再命中，
//     自定义不会被覆盖；房间配置已独立成 RoomConfig 菜单，不会和 Home 混淆。
//   - 命中的行被改成目标值后就不再匹配，之后每次启动都是空操作。
//
// 走 GORM 的 Updates 而不是原生 SQL，是为了让 BaseModel 的软删除过滤生效
// （已删除的行不该被改）。注意 BaseModel.UpdatedAt 没有 autoUpdateTime tag，
// 钩子在 map 更新时写不进 SET 子句，因此 updated_at 要显式带上。
func fixLegacyMenus(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.Menu{}) {
		return nil
	}
	now := time.Now().Unix()
	for _, fix := range legacyMenuFixes {
		res := db.Model(&model.Menu{}).
			Where("code = ? AND name IN ?", fix.Code, fix.OldNames).
			Updates(map[string]any{
				"name":       fix.NewName,
				"icon":       fix.NewIcon,
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			log.Printf("[migrate] menus 修正 %s 菜单的名称/图标完成，共 %d 条", fix.Code, res.RowsAffected)
		}
	}
	return nil
}
