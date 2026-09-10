// Package oss 提供阿里云 OSS 上传的独立封装，供各业务模块复用。
//
// 面向后台管理员维护、低频变动、可公开读的内容（如积分商城商品图）：
// 先由业务方把文件落到服务器本地（参考 pkg/fileutil），再调用 UploadFile
// 同步上传到 OSS 并拿到默认域名的公开链接；仅凭入库的 objectKey 可用
// PublicURL 重建链接，无需重复上传。本包无任何业务依赖。
//
// 使用示例：
//
//	client, err := oss.New(oss.Config{
//	    Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com", // 不带 bucket 名
//	    AccessKeyID:     "...",
//	    AccessKeySecret: "...",
//	    Bucket:          "mall-images",
//	})
//	if err != nil { /* ... */ }
//	url, err := client.UploadFile("/app/uploads/product/1.jpg", "mall/product/1.jpg")
//	if err != nil { /* ... */ }
//	// 仅凭已入库 key 重建公开链接：
//	_ = client.PublicURL("mall/product/1.jpg")
package oss

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// checkTimeout CheckConfig 的请求总超时。SDK 本身无整体超时（单次连接最多可挂约 30s），
// 配置探测应尽快给结论，避免 endpoint 配错（不可达、DNS 卡住）时长时间阻塞。
const checkTimeout = 5 * time.Second

// Config 构造客户端所需参数，四项均必填。
type Config struct {
	// Endpoint 完整 OSS 地址，如 https://oss-cn-hangzhou.aliyuncs.com，不含 bucket 名；
	// 缺 scheme 时自动补 https://。
	Endpoint string
	// AccessKeyID 阿里云 AccessKey ID（建议使用 RAM 子账号，最小授权到目标 bucket）。
	AccessKeyID string
	// AccessKeySecret 阿里云 AccessKey Secret。
	AccessKeySecret string
	// Bucket 目标 bucket 名。
	Bucket string
}

// Client 绑定单个 bucket 的 OSS 客户端，可安全地并发调用。
type Client struct {
	// probe 仅供 CheckConfig 使用的探测客户端，挂 5s 总超时，不影响上传用的默认长超时连接
	probe      *aliyunoss.Client
	bucket     *aliyunoss.Bucket
	endpoint   string
	bucketName string
}

// New 根据配置创建 OSS 客户端。仅做参数校验与句柄构造，不发任何网络请求，
// 连接错误会推迟到首次上传时暴露。
func New(cfg Config) (*Client, error) {
	switch {
	case cfg.Endpoint == "":
		return nil, errors.New("oss: 缺少必填配置 endpoint")
	case cfg.AccessKeyID == "":
		return nil, errors.New("oss: 缺少必填配置 access_key_id")
	case cfg.AccessKeySecret == "":
		return nil, errors.New("oss: 缺少必填配置 access_key_secret")
	case cfg.Bucket == "":
		return nil, errors.New("oss: 缺少必填配置 bucket")
	}

	endpoint := normalizeEndpoint(cfg.Endpoint)
	host, err := endpointHost(endpoint)
	if err != nil {
		return nil, err
	}
	sdkClient, err := aliyunoss.New(endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss: 初始化客户端失败: %w", err)
	}
	bucket, err := sdkClient.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("oss: 获取 bucket %s 句柄失败: %w", cfg.Bucket, err)
	}
	// 探测专用客户端：http.Client.Timeout 对整次请求（含 DNS/连接/响应）设硬上限，
	// 专给 CheckConfig 用，避免上传等业务请求也被 5s 截断
	probe, err := aliyunoss.New(endpoint, cfg.AccessKeyID, cfg.AccessKeySecret, aliyunoss.HTTPClient(&http.Client{Timeout: checkTimeout}))
	if err != nil {
		return nil, fmt.Errorf("oss: 初始化校验客户端失败: %w", err)
	}
	return &Client{probe: probe, bucket: bucket, endpoint: host, bucketName: cfg.Bucket}, nil
}

