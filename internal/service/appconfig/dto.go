package appconfig

// Manifest 请求返回
type ManifestResp struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	BackgroundColor string `json:"background_color"`
	Icon            string `json:"icon"`
	IconType        string `json:"icon_type"`
}

// LoginConfig 请求返回
type LoginConfig struct {
	Register bool   `json:"register"`
	Logo     string `json:"logo"`
	LoginBg  string `json:"login_bg"`
	Title    string `json:"title"`
	Slogan   string `json:"slogan"`
}
