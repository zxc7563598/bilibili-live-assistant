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

// ConfigData 请求返回
type ConfigDataResp struct {
	SiteName            string `json:"site_name"`
	SiteDescription     string `json:"site_description"`
	SiteBackgroundColor string `json:"site_background_color"`
	SiteThemeColor      string `json:"site_theme_color"`
	SiteIcon            string `json:"site_icon"`
	Register            string `json:"register"`
	Logo                string `json:"logo"`
	LoginBg             string `json:"login_bg"`
	LoginTitle          string `json:"login_title"`
	LoginSlogan         string `json:"login_slogan"`
	OssEndpoint         string `json:"oss_endpoint"`
	OssAccessKeyId      string `json:"oss_access_key_id"`
	OssAccessKeySecret  string `json:"oss_access_key_secret"`
	OssBucket           string `json:"oss_bucket"`
}

// SaveConfig 请求入参
type SaveConfigReq struct {
	SiteName            string `json:"site_name"`
	SiteDescription     string `json:"site_description"`
	SiteBackgroundColor string `json:"site_background_color"`
	SiteThemeColor      string `json:"site_theme_color"`
	SiteIcon            string `json:"site_icon"`
	Register            string `json:"register"`
	Logo                string `json:"logo"`
	LoginBg             string `json:"login_bg"`
	LoginTitle          string `json:"login_title"`
	LoginSlogan         string `json:"login_slogan"`
}

// SaveOssConfig 请求入参
type SaveOssConfigReq struct {
	OssEndpoint        string `json:"oss_endpoint"`
	OssAccessKeyId     string `json:"oss_access_key_id"`
	OssAccessKeySecret string `json:"oss_access_key_secret"`
	OssBucket          string `json:"oss_bucket"`
}
