package input

// AppConfigSaveConfigReq 站点配置保存请求入参
type AppConfigSaveConfigReq struct {
	// 站点名称（浏览器标签栏 / PWA 安装后名称）
	SiteName string `json:"site_name" binding:"max=100" err:"max=10902" example:"积分商城"`
	// 站点说明（PWA 应用描述）
	SiteDescription string `json:"site_description" binding:"max=255" err:"max=10902" example:"这是xxxxx的积分商城"`
	// PWA 启动页背景色
	SiteBackgroundColor string `json:"site_background_color" binding:"max=32" err:"max=10902" example:"#f5f6f8"`
	// 网站主题色
	SiteThemeColor string `json:"site_theme_color" binding:"max=32" err:"max=10902" example:"#965bff"`
	// 网站图标路径（标签页 / 桌面应用图标）
	SiteIcon string `json:"site_icon" binding:"max=255" err:"max=10902" example:""`
	// 是否允许用户自助注册, 0-禁止, 1-允许
	Register string `json:"register" binding:"required,oneof=0 1" err:"required=10902,oneof=10902" example:"1"`
	// 网站 Logo 路径（登录页等场景展示）
	Logo string `json:"logo" binding:"max=255" err:"max=10902" example:""`
	// 登录页背景图路径（留空则根据主题色生成背景）
	LoginBg string `json:"login_bg" binding:"max=255" err:"max=10902" example:""`
	// 登录页主标题
	LoginTitle string `json:"login_title" binding:"max=100" err:"max=10902" example:"积分商城"`
	// 登录页副标题 / Slogan
	LoginSlogan string `json:"login_slogan" binding:"max=255" err:"max=10902" example:""`
	// 登录页用户协议标题
	AgreementTitle string `json:"agreement_title" binding:"max=100" err:"max=10902" example:"关于进一步加强主播领导地位的若干规定"`
	// 登录页用户协议正文（Markdown 格式）
	AgreementContent string `json:"agreement_content" binding:"max=20000" err:"max=10902" example:"## 基本协议"`
}

// AppConfigSaveOssConfigReq OSS 配置保存请求入参
type AppConfigSaveOssConfigReq struct {
	// 完整 OSS 地址
	OssEndpoint string `json:"oss_endpoint" binding:"max=255" err:"max=10902" example:"https://oss-cn-hangzhou.aliyuncs.com"`
	// 阿里云 AccessKey ID
	OssAccessKeyId string `json:"oss_access_key_id" binding:"max=255" err:"max=10902" example:"xxxxxxxxxxxxx"`
	// 阿里云 AccessKey Secret
	OssAccessKeySecret string `json:"oss_access_key_secret" binding:"max=255" err:"max=10902" example:"xxxxxxxxxxxxx"`
	// 目标 bucket 名
	OssBucket string `json:"oss_bucket" binding:"max=255" err:"max=10902" example:"xxxxx"`
}
