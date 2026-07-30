package menu

// MenuItem 单条菜单权限树
type MenuItem struct {
	ID          uint64     `json:"id"`
	Code        string     `json:"code"`
	Enable      bool       `json:"enable"`
	Show        bool       `json:"show"`
	KeepAlive   bool       `json:"keep_alive"`
	Layout      string     `json:"layout"`
	Type        string     `json:"type"`
	ParentID    uint64     `json:"parent_id"`
	Name        string     `json:"name"`
	Icon        string     `json:"icon"`
	Path        string     `json:"path"`
	Component   string     `json:"component"`
	Order       int        `json:"order"`
	Redirect    string     `json:"redirect"`
	Method      string     `json:"method"`
	Description string     `json:"description"`
	Children    []MenuItem `json:"children"`
}

// Save 请求入参
type SaveReq struct {
	ID        *uint64 `json:"id"`
	Code      string  `json:"code"`
	Enable    bool    `json:"enable"`
	Show      bool    `json:"show"`
	KeepAlive bool    `json:"keep_alive"`
	Layout    string  `json:"layout"`
	Type      string  `json:"type"`
	ParentID  uint64  `json:"parent_id"`
	Name      string  `json:"name"`
	Icon      string  `json:"icon"`
	Path      string  `json:"path"`
	Component string  `json:"component"`
	Order     int     `json:"order"`
}
