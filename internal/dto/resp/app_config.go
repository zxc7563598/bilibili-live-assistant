package resp

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
