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
