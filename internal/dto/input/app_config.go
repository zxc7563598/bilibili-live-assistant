package input

// AppConfigSaveReq 后台整体保存 App 配置请求
type AppConfigSaveReq struct {
	// 站点名称（浏览器标签栏 / PWA 安装后名称）
	SiteName string `json:"site_name" example:"积分商城"`
	// 站点说明（PWA 应用描述）
	SiteDescription string `json:"site_description" example:"这是xxxxx的积分商城"`
	// PWA 启动页背景色
	SiteBackgroundColor string `json:"site_background_color" example:"#f5f6f8"`
	// 网站主题色
	SiteThemeColor string `json:"site_theme_color" example:"#965bff"`
	// 网站图标路径（标签页 / 桌面应用图标）
	SiteIcon string `json:"site_icon" example:""`
	// 是否允许用户自助注册, 0-禁止, 1-允许
	Register string `json:"register" binding:"required,oneof=0 1" err:"required=10001,oneof=10001" example:"1"`
	// 网站 Logo 路径（登录页等场景展示）
	Logo string `json:"logo" example:""`
	// 登录页背景图路径（留空则根据主题色生成背景）
	LoginBg string `json:"login_bg" example:""`
	// 登录页主标题
	LoginTitle string `json:"login_title" example:"积分商城"`
	// 登录页副标题 / Slogan
	LoginSlogan string `json:"login_slogan" example:""`
}
