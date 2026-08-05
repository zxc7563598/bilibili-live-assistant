package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type Admin struct {
	ID       int64      `gorm:"primaryKey"`
	Nickname string      `gorm:"type:varchar(100);not null;comment:昵称"`
	Username string      `gorm:"type:varchar(100);not null;uniqueIndex;comment:用户名"`
	Password string      `gorm:"type:varchar(255);not null;comment:密码"`
	Token    *string     `gorm:"type:varchar(255);comment:登录凭证"`
	RoleID   int64      `gorm:"comment:当前角色ID"`
	Avatar   string      `gorm:"type:varchar(255);not null;default:'https://cdn.hejunjie.life/avatars/oneadmin.jpeg';comment:头像"`
	Email    *string     `gorm:"type:varchar(255);comment:邮箱"`
	Address  *string     `gorm:"type:varchar(255);comment:地址"`
	Gender   enum.Gender `gorm:"type:smallint;not null;default:2;comment:性别"`
	Enable   enum.Enable `gorm:"type:smallint;not null;default:1;comment:是否启用"`
	BaseModel
}

func (Admin) TableName() string {
	return "admins"
}

// AdminListPageQuery 用于后台列表查询入参，不对应数据库表
type AdminListPageQuery struct {
	Username *string
	Gender   *int
	Enable   *int
	Offset   int
	Limit    int
}

// AdminListItem 用于后台列表展示，不对应数据库表
type AdminListItem struct {
	ID        int64
	Username  string
	Enable    enum.Enable
	Gender    enum.Gender
	Avatar    string
	Address   *string
	Email     *string
	RoleID    *int64
	CreatedAt int64
	UpdatedAt int64
}

// AdminUpdateProfileByIdForm 用于更新管理员个人资料，不对应数据库表
type AdminUpdateProfileByIdForm struct {
	Nickname *string
	Email    *string
	Address  *string
	Gender   *int
	Avatar   *string
}

// AdminUpdateBasicInfoByIdForm 用于更新管理员账号信息，不对应数据库表
type AdminUpdateBasicInfoByIdForm struct {
	Username *string
	Enable   *enum.Enable
	RoleID   *int64
}
