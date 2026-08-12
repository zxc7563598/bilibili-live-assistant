package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type Scene int

const (
	SceneUnlimited Scene = iota
	SceneLive
	SceneNotLive
)

func (s Scene) Key() string {
	switch s {
	case SceneUnlimited:
		return "unlimited"
	case SceneLive:
		return "scene.live"
	case SceneNotLive:
		return "scene.not_live"
	default:
		return "unknown"
	}
}

func (s Scene) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s Scene) IsValid() bool {
	switch s {
	case SceneUnlimited, SceneLive, SceneNotLive:
		return true
	default:
		return false
	}
}
