package migrate

import (
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"gorm.io/gorm"
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
	return nil
}

// seedRoles 初始化填充角色表
func seedRoles(db *gorm.DB) error {
	var count int64
	db.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return nil
	}
	role := model.Role{
		ID:     1,
		Code:   "SUPER_ADMIN",
		Name:   "超级管理员",
		Enable: enum.EnableEnable,
	}
	return db.Create(&role).Error
}

// seedMenus 初始化填充菜单表
func seedMenus(db *gorm.DB) error {
	var count int64
	db.Model(&model.Menu{}).Count(&count)
	if count > 0 {
		return nil
	}
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
			Name:      "资产大盘",
			Icon:      "i-fe:home",
			Path:      "/",
			Component: "/src/views/home/index.vue",
			Order:     0,
		},
		{
			ID:        10,
			Code:      "iFrame",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "外部链接",
			Icon:      "i-fe:insert-link",
			Path:      "",
			Component: "",
			Order:     0,
		},
		{
			ID:        11,
			Code:      "Blog",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  10,
			Name:      "个人博客",
			Icon:      "i-fe:trello",
			Path:      "https://hejunjie.life",
			Component: "",
			Order:     0,
		},
		{
			ID:        12,
			Code:      "NaiveUI",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  10,
			Name:      "Naive UI",
			Icon:      "i-me:naiveui",
			Path:      "https://www.naiveui.com/zh-CN/os-theme",
			Component: "",
			Order:     1,
		},
		{
			ID:        13,
			Code:      "Base",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  0,
			Name:      "基础功能",
			Icon:      "i-fe:grid",
			Path:      "",
			Component: "",
			Order:     1,
		},
		{
			ID:        14,
			Code:      "Icon",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  13,
			Name:      "图标 Icon",
			Icon:      "i-fe:feather",
			Path:      "/base/icon",
			Component: "/src/views/base/unocss-icon.vue",
			Order:     0,
		},
		{
			ID:        15,
			Code:      "BaseComponents",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  13,
			Name:      "基础组件",
			Icon:      "i-me:awesome",
			Path:      "/base/components",
			Component: "/src/views/base/index.vue",
			Order:     1,
		},
		{
			ID:        16,
			Code:      "Unocss",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  13,
			Name:      "基础组件",
			Icon:      "i-me:awesome",
			Path:      "/base/unocss",
			Component: "/src/views/base/unocss.vue",
			Order:     2,
		},
		{
			ID:        17,
			Code:      "KeepAlive",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.Yes,
			Layout:    "",
			Type:      "MENU",
			ParentID:  13,
			Name:      "KeepAlive",
			Icon:      "i-me:awesome",
			Path:      "/base/keep-alive",
			Component: "/src/views/base/keep-alive.vue",
			Order:     3,
		},
		{
			ID:        18,
			Code:      "MeModal",
			Enable:    enum.EnableEnable,
			Show:      enum.Yes,
			KeepAlive: enum.No,
			Layout:    "",
			Type:      "MENU",
			ParentID:  13,
			Name:      "MeModal",
			Icon:      "i-me:dialog",
			Path:      "/testModal",
			Component: "/src/views/base/test-modal.vue",
			Order:     4,
		},
	}
	return db.Create(&menus).Error
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
