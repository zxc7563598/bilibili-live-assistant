package migrate

import (
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed 填充数据
func Seed(db *gorm.DB) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedMenus(db); err != nil {
		return err
	}
	if err := seedAdmin(db); err != nil {
		return err
	}
	if err := seedAdminRole(db); err != nil {
		return err
	}
	if err := seedRobotConfigs(db); err != nil {
		return err
	}
	if err := seedAppConfigs(db); err != nil {
		return err
	}
	// 种子数据带固定 ID，PostgreSQL 需要把自增序列推进到其后（其余数据库为空操作）
	return syncPostgresSequences(db)
}

// seedRoles 初始化填充角色表
// 以 code 为唯一键做幂等 upsert：已存在的跳过，未来新增的角色会自动追加。
func seedRoles(db *gorm.DB) error {
	role := model.Role{
		ID:     1,
		Code:   "SUPER_ADMIN",
		Name:   "超级管理员",
		Enable: enum.EnableEnable,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&role).Error
}

// seedMenus 初始化填充菜单表
// 以 code 为唯一键做幂等 upsert：已存在的菜单跳过，未来新增的菜单会自动追加。
func seedMenus(db *gorm.DB) error {
	menus := []model.Menu{
		{
			ID:        1,
			Code:      "SysMgt",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "系统管理",
			Icon:      "i-fe:grid",
			Path:      "",
			Component: "",
			Order:     98,
		},
		{
			ID:        2,
			Code:      "MenuMgt",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  1,
			Name:      "菜单管理",
			Icon:      "i-fe:list",
			Path:      "/pms/resource",
			Component: "/src/views/pms/resource/index.vue",
			Order:     1,
		},
		{
			ID:        3,
			Code:      "RoleMgt",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  1,
			Name:      "角色管理",
			Icon:      "i-fe:user-check",
			Path:      "/pms/role",
			Component: "/src/views/pms/role/index.vue",
			Order:     2,
		},
		{
			ID:        4,
			Code:      "UserMgt",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.Yes,
			Layout:    "",
			Type:      "MENU",
			ParentID:  1,
			Name:      "用户管理",
			Icon:      "i-fe:user",
			Path:      "/pms/user",
			Component: "/src/views/pms/user/index.vue",
			Order:     3,
		},
		{
			ID:        5,
			Code:      "RoleUser",
			Enable:    enum.EnableEnable,
			Show:      enum.No,
			KeepAlive: enum.No,
			Layout:    "full",
			Type:      "MENU",
			ParentID:  3,
			Name:      "分配用户",
			Icon:      "i-fe:user-plus",
			Path:      "/pms/role/user/:roleId",
			Component: "/src/views/pms/role/role-user.vue",
			Order:     1,
		},
		{
			ID:        6,
			Code:      "AddRole",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "BUTTON",
			ParentID:  3,
			Name:      "新增角色",
			Icon:      "",
			Path:      "",
			Component: "",
			Order:     0,
		},
		{
			ID:        7,
			Code:      "AddUser",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "BUTTON",
			ParentID:  4,
			Name:      "添加用户",
			Icon:      "i-fe:grid",
			Path:      "",
			Component: "",
			Order:     0,
		},
		{
			ID:        8,
			Code:      "UserProfile",
			Enable:    enum.EnableEnable,
			Show:      enum.No,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "个人资料",
			Icon:      "i-fe:user",
			Path:      "/profile",
			Component: "/src/views/profile/index.vue",
			Order:     99,
		},
		{
			ID:        9,
			Code:      "Home",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "欢迎您",
			Icon:      "i-fe:disc",
			Path:      "/",
			Component: "/src/views/home/index.vue",
			Order:     0,
		},
		{
			ID:        10,
			Code:      "Robot",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "机器人配置",
			Icon:      "i-fe:gitlab",
			Path:      "/robot",
			Component: "/src/views/robotconfig/index.vue",
			Order:     1,
		},
		{
			ID:        11,
			Code:      "Analyze",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "数据分析",
			Icon:      "i-fe:pie-chart",
			Path:      "",
			Component: "",
			Order:     2,
		},
		{
			ID:        12,
			Code:      "LiveDanmu",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  11,
			Name:      "弹幕信息",
			Icon:      "i-fe:message-square",
			Path:      "/livedanmu/list",
			Component: "/src/views/analyze/livedanmu/index.vue",
			Order:     0,
		},
		{
			ID:        13,
			Code:      "GiftList",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  11,
			Name:      "礼物信息",
			Icon:      "i-fe:gift",
			Path:      "/giftlist/list",
			Component: "/src/views/analyze/giftlist/index.vue",
			Order:     1,
		},
		{
			ID:        14,
			Code:      "BlindBoxList",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  11,
			Name:      "盲盒信息",
			Icon:      "i-fe:box",
			Path:      "/blindbox/list",
			Component: "/src/views/analyze/blindbox/index.vue",
			Order:     2,
		},
		{
			ID:        15,
			Code:      "UserAnalysis",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  11,
			Name:      "用户分析",
			Icon:      "i-fe:activity",
			Path:      "/user/analysis",
			Component: "/src/views/analyze/user/index.vue",
			Order:     3,
		},
		{
			ID:        16,
			Code:      "AppConfig",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "App 配置",
			Icon:      "i-fe:smartphone",
			Path:      "/app",
			Component: "/src/views/appconfig/index.vue",
			Order:     1,
		},
		{
			ID:        17,
			Code:      "Shop",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "商城管理",
			Icon:      "i-fe:shopping-bag",
			Path:      "",
			Component: "",
			Order:     3,
		},
		{
			ID:        18,
			Code:      "ShopProduct",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  17,
			Name:      "商品管理",
			Icon:      "i-fe:shopping-cart",
			Path:      "/shop/product",
			Component: "/src/views/shop/product/index.vue",
			Order:     1,
		},
		{
			ID:        19,
			Code:      "ShopUser",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  17,
			Name:      "用户管理",
			Icon:      "i-fe:users",
			Path:      "/shop/user",
			Component: "/src/views/shop/user/index.vue",
			Order:     0,
		},
		{
			ID:        20,
			Code:      "Order",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "订单管理",
			Icon:      "i-fe:trello",
			Path:      "",
			Component: "",
			Order:     4,
		},
		{
			ID:        21,
			Code:      "OrderList",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  20,
			Name:      "订单列表",
			Icon:      "i-fe:list",
			Path:      "/order/list",
			Component: "/src/views/order/list/index.vue",
			Order:     0,
		},
		{
			ID:        22,
			Code:      "OrderDelivery",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  20,
			Name:      "发货管理",
			Icon:      "i-fe:package",
			Path:      "/order/delivery",
			Component: "/src/views/order/delivery/index.vue",
			Order:     1,
		},
		{
			ID:        23,
			Code:      "ShopProductDetails",
			Enable:    enum.EnableEnable,
			Show:      enum.No,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  18,
			Name:      "商品详情",
			Icon:      "i-fe:edit",
			Path:      "/shop/product/details",
			Component: "/src/views/shop/product/details.vue",
			Order:     1,
		},
		{
			ID:        24,
			Code:      "ShopUserDetails",
			Enable:    enum.EnableEnable,
			Show:      enum.No,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  19,
			Name:      "用户详情",
			Icon:      "i-fe:edit",
			Path:      "/shop/user/details",
			Component: "/src/views/shop/user/details.vue",
			Order:     1,
		},
		{
			ID:        25,
			Code:      "OrderDetails",
			Enable:    enum.EnableEnable,
			Show:      enum.No,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  21,
			Name:      "订单详情",
			Icon:      "i-fe:edit",
			Path:      "/order/list/details",
			Component: "/src/views/order/list/details.vue",
			Order:     2,
		},
		{
			ID:        26,
			Code:      "AddProduct",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "BUTTON",
			ParentID:  18,
			Name:      "添加商品",
			Icon:      "i-fe:edit",
			Path:      "",
			Component: "",
			Order:     0,
		},
		{
			ID:        27,
			Code:      "PkBattle",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  11,
			Name:      "PK对战",
			Icon:      "i-fe:shuffle",
			Path:      "/pk/list",
			Component: "/src/views/analyze/pk/index.vue",
			Order:     4,
		},
		{
			ID:        28,
			Code:      "Complaint",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  17,
			Name:      "投诉管理",
			Icon:      "i-fe:file-text",
			Path:      "/shop/complaint",
			Component: "/src/views/shop/complaint/index.vue",
			Order:     2,
		},
		{
			ID:        29,
			Code:      "RoomConfig",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "房间配置",
			Icon:      "i-fe:home",
			Path:      "/room",
			Component: "/src/views/room/index.vue",
			Order:     1,
		},
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoNothing: true,
	}).Create(&menus).Error
}

