package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	rsaPrivateKeyFile = "rsa_private_key.pem"
	rsaPublicKeyFile  = "rsa_public_key.pem"
	rsaKeyBits        = 2048
)

var (
	rsaOnce         sync.Once
	rsaPriv         *rsa.PrivateKey
	rsaPublicKeyB64 string
	rsaInitErr      error
)

// keyDir 返回二进制所在目录
func keyDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// EnsureRSAKeyPair 确保 RSA 密钥对已存在；不存在则生成并落盘到二进制同目录。
// 返回 SPKI DER 的 base64 编码公钥（供前端 importKey("spki",...) 使用）。
// 幂等：sync.Once 缓存，bootstrap 启动调用一次 + handler 每请求调用只付出一次锁检查。
func EnsureRSAKeyPair() (string, error) {
	rsaOnce.Do(func() {
		rsaPriv, rsaPublicKeyB64, rsaInitErr = ensureRSAKeyPair(keyDir())
	})
	return rsaPublicKeyB64, rsaInitErr
}

// RSAPrivateKey 返回当前 RSA 私钥
// 与 EnsureRSAKeyPair 共享 sync.Once，公私钥始终同源
func RSAPrivateKey() (*rsa.PrivateKey, error) {
	if _, err := EnsureRSAKeyPair(); err != nil {
		return nil, err
	}
	return rsaPriv, nil
}

func ensureRSAKeyPair(dir string) (*rsa.PrivateKey, string, error) {
	privPath := filepath.Join(dir, rsaPrivateKeyFile)
	pubPath := filepath.Join(dir, rsaPublicKeyFile)
	// 已有私钥 → 复用并派生公钥（不重复生成）
	if priv, err := loadPrivateKey(privPath); err == nil {
		pubB64, err := publicKeyFrom(priv, pubPath)
		return priv, pubB64, err
	} else if !os.IsNotExist(err) {
		return nil, "", fmt.Errorf("读取 RSA 私钥失败: %w", err) // 损坏的现有密钥视为致命，不静默覆盖
	}
	// 生成新密钥对
	priv, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return nil, "", fmt.Errorf("生成 RSA 密钥对失败: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, "", fmt.Errorf("创建密钥目录失败: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY", // PKCS#1
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	f, err := os.OpenFile(privPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			priv, err = loadPrivateKey(privPath)
			if err != nil {
				return nil, "", fmt.Errorf("读取已存在的 RSA 私钥失败: %w", err)
			}
		} else {
			return nil, "", fmt.Errorf("写入 RSA 私钥失败: %w", err)
		}
	} else {
		if _, werr := f.Write(privPEM); werr != nil {
			f.Close()
			return nil, "", fmt.Errorf("写入 RSA 私钥失败: %w", werr)
		}
		if cerr := f.Close(); cerr != nil {
			return nil, "", fmt.Errorf("关闭 RSA 私钥文件失败: %w", cerr)
		}
	}
	pubB64, err := publicKeyFrom(priv, pubPath)
	return priv, pubB64, err
}

// publicKeyFrom 从私钥派生公钥，补写 PEM 文件（缺失时，仅供人工查看），并返回 base64 SPKI。
func publicKeyFrom(priv *rsa.PrivateKey, pubPath string) (string, error) {
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey) // SubjectPublicKeyInfo → SPKI
	if err != nil {
		return "", fmt.Errorf("编码 RSA 公钥失败: %w", err)
	}
	if _, err := os.Stat(pubPath); os.IsNotExist(err) {
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
		_ = os.WriteFile(pubPath, pubPEM, 0644) // 公钥非敏感，补写失败不阻塞
	}
	return base64.StdEncoding.EncodeToString(pubDER), nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM 解码失败")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// PublicKeyID 由公钥内容派生稳定标识（sha256 前 16 位 hex）
func PublicKeyID(pubKeyB64 string) string {
	sum := sha256.Sum256([]byte(pubKeyB64))
	return hex.EncodeToString(sum[:])[:16]
}

// HMACSHA256 计算 HMAC-SHA256，返回小写 hex 编码
func HMACSHA256(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
