package appconfig

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	rdb *redis.Client
}

// New 创建 Handler 实例
func New(rdb *redis.Client) *Handler {
	return &Handler{rdb: rdb}
}

// @Summary 获取 App 的 Manifest 信息
// @Description 获取 App 的 Manifest 信息，用于前端构建 PWA 应用
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppShopManifestResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/manifest [get]
func (h *Handler) GetManifest(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	response.Success(c, lang, resp.AppShopManifestResp{
		Name:            "哎呀又胖啦的积分商城",
		ShortName:       "哎呀商城",
		Description:     "关于我也不知道在哪里才能看到的说明",
		ThemeColor:      "#f5f6f8",
		BackgroundColor: "#f5f6f8",
		Favicon:         "https://cdn.hejunjie.life/avatars/AIOVTUE-%E9%9B%AA-1782904215522.PNG",
		AppleTouchIcon:  "https://cdn.hejunjie.life/avatars/AIOVTUE-%E9%9B%AA-1782904215522.PNG",
		StartURL:        "/shop/",
		Scope:           "/shop/",
		Display:         "standalone",
		Icons: []resp.AppShopManifestIcon{
			resp.AppShopManifestIcon{
				Src:     "https://cdn.hejunjie.life/avatars/AIOVTUE-%E9%9B%AA-1782904215522.PNG",
				Sizes:   "512x512",
				Type:    "image/png",
				Purpose: "any",
			},
		},
	})
}

// @Summary 获取 RSA 公钥（带 HMAC 验签）
// @Description 获取用于前端 RSA-OAEP 加密的 RSA 公钥（SPKI DER 的 base64），并附带 HMAC 签名供前端验签
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppPublicKeyResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/public-key [get]
func (h *Handler) GetPublicKey(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	pubKeyB64, err := crypto.EnsureRSAKeyPair()
	if err != nil {
		response.Error(c, lang, 60001)
		return
	}
	keyID := crypto.PublicKeyID(pubKeyB64)
	ts := time.Now().Unix()
	// 签名消息格式必须与前端一致：pubkey:<key_id><public_key><timestamp>
	msg := "pubkey:" + keyID + pubKeyB64 + strconv.FormatInt(ts, 10)
	response.Success(c, lang, resp.AppPublicKeyResp{
		KeyID:     keyID,
		PublicKey: pubKeyB64,
		Timestamp: ts,
		Sign:      crypto.HMACSHA256(msg, crypto.SignSecret),
	})
}