// seedAdmin 初始化填充管理员表
func seedAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}
	password, err := crypto.HashPassword("123456")
	if err != nil {
		return fmt.Errorf("初始账号密码生成错误: %w", err)
	}
	role := model.Admin{
		ID:       1,
		Nickname: "默认管理员",
		Username: "admin",
		Password: password,
		RoleID:   1,
		Gender:   enum.GenderUnknown,
		Enable:   enum.EnableEnable,
	}
	return db.Create(&role).Error
}

// seedAdminRole 初始化填充管理员角色表
func seedAdminRole(db *gorm.DB) error {
	var count int64
	db.Model(&model.AdminRole{}).Count(&count)
	if count > 0 {
		return nil
	}
	role := model.AdminRole{
		ID:      1,
		AdminID: 1,
		RoleID:  1,
	}
	return db.Create(&role).Error
}

// seedRobotConfigs 初始化机器人配置表
//
// 所有功能默认关闭（enabled=false），由用户在管理后台按需开启和配置。
// 配置值使用 text 类型存储，复杂配置（如答谢模板、回复规则）使用 JSON 格式。
//
// 以 (group_name, config_key) 为唯一键做幂等 upsert：
// 已存在的配置项跳过（不覆盖用户在后台改过的值），未来新增的配置项会自动追加。
func seedRobotConfigs(db *gorm.DB) error {
	configs := []model.RobotConfig{
		{
			GroupName:   "room",
			ConfigKey:   "room_id",
			ConfigValue: "0",
			Remark:      "监听的直播间房间号, 0 表示未设置",
		},
		{
			GroupName:   "room",
			ConfigKey:   "is_listening",
			ConfigValue: "0",
			Remark:      "是否默认监听直播间, 0-否, 1-是",
		},
		{
			GroupName:   "room",
			ConfigKey:   "max_name_length",
			ConfigValue: "8",
			Remark:      "用户名最大长度, 超过此长度则裁剪",
		},
		{
			GroupName:   "room",
			ConfigKey:   "name_trim_mode",
			ConfigValue: "0",
			Remark:      "裁剪方式, 0-省略后面, 1-省略前面",
		},
		{
			GroupName:   "room",
			ConfigKey:   "consume_reward_enabled",
			ConfigValue: "0",
			Remark:      "用户消费发放奖励, 0-不发放, 1-按消费电池发放, 2-按开通航海类型发放",
		},
		{
			GroupName:   "room",
			ConfigKey:   "reward_type",
			ConfigValue: "1",
			Remark:      "奖励类型, 0-星光, 1-积分",
		},
		{
			GroupName:   "room",
			ConfigKey:   "consume_battery_rate",
			ConfigValue: "0",
			Remark:      "消费电池转换倍率, 设置为 2 则代表用户消耗 1 电池会得到 2 点奖励, 奖励设置为按消费电池发放时生效",
		},
		{
			GroupName:   "room",
			ConfigKey:   "captain_reward_amount",
			ConfigValue: "0",
			Remark:      "开通舰长奖励数量, 奖励设置为按开通航海类型发放时生效",
		},
		{
			GroupName:   "room",
			ConfigKey:   "commander_reward_amount",
			ConfigValue: "0",
			Remark:      "开通提督奖励数量, 奖励设置为按开通航海类型发放时生效",
		},
		{
			GroupName:   "room",
			ConfigKey:   "governor_reward_amount",
			ConfigValue: "0",
			Remark:      "开通总督奖励数量, 奖励设置为按开通航海类型发放时生效",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "reward_type",
			ConfigValue: "0",
			Remark:      "奖励类型, 0-星光, 1-积分",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "reward_amount",
			ConfigValue: "10",
			Remark:      "奖励数量, 正整数",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "keyword",
			ConfigValue: "#签到",
			Remark:      "签到关键词, 用户触发签到的词, 建议增加符号以避免错误触发",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "query_keyword",
			ConfigValue: "#查询",
			Remark:      "查询关键词, 用户触发查询的词, 建议增加符号以避免错误触发",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "success_reply",
			ConfigValue: "",
			Remark:      "签到成功回复, 支持占位符变量",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "fail_reply",
			ConfigValue: "",
			Remark:      "签到失败回复, 一般为不在可用场景或不符合触发门槛， 支持占位符变量",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "repeat_reply",
			ConfigValue: "",
			Remark:      "重复签到回复, 支持占位符变量",
		},
		{
			GroupName:   "sign",
			ConfigKey:   "query_reply",
			ConfigValue: "",
			Remark:      "查询成功回复, 支持占位符变量",
		},
		{
			GroupName:   "ad",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "ad",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "ad",
			ConfigKey:   "interval",
			ConfigValue: "62",
			Remark:      "发送间隔, 秒",
		},
		{
			GroupName:   "ad",
			ConfigKey:   "send_mode",
			ConfigValue: "0",
			Remark:      "发送方式, 0-随机发送, 1-顺序发送",
		},
		{
			GroupName:   "ad",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "发送内容, 支持占位符变量",
		},

		{
			GroupName:   "pk",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "pk",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "发送内容, 支持占位符变量",
		},

		{
			GroupName:   "gift",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "答谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "show_count",
			ConfigValue: "1",
			Remark:      "展示数量, 在答谢礼物时标注数量, 0-禁用, 1-启用",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "merge_gift",
			ConfigValue: "1",
			Remark:      "礼物合并, 一次性感谢用户在短时间内赠送的多个礼物, 0-禁用, 1-启用",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "include_blindbox",
			ConfigValue: "1",
			Remark:      "盲盒统计, 盲盒礼物在感谢末尾携带盈亏信息, 0-禁用, 1-启用",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "min_battery",
			ConfigValue: "10",
			Remark:      "起始感谢电池, 低于此电池数的礼物不触发感谢",
		},
		{
			GroupName:   "gift",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "感谢内容, 支持占位符变量",
		},

		{
			GroupName:   "welcome",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "welcome",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "welcome",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "欢迎门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "welcome",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "欢迎内容, 支持占位符变量",
		},
		{
			GroupName:   "follow",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "follow",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "follow",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "follow",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "感谢内容, 支持占位符变量",
		},
		{
			GroupName:   "share",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "share",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "share",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "share",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "感谢内容, 支持占位符变量",
		},
		{
			GroupName:   "reply",
			ConfigKey:   "enabled",
			ConfigValue: "0",
			Remark:      "是否启用, 0-禁用, 1-启用",
		},
		{
			GroupName:   "reply",
			ConfigKey:   "scene",
			ConfigValue: "0",
			Remark:      "可用场景, 0-不限制, 1-直播中, 2-非直播中",
		},
		{
			GroupName:   "reply",
			ConfigKey:   "requirement",
			ConfigValue: "0",
			Remark:      "触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子",
		},
		{
			GroupName:   "reply",
			ConfigKey:   "content",
			ConfigValue: "",
			Remark:      "回复内容, json 配置项",
		},
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_name"}, {Name: "config_key"}},
		DoNothing: true,
	}).Create(&configs).Error
}

