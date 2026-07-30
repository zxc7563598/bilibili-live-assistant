package api

// =========================================================================
// B站 API 端点集中管理
//
// 所有 B站 API URL 均在此定义，避免硬编码分散在多个文件中。
// 部分端点包含 %d 等格式化占位符，调用时需通过 fmt.Sprintf 填入参数。
// =========================================================================

// --- 登录认证（auth 包）---

const (
	// EndpointQRCodeGenerate 获取登录二维码，GET 请求，无需登录态
	EndpointQRCodeGenerate = "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"

	// EndpointQRCodePoll 轮询扫码状态，qrcode_key 拼接到 URL 末尾
	EndpointQRCodePoll = "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key="

	// EndpointNav B站导航信息接口，同时用于获取 WBI 密钥和当前登录用户信息
	EndpointNav = "https://api.bilibili.com/x/web-interface/nav"

	// EndpointFingerSpy 获取设备指纹 Buvid3/Buvid4，无需登录态
	EndpointFingerSpy = "https://api.bilibili.com/x/frontend/finger/spi"
)

// --- 直播间（room 包）---

const (
	// EndpointRoomInfo 直播间基本信息，查询参数 room_id=%d
	EndpointRoomInfo = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%d"

	// EndpointDanmuInfo 弹幕 WebSocket 连接信息，查询参数为 WBI 签名后的完整 query string
	EndpointDanmuInfo = "https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?%s"

	// EndpointBarragePermission 当前用户在直播间的弹幕发送权限，查询参数 room_id=%d
	EndpointBarragePermission = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByUser?room_id=%d"

	// EndpointSendDanmu 发送弹幕（form-urlencoded POST）
	EndpointSendDanmu = "https://api.live.bilibili.com/msg/send"
)

// --- 排行榜与大航海 ---

const (
	// EndpointOnlineGoldRank 在线金瓜子榜，查询参数 ruid=%d&roomId=%d
	EndpointOnlineGoldRank = "https://api.live.bilibili.com/xlive/general-interface/v1/rank/getOnlineGoldRank?ruid=%d&roomId=%d&page=1&pageSize=5000"

	// EndpointGuardTopList 大航海（舰长/提督/总督）列表，查询参数 ruid=%d&roomid=%d
	EndpointGuardTopList = "https://api.live.bilibili.com/xlive/app-room/v2/guardTab/topListNew?ruid=%d&roomid=%d&page=1&page_size=20&typ=5&platform=web"
)

// --- 禁言管理 ---

const (
	// EndpointAddSilentUser 禁言用户（form-urlencoded POST）
	EndpointAddSilentUser = "https://api.live.bilibili.com/xlive/web-ucenter/v1/banned/AddSilentUser"

	// EndpointGetSilentUserList 获取禁言用户列表（form-urlencoded POST）
	EndpointGetSilentUserList = "https://api.live.bilibili.com/xlive/web-ucenter/v1/banned/GetSilentUserList"

	// EndpointDelSilentUser 解除禁言（form-urlencoded POST）
	EndpointDelSilentUser = "https://api.live.bilibili.com/banned_service/v1/Silent/del_room_block_user"
)

// --- 用户信息（user 包）---

const (
	// EndpointMasterInfo 主播基本信息，查询参数 uid=%d，无需登录态
	EndpointMasterInfo = "https://api.live.bilibili.com/live_user/v1/Master/info?uid=%d"

	// EndpointStreamerInfo 用户空间信息，查询参数为 WBI 签名后的完整 query string，需要登录态
	EndpointStreamerInfo = "https://api.bilibili.com/x/space/wbi/acc/info?%s"
)