// UploadFile 把服务器本地文件 localPath 同步上传到 OSS 的 objectKey 下，返回公开可访问 URL。
// objectKey 为空时按 <纳秒时间戳>_<8位随机hex>[扩展名] 自动生成（与 pkg/fileutil 本地命名一致）。
// 上传自动携带 Content-Type（按 objectKey 扩展名推断，兜底 application/octet-stream）
// 并把 object ACL 置为 public-read，bucket 即使保持 private 也能拿到公开链接。
func (c *Client) UploadFile(localPath, objectKey string) (string, error) {
	if objectKey == "" {
		name, err := uploadFileName(localPath)
		if err != nil {
			return "", err
		}
		objectKey = name
	}
	opts := []aliyunoss.Option{
		// 不显式声明 Content-Type 会被 OSS 存成 application/octet-stream，浏览器可能直接下载
		aliyunoss.ContentType(contentTypeFor(objectKey)),
		aliyunoss.ObjectACL(aliyunoss.ACLPublicRead),
	}
	if err := c.bucket.PutObjectFromFile(objectKey, localPath, opts...); err != nil {
		return "", fmt.Errorf("oss: 上传文件 %s 失败: %w", objectKey, err)
	}
	return c.PublicURL(objectKey), nil
}

// PublicURL 根据已入库的 objectKey 重建 OSS 默认域名公开链接（https://<bucket>.<endpoint>/<key>），
// 不触发任何网络请求。objectKey 按 / 分段做 URL 转义以保留目录层级。
func (c *Client) PublicURL(objectKey string) string {
	return "https://" + c.bucketName + "." + c.endpoint + "/" + escapeObjectKey(objectKey)
}

// CheckConfig 通过一次真实的 OSS 请求校验配置是否可用，供管理员录入 Endpoint/AccessKey/bucket 后自查：
// Endpoint 可达、AK/SK 有效、bucket 存在且可访问。成功返回 nil。
// 请求挂 5 秒总超时（专用探测客户端，不影响上传），endpoint 配错时也能较快给出结论。
// 注意：校验走 GetBucketInfo，需要 AccessKey 对该 bucket 有只读（oss:GetBucketInfo）权限；
// 若只授权了对象上传（oss:PutObject），会返回 403 权限不足——此时上传功能本身仍可用。
func (c *Client) CheckConfig() error {
	_, err := c.probe.GetBucketInfo(c.bucketName)
	if err == nil {
		return nil
	}
	var serr *aliyunoss.ServiceError
	if errors.As(err, &serr) {
		switch serr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("oss: 配置校验失败：bucket %q 不存在或不属于该 AccessKey: %w", c.bucketName, err)
		case http.StatusForbidden:
			return fmt.Errorf("oss: 配置校验失败：AccessKey 无效，或权限不足（需 oss:GetBucketInfo 读权限）: %w", err)
		}
	}
	return fmt.Errorf("oss: 配置校验失败：Endpoint 不可达或网络异常: %w", err)
}

// normalizeEndpoint 补齐 scheme 并去掉尾部斜杠，SDK 要求端点带完整 scheme。
func normalizeEndpoint(endpoint string) string {
	e := strings.TrimSpace(endpoint)
	e = strings.TrimRight(e, "/")
	if !strings.Contains(e, "://") {
		e = "https://" + e
	}
	return e
}

// endpointHost 提取端点去掉 scheme 后的 host，用于拼默认域名公开链接。
func endpointHost(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("oss: 非法 endpoint %q", endpoint)
	}
	return u.Hostname(), nil
}

// uploadFileName 生成唯一的 objectKey 文件名：<纳秒时间戳>_<8位随机hex>[安全扩展名]，
// 命名规则与 pkg/fileutil.uploadFileName 保持一致。
func uploadFileName(localPath string) (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("oss: 生成随机文件名失败: %w", err)
	}
	ext := safeExt(path.Ext(localPath))
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(buf), ext), nil
}

// safeExt 返回净化后的扩展名（含点、小写字母数字、长度 2-6），不合法返回空串。
func safeExt(ext string) string {
	ext = strings.ToLower(ext)
	for i := 1; i < len(ext); i++ {
		c := ext[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return ""
		}
	}
	if len(ext) < 2 || len(ext) > 6 {
		return ""
	}
	return ext
}

// contentTypeFor 按 objectKey 扩展名推断 Content-Type，未知扩展名兜底 application/octet-stream。
func contentTypeFor(objectKey string) string {
	if ct := mime.TypeByExtension(strings.ToLower(path.Ext(objectKey))); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// escapeObjectKey 按 / 分段转义 objectKey，避免整串 PathEscape 把目录分隔符变成 %2F。
func escapeObjectKey(key string) string {
	segments := strings.Split(strings.TrimPrefix(key, "/"), "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return strings.Join(segments, "/")
}
