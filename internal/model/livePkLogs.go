package model

// LivePkLog PK 对战记录
type LivePkLog struct {
	ID          int64  `gorm:"primaryKey"`
	RoomID      int64  `gorm:"not null;default:0;index;comment:本直播间真实房间号"`
	PkID        int64  `gorm:"not null;uniqueIndex;comment:PK ID"`
	PkStatus    int64  `gorm:"not null;default:0;comment:PK 状态"`
	BattleType  int64  `gorm:"not null;default:0;comment:对战类型"`
	MatchType   int64  `gorm:"not null;default:0;comment:匹配类型"`
	RivalUID    int64  `gorm:"not null;default:0;comment:对方UID"`
	RivalUname  string `gorm:"type:varchar(100);not null;default:'';comment:对方用户名"`
	RivalRoomID int64  `gorm:"not null;default:0;comment:对方房间ID"`
	SelfVotes   int64  `gorm:"not null;default:0;comment:我方最终PK值"`
	RivalVotes  int64  `gorm:"not null;default:0;comment:对方最终PK值"`
	SelfResult  int64  `gorm:"not null;default:0;comment:我方胜负 2=获胜 -1=落败"`
	RivalResult int64  `gorm:"not null;default:0;comment:对方胜负 2=获胜 -1=落败"`
	StartAt     int64  `gorm:"not null;default:0;comment:PK开始时间（秒级时间戳）"`
	SettleAt    int64  `gorm:"not null;default:0;comment:PK结算时间（秒级时间戳）"`
	BaseModel
}

func (LivePkLog) TableName() string {
	return "live_pk_logs"
}

// LivePkLogListPageQuery PK 记录分页查询入参，不对应数据库表
type LivePkLogListPageQuery struct {
	// RoomID 本直播间真实房间号，精确匹配
	RoomID *int64
	// RivalUID 对方 UID，精确匹配
	RivalUID *int64
	// RivalUname 对方用户名，模糊匹配
	RivalUname *string
	// SelfResult 我方胜负，与库中 self_result 同口径：-1=落败 2=获胜
	SelfResult   *int
	StartAtStart *int64
	StartAtEnd   *int64
	Offset       int
	Limit        int
	SortField    *string
	SortOrder    *string
}
