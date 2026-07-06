# 阿里云 OSS / STS / CDN 配置指南

本文档用于配置后台上传、公开图片访问和私有资料预签名下载。域名口径见 [domain-vunbo.md](./domain-vunbo.md)。

## 一、能力分工

```text
Admin 浏览器
  → POST /admin/api/v1/upload/sts
  → 获取 STS 临时凭证
  → 直传 OSS 指定目录

小程序下载资料
  → GET /api/v1/download-files/:id/sign-url
  → 后端校验免费下载/权益
  → 后端返回 OSS 预签名 URL
```

凭证分工：

| 凭证 | 用途 |
|---|---|
| RAM 用户 AK/SK | 后端签发 STS、生成私有下载预签名 URL |
| STS 临时凭证 | Admin 浏览器直传 OSS |
| CDN 域名 | 展示封面、正文图、公开图片资源 |

## 二、需要准备的信息

生产 `config.yaml` 的 `oss` 段需要以下值：

| 字段 | 来源 | 示例 |
|---|---|---|
| `endpoint` | Bucket 概览中的外网 Endpoint | `oss-cn-guangzhou.aliyuncs.com` |
| `bucket` | Bucket 名称 | `essay-mvp` |
| `access_key_id` | RAM 用户 AccessKey ID | `LTAI...` |
| `access_key_secret` | RAM 用户 AccessKey Secret | 仅写入服务器配置 |
| `sts_role_arn` | RAM 角色 ARN | `acs:ram::<账号ID>:role/essay-mvp-oss-upload` |
| `cdn_domain` | CDN 自定义域名 | `https://cdn.vunbo.com` |
| `upload_dir` | 上传目录前缀 | `essay-mvp` |

## 三、Bucket 配置

建议配置：

- 地域：与后端服务器同区域，降低延迟和费用。
- 存储类型：标准存储。
- 读写权限：私有。
- 版本控制：一期关闭。
- CORS：允许本地 Admin 和生产 Admin Origin。

CORS：

| 项 | 值 |
|---|---|
| Origin | `http://localhost:5173`、`https://admin.vunbo.com` |
| Methods | `GET, PUT, POST, HEAD` |
| Allow-Headers | `*` |
| Expose-Headers | `ETag, x-oss-request-id` |

## 四、RAM 用户与 STS 角色

### 1. 创建 RAM 用户

用户建议命名为 `essay-mvp-api`，只开启 OpenAPI/AccessKey 访问，不开控制台登录。

该用户需要：

- `oss:GetObject`：后端生成下载预签名 URL。
- `sts:AssumeRole`：后端签发 Admin 上传临时凭证。

用户策略示例：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["oss:GetObject"],
      "Resource": ["acs:oss:*:*:<bucket>/*"]
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Resource": ["acs:ram::*:role/essay-mvp-oss-upload"]
    }
  ]
}
```

### 2. 创建 RAM 角色

角色建议命名为 `essay-mvp-oss-upload`。

信任策略允许 RAM 用户 `essay-mvp-api` 扮演该角色：

```json
{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "RAM": [
          "acs:ram::<主账号ID>:user/essay-mvp-api"
        ]
      }
    }
  ],
  "Version": "1"
}
```

角色权限只允许上传对象：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["oss:PutObject"],
      "Resource": ["acs:oss:*:*:<bucket>/*"]
    }
  ]
}
```

后端在 AssumeRole 时还会追加会话策略，将上传范围收窄到当日目录，例如 `essay-mvp/20260704/*`。

### 3. 常见授权错误

| 现象 | 处理 |
|---|---|
| `mode: placeholder` | `config.yaml` 的 OSS 配置不完整 |
| `failed to issue oss sts credentials` | RAM 用户缺少 `sts:AssumeRole`，或角色信任策略未允许该用户 |
| PUT 403 且 XML 含 `AssumedRoleUser` | 扮演角色成功，但角色未绑定有效 `oss:PutObject` 策略 |
| CORS 报错 | Bucket CORS 的 Origin 或 Method 未覆盖当前 Admin 地址 |
| `SignatureDoesNotMatch` | 检查 endpoint、region、代理和系统时间 |

