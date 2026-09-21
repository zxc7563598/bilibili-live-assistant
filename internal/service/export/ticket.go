package export

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ticketType 凭证类型标识，防止把其它凭证当作凭证使用
const ticketType = "export-ticket"

// ticketPayload 凭证内容
type ticketPayload struct {
	Typ  string          `json:"t"` // 固定 ticketType
	Sub  int64           `json:"s"` // 管理员 ID，用于日志溯源
	Mod  string          `json:"m"` // 导出模块
	Cols []string        `json:"c"` // 列 key，顺序即列序
	Filt json.RawMessage `json:"f"` // Normalize 之后的规范化筛选条件
	Lang string          `json:"l"` // 获取凭证时快照的语言，决定表头与枚举文案
	Exp  int64           `json:"e"` // 过期时间（unix 秒）
}

// signer 凭证签发与校验。
type signer struct {
	key []byte
}

// newSigner 由 jwt.secret 派生子密钥
func newSigner(secret string) *signer {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ticketType + ":v1"))
	return &signer{key: mac.Sum(nil)}
}

// sign 签发凭证，格式为：base64url(payload).base64url(hmac)
func (s *signer) sign(p ticketPayload) (string, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	return body + "." + s.digest(body), nil
}

// digest 计算 body 的 HMAC 摘要(base64url)
func (s *signer) digest(body string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// verify 校验凭证并返回内容。
func (s *signer) verify(token string) (ticketPayload, error) {
	var p ticketPayload
	body, sig, ok := strings.Cut(token, ".")
	if !ok {
		return p, errors.New("凭证格式不正确")
	}
	// 先验签再解析内容：签名不通过就不该去看里面写了什么
	if !hmac.Equal([]byte(sig), []byte(s.digest(body))) {
		return p, errors.New("凭证签名校验失败")
	}
	raw, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, err
	}
	if p.Typ != ticketType {
		return p, errors.New("凭证类型不匹配")
	}
	if p.Exp <= 0 || time.Now().Unix() > p.Exp {
		return p, errors.New("凭证已过期")
	}
	if p.Mod == "" || len(p.Cols) == 0 {
		return p, errors.New("凭证内容不完整")
	}
	return p, nil
}
