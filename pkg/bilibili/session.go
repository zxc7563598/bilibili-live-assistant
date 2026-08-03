package bilibili

// Session 保存 B站 用户身份信息
// Cookie 由 http.Client 内置的 CookieJar 自动管理，不在此结构体中维护
type Session struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Face     string `json:"face"`
	Buvid    string `json:"buvid"`
}