## 五、填写 `config.yaml`

生产服务器复制 [../config.prod.yaml.example](../config.prod.yaml.example) 为 `config.yaml` 后填写：

```yaml
oss:
  endpoint: "oss-cn-guangzhou.aliyuncs.com"
  bucket: "essay-mvp"
  access_key_id: "LTAI..."
  access_key_secret: "..."
  cdn_domain: "" # CDN 未启用可留空
  sts_role_arn: "acs:ram::<账号ID>:role/essay-mvp-oss-upload"
  upload_dir: "essay-mvp"
```

CDN 生效后改为：

```yaml
oss:
  cdn_domain: "https://cdn.vunbo.com"
```

修改后重启后端：

```bash
systemctl restart essay-mvp
curl -s http://127.0.0.1:8080/health
```

## 六、objectKey 路径规范

| 场景 | Admin `subdir` | key 示例 |
|---|---|---|
| 文章封面 | `covers` | `essay-mvp/20260704/covers/1730_abc.jpg` |
| 正文长图 | `body` | `essay-mvp/20260704/body/1730_def.jpg` |
| 下载资料 | `downloads` | `essay-mvp/20260704/downloads/1730.pdf` |
| 首页图标 | `icons` | `essay-mvp/20260704/icons/1730.png` |

数据库可保存 objectKey 或可解析为 objectKey 的 OSS/CDN URL。读接口统一转换为公网访问 URL；私有下载仍以预签名 URL 为准。

## 七、CDN 分阶段策略

| 阶段 | `cdn_domain` | 行为 |
|---|---|---|
| CDN 未启用 | `""` | 读接口走 OSS HTTPS 直链 |
| CDN 已配置未验收 | `""` | 保持直链，避免影响线上 |
| CDN 验收通过 | `"https://cdn.vunbo.com"` | 读接口拼 CDN URL，库内数据不变 |

CDN 启用步骤：

1. CDN 控制台添加 `cdn.vunbo.com`，源站选择 OSS Bucket。
2. DNS 添加 `cdn` CNAME 指向 CDN 分配域名。
3. CDN 控制台绑定 HTTPS 证书。
4. OSS CORS 增加 `https://admin.vunbo.com`。
5. 微信公众平台配置 `cdn.vunbo.com` 为 downloadFile/uploadFile 合法域名。
6. 生产 `config.yaml` 设置 `cdn_domain` 并重启后端。

## 八、验收步骤

### 1. STS

```bash
curl -s -X POST http://127.0.0.1:8080/admin/api/v1/upload/sts \
  -H "Authorization: Bearer <admin_jwt>" \
  -H "Content-Type: application/json"
```

期望：返回 `mode=oss`，并含 `accessKeyId`、`securityToken`、`bucket`、`dir`。

### 2. Admin 上传

1. 启动 Admin：`cd Admin && pnpm dev`
2. 登录后台，上传封面、正文图或下载资料。
3. 表单中保存的地址应为 OSS/CDN HTTPS 地址或 objectKey。

### 3. 预签名下载

1. 在后台创建上架的免费资料：`accessType=free`，`tags` 含「免费」。
2. 调用：

```bash
curl -s "http://127.0.0.1:8080/api/v1/download-files/<id>/sign-url"
```

期望：`data.url` 含 OSS 签名参数，可打开下载；去掉签名参数访问私有 Bucket 应返回 403。

### 4. 小程序真机

真机 `wx.downloadFile` 必须在微信公众平台配置合法域名：

- CDN 已启用：`cdn.vunbo.com`
- CDN 未启用：对应 Bucket HTTPS 域名

## 九、头像上传技术债

一期我的 Tab 不展示头像，但后端仍保留 `POST /api/v1/user/avatar` 和 `/uploads` 本地静态目录，用于兼容后续个人资料页。

已知限制：

- 本地 HTTP 头像真机无法作为合法域名资源展示。
- 文件落在应用服务器本地磁盘，不适合多实例。
- 历史头像不自动清理。

二期若恢复个人资料页，建议将头像上传纳入 OSS 统一链路，`users.avatar` 存 objectKey 或 CDN URL，并按环境决定是否保留本地上传兼容。
