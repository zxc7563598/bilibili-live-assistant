package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
)

// shopEncryptedBody 对应 hejunjie-encrypted-request encryptRequest 的线格式。
type shopEncryptedBody struct {
	EnData     string `json:"en_data"`
	EncPayload string `json:"enc_payload"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

// ShopEncrypt 商城 API 请求体解密中间件。
//
// 验证并解密前端 encryptRequest 加密的请求体，将解密后的 JSON 还原为
// c.Request.Body，使下游 ShouldBindJSON 绑定逻辑无需感知加密。
//
// requireEncryption 对应 config.yaml crypto.require_encryption：
//   - true：明文（无加密字段）请求体一律拒绝，仅接受加密请求
//   - false：明文与加密均接受（纯 HTTP 部署前端无法使用 Web Crypto 加密，必须放行明文）
//
// 规则：
//   - GET/HEAD/OPTIONS 等无请求体的请求直接放行（公钥下发 / manifest 不受影响）
//   - 非 JSON 或明文（无加密字段）请求体：requireEncryption 为 true 时拒绝，否则放行
//   - 加密字段部分缺失视为损坏的加密体，直接拒绝
//   - 加密字段齐全则强制验签 + 解密，失败返回 10009
func ShopEncrypt(requireEncryption bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 无请求体的方法不处理
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil || len(bytes.TrimSpace(body)) == 0 {
			c.Next()
			return
		}
		var enc shopEncryptedBody
		if err := json.Unmarshal(body, &enc); err != nil {
			// 非 JSON 请求体：还原后交给下游处理（ShouldBindJSON 会给出参数错误）
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}
		// 部分加密字段 → 损坏的加密体，拒绝
		if (enc.EnData != "") != (enc.EncPayload != "") ||
			(enc.EnData != "") != (enc.Sign != "") {
			response.Error(c, "", 10009)
			c.Abort()
			return
		}
		// 明文请求体（无加密字段）
		if enc.EnData == "" {
			if requireEncryption {
				response.Error(c, "", 10009)
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}
		// 加密请求体：验签 + 解密
		plain, err := crypto.DecryptRequest(enc.EnData, enc.EncPayload, enc.Timestamp, enc.Sign)
		if err != nil {
			response.Error(c, "", 10009)
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Next()
	}
}
