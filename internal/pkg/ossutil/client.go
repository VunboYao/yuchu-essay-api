// Package ossutil 封装阿里云 OSS STS 签发与下载预签名；未配置时退回开发期伪签名。
package ossutil

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"course_server/internal/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const defaultUploadDir = "essay-mvp"
const defaultSTSExpireSec = 3600

// STSCredentials Admin 直传所需的临时凭证。
type STSCredentials struct {
	Mode            string `json:"mode"`
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	SecurityToken   string `json:"securityToken"`
	Expiration      string `json:"expiration"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Dir             string `json:"dir"`
	CDNDomain       string `json:"cdnDomain,omitempty"`
}

// Client 持有 OSS 运行时配置。
type Client struct {
	cfg config.OSSConfig
}

func NewClient(cfg config.OSSConfig) *Client {
	return &Client{cfg: cfg}
}

// IsConfigured 主 AK/SK + bucket + endpoint 齐全，可用于预签名下载。
func (c *Client) IsConfigured() bool {
	return c.cfg.IsConfigured()
}

// IsSTSEnabled STS 直传所需角色 ARN 已配置。
func (c *Client) IsSTSEnabled() bool {
	return c.cfg.IsSTSEnabled()
}

// UploadPrefix 按日期生成上传目录前缀（essay-mvp/20260627/）。
func (c *Client) UploadPrefix() string {
	dir := strings.Trim(c.cfg.UploadDir, "/")
	if dir == "" {
		dir = defaultUploadDir
	}
	return dir + "/" + time.Now().Format("20060102") + "/"
}

// IssueSTSCredentials 为 Admin 浏览器直传签发 STS 临时凭证。
func (c *Client) IssueSTSCredentials() (*STSCredentials, error) {
	if !c.IsSTSEnabled() {
		return nil, fmt.Errorf("oss sts not configured: set sts_role_arn")
	}
	prefix := c.UploadPrefix()
	region := regionFromEndpoint(c.cfg.Endpoint)
	stsClient, err := sts.NewClientWithAccessKey(region, c.cfg.AccessKeyID, c.cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create sts client: %w", err)
	}
	req := sts.CreateAssumeRoleRequest()
	req.Scheme = "https"
	req.RoleArn = c.cfg.STSRoleARN
	req.RoleSessionName = "essay-mvp-admin"
	req.DurationSeconds = requests.NewInteger(defaultSTSExpireSec)
	req.Policy = uploadPolicy(c.cfg.Bucket, prefix)
	resp, err := stsClient.AssumeRole(req)
	if err != nil {
		return nil, fmt.Errorf("assume role: %w", err)
	}
	cred := resp.Credentials
	return &STSCredentials{
		Mode:            "oss",
		AccessKeyID:     cred.AccessKeyId,
		AccessKeySecret: cred.AccessKeySecret,
		SecurityToken:   cred.SecurityToken,
		Expiration:      cred.Expiration,
		Bucket:          c.cfg.Bucket,
		Endpoint:        normalizeEndpoint(c.cfg.Endpoint),
		Region:          region,
		Dir:             prefix,
		CDNDomain:       strings.TrimSpace(c.cfg.CDNDomain),
	}, nil
}

// SignDownloadURL 对自家 bucket 文件签发 GET 预签名；未配置或外链时退回伪签名。
func (c *Client) SignDownloadURL(fileURL string, expireSec int64) string {
	if expireSec <= 0 {
		expireSec = 300
	}
	if !c.IsConfigured() {
		return pseudoSignedURL(fileURL, expireSec)
	}
	key, ok := c.parseObjectKey(fileURL)
	if !ok {
		return pseudoSignedURL(fileURL, expireSec)
	}
	client, err := oss.New(normalizeEndpoint(c.cfg.Endpoint), c.cfg.AccessKeyID, c.cfg.AccessKeySecret)
	if err != nil {
		return pseudoSignedURL(fileURL, expireSec)
	}
	bucket, err := client.Bucket(c.cfg.Bucket)
	if err != nil {
		return pseudoSignedURL(fileURL, expireSec)
	}
	signed, err := bucket.SignURL(key, oss.HTTPGet, expireSec)
	if err != nil {
		return pseudoSignedURL(fileURL, expireSec)
	}
	return signed
}

// PublicURL 将 objectKey 拼为对外 HTTPS URL（优先 CDN 域名）。
func (c *Client) PublicURL(objectKey string) string {
	key := strings.TrimPrefix(strings.TrimSpace(objectKey), "/")
	if key == "" {
		return ""
	}
	if domain := strings.TrimSpace(c.cfg.CDNDomain); domain != "" {
		return strings.TrimRight(domain, "/") + "/" + key
	}
	ep := normalizeEndpoint(c.cfg.Endpoint)
	return "https://" + c.cfg.Bucket + "." + ep + "/" + key
}

// ObjectKeyFromURL 从完整 URL 或裸 key 解析 OSS objectKey。
func (c *Client) ObjectKeyFromURL(fileURL string) (string, bool) {
	return c.parseObjectKey(fileURL)
}

func (c *Client) parseObjectKey(fileURL string) (string, bool) {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return "", false
	}
	if !strings.Contains(fileURL, "://") {
		return strings.TrimPrefix(fileURL, "/"), true
	}
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", false
	}
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return "", false
	}
	host := strings.ToLower(u.Host)
	bucket := strings.ToLower(c.cfg.Bucket)
	endpoint := strings.ToLower(normalizeEndpoint(c.cfg.Endpoint))
	if cdn := strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(c.cfg.CDNDomain), "https://"), "http://")); cdn != "" && host == cdn {
		return path, true
	}
	if strings.HasPrefix(host, bucket+".") && strings.Contains(host, endpoint) {
		return path, true
	}
	if host == endpoint && strings.HasPrefix(path, bucket+"/") {
		return strings.TrimPrefix(path, bucket+"/"), true
	}
	return "", false
}

func uploadPolicy(bucket, prefix string) string {
	prefix = strings.Trim(prefix, "/")
	doc := map[string]any{
		"Version": "1",
		"Statement": []map[string]any{{
			"Effect":   "Allow",
			"Action":   []string{"oss:PutObject"},
			"Resource": []string{fmt.Sprintf("acs:oss:*:*:%s/%s/*", bucket, prefix)},
		}},
	}
	b, _ := json.Marshal(doc)
	return string(b)
}

func regionFromEndpoint(endpoint string) string {
	ep := normalizeEndpoint(endpoint)
	parts := strings.Split(ep, ".")
	if len(parts) > 0 && strings.HasPrefix(parts[0], "oss-") {
		return strings.TrimPrefix(parts[0], "oss-")
	}
	return "cn-hangzhou"
}

func normalizeEndpoint(endpoint string) string {
	ep := strings.TrimSpace(endpoint)
	ep = strings.TrimPrefix(ep, "https://")
	ep = strings.TrimPrefix(ep, "http://")
	return strings.TrimSuffix(ep, "/")
}

// pseudoSignedURL 开发期伪签名（拼接 expires）；OSS 未配置或外链时兼容。
func pseudoSignedURL(fileURL string, expireSec int64) string {
	sep := "?"
	if strings.Contains(fileURL, "?") {
		sep = "&"
	}
	return fileURL + sep + "expires=" + fmt.Sprintf("%d", time.Now().Add(time.Duration(expireSec)*time.Second).Unix())
}
