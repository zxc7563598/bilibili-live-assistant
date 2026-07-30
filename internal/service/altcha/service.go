package altcha

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/rand/v2"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
)

// Service 用于处理 altcha 验证码相关业务逻辑
type Service struct {
	hmacKey string
}

// New 返回一个新的 Altcha Service 实例
func New(hmacKey string) *Service {
	return &Service{hmacKey: hmacKey}
}

// CreateChallenge 生成 altcha 验证码挑战
// 若未配置 hmacKey 则返回 nil 表示不启用验证码
func (s *Service) CreateChallenge(_ context.Context) (*altcha.Challenge, int, error) {
	if s.hmacKey == "" {
		return nil, 0, nil
	}
	counter := rand.IntN(5001) + 5000 // 5000 ~ 10000
	expiresAt := time.Now().Add(10 * time.Minute)
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           "PBKDF2/SHA-256",
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: s.hmacKey,
		Cost:                5000,
		Counter:             &counter,
		ExpiresAt:           &expiresAt,
		KeyLength:           32,
	})
	if err != nil {
		return nil, 60116, err
	}
	return &challenge, 0, nil
}

// IsEnabled 返回 altcha 验证码是否已配置并启用
func (s *Service) IsEnabled() bool {
	return s.hmacKey != ""
}

// VerifySolution 验证前端提交的 altcha 验证码 payload
// captcha 为 altcha widget 提交的 base64 编码 JSON 字符串
// 返回 (errorCode, error)，errorCode == 0 表示验证通过
func (s *Service) VerifySolution(_ context.Context, captcha *string) (int, error) {
	// 未配置 hmacKey 时不进行验证
	if s.hmacKey == "" {
		return 0, nil
	}
	// 已配置但未提交验证码
	if captcha == nil || *captcha == "" {
		return 10109, nil
	}
	// 解码 base64 payload
	decoded, err := base64.StdEncoding.DecodeString(*captcha)
	if err != nil {
		return 10001, err
	}
	// 解析 altcha payload
	var payload altcha.Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return 10001, err
	}
	// 验证 solution
	result, err := altcha.VerifySolution(altcha.VerifySolutionOptions{
		Challenge:           payload.Challenge,
		Solution:            payload.Solution,
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: s.hmacKey,
	})
	if err != nil {
		return 60116, err
	}
	if result.Expired {
		return 40106, nil
	}
	if !result.Verified {
		return 40105, nil
	}
	return 0, nil
}
