# 阿里云 OSS 接入指南（P0-①）

> **适用阶段**：第五阶段 P0-① 环境验收；代码已交付，按本文在阿里云控制台完成资源与 `config.yaml` 配置后即可联调。  
> **生产域名**：**vunbo.com**（CDN `cdn.vunbo.com`、Admin `admin.vunbo.com`），详见 [`domain-vunbo.md`](./domain-vunbo.md)。  
> **关联**：[008 开发计划 §三](../../prd/008-开发计划-第五阶段：联调部署与上线（Day27-28）.md)、[`config.yaml.example`](../config.yaml.example)

---

## 一、你需要准备的信息

完成下列步骤后，应能填齐 `config.yaml` 的 `oss` 段（**勿提交 Git**）：

| 字段 | 来源 | 示例 |
|------|------|------|
| `endpoint` | Bucket 概览 → 外网 Endpoint | `oss-cn-hangzhou.aliyuncs.com` |
| `bucket` | Bucket 名称 | `essay-mvp-prod` |
| `access_key_id` | RAM 用户 AccessKey | `LTAI5t...` |
| `access_key_secret` | RAM 用户 Secret | （仅本地保存） |
| `sts_role_arn` | RAM 角色详情 | `acs:ram::1234567890123456:role/essay-mvp-oss-upload` |
| `cdn_domain` | CDN 自定义域名 | `https://cdn.vunbo.com` |
| `upload_dir` | 项目默认即可 | `essay-mvp` |

**能力分工**：

- **主 AK/SK**（RAM 用户）：后端 `SignDownloadURL` 预签名下载、`AssumeRole` 签发 STS
- **STS 临时凭证**：Admin 浏览器直传 `PutObject`（权限最小化，仅当日目录）

---

## 二、Bucket 配置

### 1. 创建 Bucket（若已购买可跳过）

