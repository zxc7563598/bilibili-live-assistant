package appconfig

// Manifest 请求返回
type ManifestResp struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	BackgroundColor string `json:"background_color"`
	Icon            string `json:"icon"`
	IconType        string `json:"icon_type"`
}
