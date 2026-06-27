# 部署 / CDN / 环境隔离实施计划

> **状态**：代码改造进行中（2026-06-28）。  
> **关联**：[`domain-vunbo.md`](./domain-vunbo.md)、[`oss-setup-guide.md`](./oss-setup-guide.md)、[008 开发计划 §五](../../prd/008-开发计划-第五阶段：联调部署与上线（Day27-28）.md)

---

## 一、背景与目标

| 现状 | 目标 |
|------|------|
| 本地开发，无生产存量数据 | 一次性定好 OSS 存储与 CDN 策略，避免上线后返工 |
| OSS 上传返回完整 HTTPS URL 写入库 | **库内只存 objectKey**，读时按当前 `cdn_domain` 拼 URL |
| 各端环境配置分散 | **dev / prod 明确分离**，部署时不手改业务代码 |
| CDN 尚未购买 | 域名先定为 `cdn.vunbo.com`，配置等 CDN 真通再填 |

---

## 二、域名规划（vunbo.com）

| 用途 | 域名 | 说明 |
|------|------|------|
| 小程序 / 公开 API | `api.vunbo.com` | 小程序 `BASE_URL` = `https://api.vunbo.com/api/v1` |
| 后台 Admin | `admin.vunbo.com` | 静态站点 + Nginx 反代 `/admin/api` |
| 对象存储 / CDN | `cdn.vunbo.com` | OSS CDN 自定义域名；微信 downloadFile 合法域名 |

根域名 `vunbo.com` 本期不承载业务。

---

## 三、OSS / CDN 策略

### 3.1 数据流（定稿）

```
Admin 上传 OSS
    → 表单 / API 提交 objectKey（如 essay-mvp/20260628/covers/xxx.jpg）
    → 写入 DB（cover_img / file_url / icon 等字段）
    → 读接口：PublicURL(objectKey) → 对外 HTTPS URL

cdn_domain 为空 → https://{bucket}.{endpoint}/{key}   （开发 / 未绑 CDN）
cdn_domain 已填 → https://cdn.vunbo.com/{key}           （生产 CDN）
```

### 3.2 CDN 分阶段

| 阶段 | `config.yaml` `cdn_domain` | 行为 |
|------|---------------------------|------|
| **现在（未买 CDN）** | 留空 `""` | 读写正常，走 OSS 直链 |
| **服务跑 1～2 周后** | 控制台开 CDN、绑 `cdn.vunbo.com`、DNS CNAME | 运维操作 |
| **CDN 生效** | `"https://cdn.vunbo.com"` + 重启后端 | **只改一行配置**，库内数据不动 |

### 3.3 私有下载

`SignDownloadURL` 已从 `file_url`（key 或历史 URL）解析 objectKey 再预签名，与 CDN 域名无关。

---

## 四、各端 dev / prod 环境隔离

### 4.1 后端 `course_server`

| 项 | 开发（本机） | 生产（云服务器） |
|----|-------------|-----------------|
| 配置 | `config.yaml`（`.gitignore`） | 服务器单独一份，**不进 Git** |
| 模板 | `config.yaml.example` | `config.prod.yaml.example` |
| 监听 | `:8080` | `127.0.0.1:8080`，Nginx 443 反代 |
| MySQL / Redis | 本机 | 同机或云 RDS |
| JWT | 开发密钥 | **随机强密钥** |
| OSS `cdn_domain` | 留空 | CDN 就绪后填写 |

### 4.2 Admin（Vite）

| 文件 | 用途 |
|------|------|
| `.env.development` | `VITE_API_BASE=/admin/api/v1` + Vite 代理 → `localhost:8080` |
| `.env.production` | 同路径，由 `admin.vunbo.com` Nginx 反代 |

开发：`pnpm dev`；上线：`pnpm build` 上传 `dist/`。

**本地 vs 生产启动方式（勿混用）**

| 组件 | 本地开发 | 生产服务器 |
|------|----------|------------|
| Go 后端 | `air`（热重载） | `go build -o essay-mvp .` + **systemd** |
| Admin | `pnpm dev`（:5173） | **`pnpm build`** → Nginx 托管 `dist/` |
| 对外 | 直连 localhost | 仅 **Nginx :443** |

可复制配置见 [`course_server/deploy/`](../deploy/)（`essay-mvp.service`、`nginx-vunbo.conf`）。

### 4.3 小程序

| 文件 | 用途 |
|------|------|
| `config/env.js` | `ENV` 开关：`development` / `production` |
| `config/env.example.js` | 模板（不含真实 IP） |

上线前：

1. `ENV = "production"`
2. `project.config.json` → `"urlCheck": true`
3. 微信公众平台配置合法域名

---

## 五、部署清单（P0）

### 5.1 现在已完成 / 进行中（代码）

- [x] 库内 OSS 媒体字段存 objectKey（写入规范化）
- [x] 读接口 `ResolvePublicURL` 统一拼 URL
- [x] Admin 上传回调存 `objectKey`
- [x] 小程序 `env.js` dev/prod 开关
- [x] `config.prod.yaml.example` 生产配置模板

