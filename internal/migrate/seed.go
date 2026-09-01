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
	return nil
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
			Name:      "房间配置",
			Icon:      "i-fe:home",
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
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoNothing: true,
	}).Create(&configs).Error
}
