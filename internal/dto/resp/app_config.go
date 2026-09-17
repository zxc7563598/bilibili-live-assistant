package resp

// AppShopManifestResp 商城 PWA / 站点配置返回
type AppShopManifestResp struct {
	// 应用名称（全称，用于页面标题与安装名称）
	Name string `json:"name" example:"积分商城"`
	// 应用短名称（空间不足时展示）
	ShortName string `json:"short_name" example:"商城"`
	// 应用描述
	Description string `json:"description" example:"积分商城的描述"`
	// 主题色（浏览器地址栏 / 窗口标题栏颜色）
	ThemeColor string `json:"theme_color" example:"#ffffff"`
	// 启动屏背景色
	BackgroundColor string `json:"background_color" example:"#ffffff"`
	// 浏览器标签页小图标
	Favicon string `json:"favicon" example:"https://cdn.example.com/favicon.svg"`
	// iOS 添加到主屏的图标（180x180 PNG）
	AppleTouchIcon string `json:"apple_touch_icon" example:"https://cdn.example.com/icon-180.png"`
	// 启动地址
	StartURL string `json:"start_url" example:"/shop/"`
	// 作用域
	Scope string `json:"scope" example:"/shop/"`
	// 显示模式
	Display string `json:"display" example:"standalone"`
	// 安装图标列表
	Icons []AppShopManifestIcon `json:"icons"`
}

// AppShopManifestIcon 商城 PWA 安装图标
type AppShopManifestIcon struct {
	// 图标地址
	Src string `json:"src" example:"https://cdn.example.com/icon-192.png"`
	// 尺寸
	Sizes string `json:"sizes" example:"192x192"`
	// 图片类型
	Type string `json:"type" example:"image/png"`
	// 用途（any / maskable）
	Purpose string `json:"purpose" example:"any"`
}

// AppPublicKeyResp 商城前端加密所需的 RSA 公钥响应
type AppPublicKeyResp struct {
	// 公钥标识（公钥内容 sha256 前 16 位 hex，用于前端验签与密钥轮换识别）
	KeyID string `json:"key_id" example:"3f2ab8d0e1c4a9b7"`
	// RSA 公钥（SPKI DER 的 base64 编码，前端 atob 后 importKey("spki") 使用）
	PublicKey string `json:"public_key" example:"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A..."`
	// 签名生成时间戳（Unix 秒，前端校验时间窗口）
	Timestamp int64 `json:"timestamp" example:"1724716800"`
	// HMAC-SHA256 签名（对 "pubkey:"+key_id+public_key+timestamp 计算，hex 编码）
	Sign string `json:"sign" example:"a1b2c3d4e5f60718293a4b5c6d7e8f90..."`
}

// AppShopThemeColorResp 商城 PWA / 站点配置返回
type AppShopThemeColorResp struct {
	Color string `json:"color" example:"#ffffff"`
}

// AppShopLoginConfigResp 获取登录页面配置信息返回
type AppShopLoginConfigResp struct {
	Register bool   `json:"register" example:"false"`
	Logo     string `json:"logo" example:"https://cdn.hejunjie.life/avatars/shop.png"`
	LoginBg  string `json:"login_bg" example:""`
	Title    string `json:"title" example:"积分商城"`
	Slogan   string `json:"slogan" example:"登录后可兑换积分好礼"`
}

// AppConfigDataResp 后台查询 App 全部配置返回
type AppConfigDataResp struct {
	// 站点名称（浏览器标签栏 / PWA 安装后名称）
	SiteName string `json:"site_name" example:"积分商城"`
	// 站点说明（PWA 应用描述）
	SiteDescription string `json:"site_description" example:"这是xxxxx的积分商城"`
	// PWA 启动页背景色
	SiteBackgroundColor string `json:"site_background_color" example:"#f5f6f8"`
	// 网站主题色
	SiteThemeColor string `json:"site_theme_color" example:"#965bff"`
	// 网站图标路径（标签页 / 桌面应用图标）
	SiteIcon string `json:"site_icon" example:"https://cdn.hejunjie.life/avatars/shop.png"`
	// 是否允许用户自助注册, 0-禁止, 1-允许
	Register string `json:"register" example:"1"`
	// 网站 Logo 路径（登录页等场景展示）
	Logo string `json:"logo" example:"https://cdn.hejunjie.life/avatars/shop.png"`
	// 登录页背景图路径（留空则根据主题色生成背景）
	LoginBg string `json:"login_bg" example:""`
	// 登录页主标题
	LoginTitle string `json:"login_title" example:"积分商城"`
	// 登录页副标题 / Slogan
	LoginSlogan string `json:"login_slogan" example:""`
	// 完整 OSS 地址
	OssEndpoint string `json:"oss_endpoint" example:"https://oss-cn-hangzhou.aliyuncs.com"`
	// 阿里云 AccessKey ID
	OssAccessKeyId string `json:"oss_access_key_id" example:"xxxxxxxxxxxxx"`
	// 阿里云 AccessKey Secret
	OssAccessKeySecret string `json:"oss_access_key_secret" example:"xxxxxxxxxxxxx"`
	// 目标 bucket 名
	OssBucket string `json:"oss_bucket" example:"xxxxx"`
}