### 5.2 部署阶段（运维）

- [ ] DNS：`api` / `admin` → 云服务器 IP
- [ ] 云服务器：MySQL + Redis + Go 后端 + Nginx
- [ ] TLS：`api.vunbo.com`、`admin.vunbo.com`（Let's Encrypt 或云证书）
- [ ] 生产 `config.yaml`（JWT、DSN、微信、OSS）
- [ ] Admin `pnpm build` → `/var/www/admin`
- [ ] 小程序 `ENV=production` + 上传审核

### 5.3 CDN 就绪后（约 5 分钟）

- [ ] 阿里云 CDN 加速 OSS，绑 `cdn.vunbo.com`
- [ ] DNS：`cdn` CNAME → CDN 分配域名
- [ ] CDN 控制台 HTTPS 证书
- [ ] 生产 `cdn_domain: "https://cdn.vunbo.com"` + 重启
- [ ] OSS CORS 加 `https://admin.vunbo.com`
- [ ] 微信 downloadFile 合法域名加 `cdn.vunbo.com`

---

## 六、服务器规格

| 规格 | 说明 |
|------|------|
| **2C2G** | MVP 可跑（Go + MySQL + Redis + Nginx 同机） |
| **2C4G** | Admin 构建、MySQL 占内存偏高时升级 |

进程建议绑定 `127.0.0.1:8080`，仅 Nginx 对外 443。

---

## 七、生产部署步骤（copy-paste）

### 7.1 目录约定（可按需改路径）

```text
/opt/essay-mvp/course_server/   # Go 源码 + essay-mvp 二进制 + config.yaml
/var/www/admin/                 # Admin pnpm build 的 dist 内容
/etc/systemd/system/essay-mvp.service
/etc/nginx/sites-enabled/essay-mvp   # 链到 deploy/nginx-vunbo.conf
```

### 7.2 后端（不用 air）

```bash
# 云服务器上
cd /opt/essay-mvp/course_server
cp config.prod.yaml.example config.yaml   # 编辑 dsn、jwt、oss、wechat
go build -o essay-mvp .

# 试跑（Ctrl+C 停掉后再走 systemd）
./essay-mvp
curl -s http://127.0.0.1:8080/health

sudo cp deploy/essay-mvp.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now essay-mvp
sudo systemctl status essay-mvp
```

改代码上线：`git pull && go build -o essay-mvp . && sudo systemctl restart essay-mvp`

### 7.3 Admin（不用 pnpm dev）

```bash
# 可在本机或服务器构建
cd Admin
pnpm install
pnpm build

# 上传到服务器（示例）
rsync -avz dist/ user@你的IP:/var/www/admin/
```

`.env.production` 保持 `VITE_API_BASE=/admin/api/v1`（由 Nginx 反代，无需改域名）。

### 7.4 Nginx + HTTPS

```bash
sudo apt install nginx certbot python3-certbot-nginx   # Debian/Ubuntu
sudo certbot certonly --nginx -d api.vunbo.com
sudo certbot certonly --nginx -d admin.vunbo.com

sudo cp deploy/nginx-vunbo.conf /etc/nginx/sites-available/essay-mvp
sudo ln -sf /etc/nginx/sites-available/essay-mvp /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

证书自动续期：`certbot renew`（certbot 安装后通常已有 timer）。

### 7.5 Nginx 配置摘要

完整文件：[`deploy/nginx-vunbo.conf`](../deploy/nginx-vunbo.conf)

```nginx
# api.vunbo.com → 127.0.0.1:8080
location /api/v1/ { proxy_pass http://127.0.0.1:8080/api/v1/; }
location /health  { proxy_pass http://127.0.0.1:8080/health; }

# admin.vunbo.com → 静态 + 反代
root /var/www/admin;
location /admin/api/ { proxy_pass http://127.0.0.1:8080/admin/api/; }
try_files $uri $uri/ /index.html;
```

### 7.6 systemd 单元摘要

完整文件：[`deploy/essay-mvp.service`](../deploy/essay-mvp.service)

```ini
WorkingDirectory=/opt/essay-mvp/course_server
ExecStart=/opt/essay-mvp/course_server/essay-mvp
Restart=always
Environment=GIN_MODE=release
```

日志：`journalctl -u essay-mvp -f`

---

## 八、验收 URL

- 健康检查：`https://api.vunbo.com/health`
- 后台：`https://admin.vunbo.com`
- CDN 资料示例：`https://cdn.vunbo.com/essay-mvp/20260628/downloads/xxx.pdf`

---

## 九、相关代码与部署文件

```
course_server/deploy/essay-mvp.service      # systemd（生产后端）
course_server/deploy/nginx-vunbo.conf     # Nginx api + admin
course_server/config.prod.yaml.example    # 生产 config 模板
course_server/internal/pkg/ossutil/       # OSS URL 读写
Admin/src/utils/ossUpload.ts
miniApp/config/env.js
```
