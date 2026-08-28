package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// SignSecret 前端加密/签名 HMAC 密钥
var SignSecret = ""

// RequestTimestampWindow 请求时间偏差窗口（秒）
var RequestTimestampWindow = int64(60)

// SetSignSecret 注入签名密钥
func SetSignSecret(s string, r int64) {
	SignSecret = s
	RequestTimestampWindow = r
}

// encryptedKeyMaterial 对应 JS encryptRequest 打包进 enc_payload 的密钥材料。
type encryptedKeyMaterial struct {
	V      int    `json:"v"`
	AesKey string `json:"aes_key"`
	AesIv  string `json:"aes_iv"`
	Nonce  string `json:"nonce"`
}

// DecryptRequest 校验并解密 hejunjie-encrypted-request v3 加密的请求体，
// 返回 JS 端 JSON.stringify(data) 的原始 UTF-8 字节。
//
// 与前端线格式对应（已逐行核对 dist/encrypt.js）：
//   - sign = hex(HMAC-SHA256(en_data + enc_payload + timestamp, SignSecret))
//   - enc_payload = base64(RSA-OAEP(SHA-1) over {"v":v,"aes_key":b64,"aes_iv":b64,"nonce":hex})
//   - en_data = base64(AES-256-GCM(JSON.stringify(data)))，GCM tag 附在密文尾部
//
// 顺序为先做廉价认证（HMAC / 时间窗），再做重 RSA 解密，避免未认证垃圾触发私钥操作。
func DecryptRequest(enData, encPayload string, timestamp int64, sign string) ([]byte, error) {
	// HMAC-SHA256 验签（hex 解码后常量时间比对）
	mac := hmac.New(sha256.New, []byte(SignSecret))
	mac.Write([]byte(enData + encPayload + strconv.FormatInt(timestamp, 10)))
	want := mac.Sum(nil)
	got, err := hex.DecodeString(sign)
	if err != nil || !hmac.Equal(want, got) {
		return nil, errors.New("签名校验失败")
	}
	// 时间窗防重放
	if now := time.Now().Unix(); now < timestamp-RequestTimestampWindow || now > timestamp+RequestTimestampWindow {
		return nil, errors.New("时间戳超出允许窗口")
	}
	// RSA-OAEP(SHA-1) 解密密钥材料
	priv, err := RSAPrivateKey()
	if err != nil {
		return nil, err
	}
	ct, err := base64.StdEncoding.DecodeString(encPayload)
	if err != nil {
		return nil, errors.New("enc_payload base64 解码失败")
	}
	payload, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, ct, nil)
	if err != nil {
		return nil, errors.New("enc_payload 解密失败")
	}
	// 解析密钥材料
	var km encryptedKeyMaterial
	if err := json.Unmarshal(payload, &km); err != nil {
		return nil, errors.New("密钥材料 JSON 解析失败")
	}
	aesKey, err := base64.StdEncoding.DecodeString(km.AesKey)
	if err != nil || len(aesKey) != 32 {
		return nil, errors.New("aes_key 格式非法")
	}
	aesIv, err := base64.StdEncoding.DecodeString(km.AesIv)
	if err != nil || len(aesIv) != 12 {
		return nil, errors.New("aes_iv 格式非法")
	}
	// AES-256-GCM 解密业务数据（无 AAD；tag 已附在密文尾部）
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.New("AES 密钥初始化失败")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("GCM 初始化失败")
	}
	ctData, err := base64.StdEncoding.DecodeString(enData)
	if err != nil {
		return nil, errors.New("en_data base64 解码失败")
	}
	plain, err := gcm.Open(nil, aesIv, ctData, nil)
	if err != nil {
		return nil, errors.New("en_data 解密失败")
	}
	return plain, nil
}
