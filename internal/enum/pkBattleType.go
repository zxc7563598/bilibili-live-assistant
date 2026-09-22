package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type PkBattleType int

// PK 对战类型，取值来自 B 站协议（非连续，所以不能直接用 iota）
const (
	PkBattleTypeClassic PkBattleType = 2
	PkBattleTypeRoyale  PkBattleType = 6
)

func (p PkBattleType) Key() string {
	switch p {
	case PkBattleTypeClassic:
		return "pk_battle_type.classic_pk"
	case PkBattleTypeRoyale:
		return "pk_battle_type.battle_royale"
	default:
		return "unknown"
	}
}

func (p PkBattleType) Text(lang string) string {
	return i18n.T(lang, p.Key())
}

func (p PkBattleType) IsValid() bool {
	switch p {
	case PkBattleTypeClassic, PkBattleTypeRoyale:
		return true
	default:
		return false
	}
}
