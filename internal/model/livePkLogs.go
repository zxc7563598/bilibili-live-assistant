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