// agreementDefaultContent 用户协议默认正文，Markdown 格式
const agreementDefaultContent = `**特别提醒：请用户仔细阅读本协议所有条款（包括附件），尤其是加粗字体及加下划线部分，请务必仔细阅读。如您对本协议内容有任何疑问，那就别有疑问。一旦您通过勾选/点击或确认本协议，即意味着您已阅读本协议所有条款，并对本协议条款的含义及相应的法律后果已全部通晓并充分理解，您同意以数据电文形式订立本协议，虽然本协议完全不具有任何法律约束力。为重视未成年人权益的保障，您在使用本服务时应具备完全民事行为能力。若您不具备完全民事行为能力，请立即v我50。**

## 基本协议

1. 坚持主播的绝对领导。直播间里主播永远是第一位
2. 爱护主播，做文明观众，做到“打不还手，骂不还口，笑脸迎送冷屁股”
3. 诚心接受主播感情上的独裁，“不要和陌生人说话”，尤其不能跟陌生主播说话。当然，问路的老太太除外。
4. 坚持工资奖金全部上缴制度。不涂改工资条，不在衣柜里藏钱。不过，每月可以申请领取500元零花（日元）
5. 用户有义务从心里热爱主播崇敬主播，视主播为自己的上帝，与主播沟通时须使用低三下四/讨好/献媚等温柔语气，不得用生硬、顶撞的语气
6. 主播拥有精神羞辱权，有权剥夺用户的一切自由和尊严，用语言和行为强迫甲方，达到羞辱的目的。`