1. 登录 [OSS 控制台](https://oss.console.aliyun.com/)
2. 创建 Bucket，建议：
   - **地域**：与后端服务器同区域（降低延迟与出网费）
   - **存储类型**：标准存储
   - **读写权限**：**私有**（通过预签名 URL 下载，符合一期安全要求）
   - **版本控制**：关闭（一期不需要）

### 2. 跨域 CORS（Admin 浏览器直传必配）

Bucket → **数据安全** → **跨域设置** → 创建规则：

| 项 | 开发联调 | 生产 |
|----|----------|------|
| 来源 Origin | `http://localhost:5173` | `https://admin.vunbo.com` |
| Methods | `GET, PUT, POST, HEAD` | 同上 |
| Allow-Headers | `*` | `*` |
| Expose-Headers | `ETag, x-oss-request-id` | 同上 |

> 本地 Admin 默认 Vite 端口 `5173`；若端口不同请改成实际 Origin。生产 Origin 在 P0-③ 部署 Admin 后再补一条。

---

## 三、RAM 用户与 STS 角色

### 架构

```text
course_server（主 AK/SK）
  ├─ sts:AssumeRole → RAM 角色 essay-mvp-oss-upload
  │                     └─ 返回临时凭证 → Admin ali-oss 直传 PutObject
  └─ oss:GetObject   → SignDownloadURL 预签名下载
```

### Step 1 — 创建 RAM 用户 `essay-mvp-api`

[RAM 控制台 → 用户](https://ram.console.aliyun.com/users) → 创建用户：

- 登录名称：`essay-mvp-api`
- 访问方式：勾选 **「使用永久 AccessKey 访问」**（旧版控制台称 OpenAPI 调用访问；**不要**勾控制台访问）
- 创建后保存 **AccessKey ID / Secret** → 写入 `config.yaml` 的 `access_key_id` / `access_key_secret`

### Step 2 — 创建 RAM 角色 `essay-mvp-oss-upload`

[RAM 控制台 → 角色](https://ram.console.aliyun.com/roles) → 创建角色：

- 信任主体：**阿里云账号**
- 信任策略：允许本账号下 RAM 用户 `essay-mvp-api` 扮演（见下方 JSON）
- 复制角色 ARN → 写入 `config.yaml` 的 `sts_role_arn`

**信任策略示例**（将 `<主账号ID>`、`essay-mvp-api` 换成实际值）：

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

**角色权限策略**（将 `<bucket>` 换成 Bucket 名；与代码 `upload_dir` 一致）：

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

> 后端 `IssueSTSCredentials` 还会在 AssumeRole 时附加**会话策略**，进一步限制为**当日目录** `essay-mvp/YYYYMMDD/*`，双重收窄上传范围。

**策略绑定位置（易错）**：

| 策略 | 授权对象 |
|------|----------|
| `essay-mvp-oss-upload-policy`（`oss:PutObject`） | **角色** `essay-mvp-oss-upload` |
| 含 `oss:GetObject` + `sts:AssumeRole` 的用户策略 | **用户** `essay-mvp-api` |

**控制台可视化编辑踩坑**：弹窗填 `cn-guangzhou` 后，摘要可能自动变成 `oss-cn-guangzhou` 并显示「未识别资源」。**建议用脚本编辑**保存下列 JSON（地域用 `*` 即可），保存后勿再切回可视化：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["oss:PutObject"],
      "Resource": ["acs:oss:*:*:essay-mvp/*"]
    }
  ]
}
```

> 角色策略用 `essay-mvp/*` 覆盖 Bucket 内所有 object（含 `essay-mvp/20260628/...`）；当日目录收窄由后端 STS **会话策略**负责，无需在角色策略里写日期。

**PUT 403 + `AccessDenied` + `AssumedRoleUser`（STS 已成功）**：

XML 含 `AuthAction: oss:PutObject`、`ImplicitDeny`、`essay-mvp-oss-upload:essay-mvp-admin` 时，表示**扮演角色成功，但角色上没有生效的 PutObject 权限**（不是 CORS、不是 region）。

| 检查项 | 正确状态 |
|--------|----------|
| `essay-mvp-oss-upload-policy` 授权对象 | **角色** `essay-mvp-oss-upload`（❌ 不能只在用户上） |
| 角色 → 权限管理 | 列表中有 `essay-mvp-oss-upload-policy`，且策略脚本非空 |
| 策略 Resource | `acs:oss:*:*:essay-mvp/*`（脚本编辑保存，摘要黄标可忽略） |
| 用户上的 PutObject | **不应**有；用户只需 GetObject + AssumeRole |

**快速验证**：角色 → 新增授权 → 临时挂系统策略 **AliyunOSSFullAccess** → 再上传；若成功则一定是自定义角色策略未生效，按上表重建并绑角色后去掉 FullAccess。

> RAM 策略 regionId 写 `cn-guangzhou`；`config.yaml` 的 `endpoint` 才写 `oss-cn-guangzhou.aliyuncs.com`。二者勿混用。

### Step 3 — 给 RAM 用户授权

为 `essay-mvp-api` 添加自定义策略（将 `<bucket>`、`<role>` 换成实际值）：

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

---

## 四、填写 config.yaml

复制 [`config.yaml.example`](../config.yaml.example) 为 `config.yaml`（已在 `.gitignore`），填入 OSS 段：

```yaml
oss:
  endpoint: "oss-cn-guangzhou.aliyuncs.com"   # 勿带 https://；与 Bucket 地域一致
  bucket: "your-bucket-name"
  access_key_id: "LTAI..."
  access_key_secret: "..."
  cdn_domain: "https://cdn.vunbo.com"         # 绑 CDN 后填写；未绑可留空
  sts_role_arn: "acs:ram::1234567890123456:role/essay-mvp-oss-upload"
  upload_dir: "essay-mvp"
```

重启 `course_server`（或等待 air 热重载）使配置生效。

---

## 五、验收步骤

### 1. STS 接口（Postman / curl）

Admin 登录拿 JWT 后：

```bash
curl -s -X POST http://127.0.0.1:8080/admin/api/v1/upload/sts \
  -H "Authorization: Bearer <admin_jwt>" \
  -H "Content-Type: application/json" | jq .
```

**期望**：`mode` 为 `"oss"`，含 `accessKeyId`、`securityToken`、`bucket`、`dir`（形如 `essay-mvp/20260628/`）。

**失败排查**：

| 现象 | 可能原因 |
|------|----------|
| `mode: placeholder` | `sts_role_arn` 未填或主 AK/SK/bucket/endpoint 不完整 |
| `failed to issue oss sts credentials` | RAM 用户无 `sts:AssumeRole`、角色信任策略未允许该用户、或角色名 ARN 写错 |
| Admin 上传 CORS 报错 | Bucket CORS 未配或 Origin 不匹配 |

### 2. Admin 上传

1. 启动 Admin：`cd Admin && pnpm dev`
2. 登录后台 → 下载资料 / 文章封面 → 点击「上传」
3. **期望**：表单 `fileUrl` / `coverUrl` 为 `https://<bucket>.<endpoint>/essay-mvp/...` 或 CDN 地址

**PUT 403 排查**（STS 已成功但浏览器上传失败）：

| 现象 | 处理 |
|------|------|
| Network 里 Remote Address 为 `127.0.0.1:7890` 等代理 | Clash/系统代理添加绕过：`*.aliyuncs.com`，或上传时关代理 |
| Response XML 为 `AccessDenied` + `AssumedRoleUser` | **角色**未绑有效 PutObject 策略；见 §三 Step 2「PUT 403 + AccessDenied」 |
| Response XML 为 `AccessDenied`（其他） | 确认 `essay-mvp-oss-upload-policy` 绑在**角色**上（非用户） |
| Response XML 为 `SignatureDoesNotMatch` | 多为代理或 region 不对；STS 应返回 `region: oss-cn-guangzhou`（非 `cn-guangzhou`） |
| 改代码后仍失败 | 硬刷新 Admin（清 STS 缓存）并等 air 重载后端 |

### 3. 预签名下载

1. 在 Admin 创建一条**免费**下载资料（tags 含「免费」，`access_type=free`，`status=1`），`fileUrl` 填上一步 OSS URL
2. 调用：

```bash
curl -s "http://127.0.0.1:8080/api/v1/download-files/<id>/sign-url" | jq .
```

**期望**：`data.url` 含 OSS 签名参数（`Expires`、`Signature`、`OSSAccessKeyId` 等），非仅 `expires=` 伪签名。

3. 浏览器打开返回 URL → 应能下载/预览 PDF
4. 去掉 URL 查询参数直接访问原 `fileUrl` → 私有 Bucket 应 **403**

### 4. 小程序（依赖 P0-③ 合法域名）

真机 `wx.downloadFile` 须在微信公众平台配置 **downloadFile 合法域名**：

- 无 CDN：`your-bucket.oss-cn-<region>.aliyuncs.com`
- 有 CDN：`cdn.vunbo.com`

---

## 六、可选：CDN 自定义域名

一期可先用 OSS 虚拟主机 HTTPS URL 完成 Admin 上传与后端预签名；上线前建议绑 CDN：

1. [CDN 控制台](https://cdn.console.aliyun.com/) 添加加速域名 **`cdn.vunbo.com`**，源站选 OSS Bucket
2. 在 DNS 添加 `cdn` CNAME 指向 CDN 分配域名
3. 完成 HTTPS 证书绑定
4. `config.yaml` 填 `cdn_domain: "https://cdn.vunbo.com"`
5. Admin 上传与 `sign-url` 解析均优先走 CDN 域名（代码已支持）

---

## 七、objectKey 路径规范（与代码一致）

| 场景 | Admin `subdir` | 最终 key 示例 |
|------|----------------|---------------|
| 文章封面 | `covers` | `essay-mvp/20260628/covers/1730_abc.jpg` |
| 正文长图 | `body` | `essay-mvp/20260628/body/1730_def.jpg` |
| 下载资料 | `downloads` | `essay-mvp/20260628/downloads/1730.pdf` |
| 首页图标 | `icons` | `essay-mvp/20260628/icons/1730.png` |

---

## 八、勾选清单（P0-① 环境验收）

完成联调后在 [008 开发计划 §3.5](../../prd/008-开发计划-第五阶段：联调部署与上线（Day27-28）.md) 与 [api-self-test-checklist.md](./api-self-test-checklist.md) 打勾：

- [x] `POST /admin/api/v1/upload/sts` 返回真实 OSS 凭证（#36）— 2026-06-28 本地验收通过
- [ ] Admin 上传 PDF → `fileUrl` 为 HTTPS OSS/CDN 地址
- [ ] `GET sign-url` 返回带 OSS 签名参数的 URL（#46）
- [ ] 未签名直链访问私有 Bucket → 403
- [ ] （P0-③ 后）真机 `wx.downloadFile` 可打开 PDF
