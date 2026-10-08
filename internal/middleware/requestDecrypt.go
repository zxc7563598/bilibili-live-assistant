package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
)

// encryptedRequestBody 对应 hejunjie-encrypted-request encryptRequest 的线格式。
type encryptedRequestBody struct {
	EnData     string `json:"en_data"`
	EncPayload string `json:"enc_payload"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

// RequestDecrypt 请求体解密中间件，商城与管理后台共用。
//
// 验证并解密前端 encryptRequest 加密的请求体，将解密后的 JSON 还原为
// c.Request.Body，使下游 ShouldBindJSON 绑定逻辑无需感知加密。
//
// requireEncryption 对应 config.yaml crypto.require_encryption：
//   - true：明文（无加密字段）请求体一律拒绝，仅接受加密请求
//   - false：明文与加密均接受（纯 HTTP 部署前端无法使用 Web Crypto 加密，必须放行明文）
//
// 规则：
//   - GET/HEAD/OPTIONS 等无请求体的请求直接放行（公钥下发、WebSocket 升级、原生下载不受影响）
//   - 表单类请求体（multipart / urlencoded）不做处理
//   - 非 JSON 或明文（无加密字段）请求体：requireEncryption 为 true 时拒绝，否则放行
//   - 加密字段部分缺失视为损坏的加密体，直接拒绝
//   - 加密字段齐全则强制验签 + 解密，失败返回 i18n.CodeDecryptFailed
//
// 注意：只能挂在需要解密的各个路由分组上，**不要挂到引擎全局**。全局中间件先于分组中间件执行，
// 会把密文解成明文写回 body，分组上再挂一次就会把明文当成"未加密请求"，require_encryption
// 为 true 时直接拒掉自己解过的请求。
func RequestDecrypt(requireEncryption bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 无请求体的方法不处理
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		// 表单类请求体不可能是密文，且上传接口的 body 可能有几百 MB（旧版数据导入上限 500MB），
		// 绝不能先读进内存再判断，这里直接放行给下游解析
		if isFormContentType(c.GetHeader("Content-Type")) {
			c.Next()
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil || len(bytes.TrimSpace(body)) == 0 {
			c.Next()
			return
		}
		var enc encryptedRequestBody
		if err := json.Unmarshal(body, &enc); err != nil {
			// 非 JSON 请求体：还原后交给下游处理（ShouldBindJSON 会给出参数错误）
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}
		// 部分加密字段 → 损坏的加密体，拒绝
		if (enc.EnData != "") != (enc.EncPayload != "") ||
			(enc.EnData != "") != (enc.Sign != "") {
			logReject(c, "加密字段不完整", nil)
			response.Error(c, "", i18n.CodeDecryptFailed)
			c.Abort()
			return
		}
		// 明文请求体（无加密字段）
		if enc.EnData == "" {
			if requireEncryption {
				logReject(c, "明文请求被拒（require_encryption=true，请确认前端构建已注入 VITE_SIGN_SECRET 且与 crypto.sign_secret 一致）", nil)
				response.Error(c, "", i18n.CodeDecryptFailed)
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
			logReject(c, "解密失败", err)
			response.Error(c, "", i18n.CodeDecryptFailed)
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Next()
	}
}

// isFormContentType 判断是否为表单类请求体。
//
// multipart 与 urlencoded 都不可能是加密请求体（加密后一定是 JSON），
// 提前放行可以避免把上传的大 body 整包读进内存。
func isFormContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "multipart/form-data") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded")
}

// logReject 记录请求被拒的原因。
//
// 前端只会拿到 10009「请求参数不合法」，仅凭它分不清是验签失败、时间戳超窗、
// RSA / AES 解密失败，还是前端压根没加密（构建时漏了 VITE_SIGN_SECRET），
// 排查时全靠这条日志。注意不要打印 sign_secret 与完整密文。
func logReject(c *gin.Context, reason string, cause error) {
	if cause != nil {
		reason += ": " + cause.Error()
	}
	log.Printf("[middleware.decrypt] 拒绝请求 method=%s path=%s ip=%s content_type=%q reason=%s",
		c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.GetHeader("Content-Type"), reason)
}