// seedAppConfigs 初始化 APP 配置表
//
// 以 config_key 为唯一键做幂等 upsert：
// 已存在的配置项跳过（不覆盖用户在后台改过的值），未来新增的配置项会自动追加。
func seedAppConfigs(db *gorm.DB) error {
	configs := []model.AppConfig{
		{
			ConfigKey:   "site_name",
			ConfigValue: "积分商城",
			Remark:      "显示在浏览器标签栏、收藏夹以及手机桌面图标下方的名称",
		},
		{
			ConfigKey:   "site_description",
			ConfigValue: "这是xxxxx的积分商城",
			Remark:      "PWA 应用描述（手机浏览器提示「添加到主屏幕」时显示的说明文案）",
		},
		{
			ConfigKey:   "site_background_color",
			ConfigValue: "#f5f6f8",
			Remark:      "PWA 启动页背景色（应用打开瞬间到首页渲染完成前显示的背景颜色）",
		},
		{
			ConfigKey:   "site_theme_color",
			ConfigValue: "#965bff",
			Remark:      "网站主题色（应用于按钮、边框、图标、选中状态等主要 UI 元素的颜色）",
		},
		{
			ConfigKey:   "site_icon",
			ConfigValue: "https://cdn.hejunjie.life/avatars/shop.png",
			Remark:      "浏览器标签页图标，同时也是手机添加到桌面时的应用图标",
		},
		{
			ConfigKey:   "register",
			ConfigValue: "1",
			Remark:      "是否允许用户自助注册。0-禁止注册，1-允许注册（默认开启）",
		},
		{
			ConfigKey:   "logo",
			ConfigValue: "https://cdn.hejunjie.life/avatars/shop.png",
			Remark:      "网站 Logo 图片地址。建议与浏览器 favicon 图标保持一致，以便在标签页和收藏夹中统一显示",
		},
		{
			ConfigKey:   "login_bg",
			ConfigValue: "",
			Remark:      "登录页背景图片地址。留空则自动根据当前主题主色生成纯色背景",
		},
		{
			ConfigKey:   "login_title",
			ConfigValue: "积分商城",
			Remark:      "登录页顶部显示的主标题，用于品牌或产品名称展示",
		},
		{
			ConfigKey:   "login_slogan",
			ConfigValue: "纯美女神伊德利拉美貌盖世无双！",
			Remark:      "登录页副标题或宣传语（Slogan），可填写品牌口号、活动标语等内容",
		},
		{
			ConfigKey:   "agreement_title",
			ConfigValue: "关于进一步加强主播领导地位的若干规定",
			Remark:      "登录页用户协议标题。留空则登录页不展示协议入口",
		},
		{
			ConfigKey:   "agreement_content",
			ConfigValue: agreementDefaultContent,
			Remark:      "登录页用户协议正文，支持 Markdown 语法",
		},
		{
			ConfigKey:   "oss_endpoint",
			ConfigValue: "https://oss-cn-hangzhou.aliyuncs.com",
			Remark:      "完整 OSS 地址",
		},
		{
			ConfigKey:   "oss_access_key_id",
			ConfigValue: "",
			Remark:      "阿里云 AccessKey ID",
		},
		{
			ConfigKey:   "oss_access_key_secret",
			ConfigValue: "",
			Remark:      "阿里云 AccessKey Secret",
		},
		{
			ConfigKey:   "oss_bucket",
			ConfigValue: "",
			Remark:      "Bucket 名",
		},
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoNothing: true,
	}).Create(&configs).Error
}
