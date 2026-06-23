# OSS 上传统一改造待办

> **状态**：一期不处理，待 OSS 直传/转存方案落地时一并改造。  
> **关联**：Admin 上传占位见 [Admin/README.md](../../Admin/README.md#上传方案重要)；部署阶段见 PRD Day29 OSS/CDN。

---

## 1. 小程序用户头像（本地上传占位）

### 现状

| 环节 | 实现 |
|------|------|
| 上传 | `POST /api/v1/user/avatar`（需用户 JWT） |
| 存储 | 落盘 `course_server/uploads/avatars/{userId}_{timestamp}.jpg\|png` |
| 访问 | `GET /uploads/avatars/...`（`router.go` 中 `r.Static("/uploads", "uploads")`） |
| 资料库 | `users.avatar` 存相对路径（如 `/uploads/avatars/xxx.jpg`）或完整 URL |
| 小程序 | `miniApp/utils/upload.js` 上传；`utils/media.js` 可选 `downloadFile` 缓存后展示 |

### 已知问题（一期接受，不单独修）

1. **真机预览无法显示头像**  
   微信要求网络图片走 **HTTPS + downloadFile 合法域名**；局域网 `http://192.168.x.x` 无法配置进公众平台。手机浏览器可访问不代表小程序可显示。模拟器联调需勾选「不校验合法域名…」。

2. **与应用服务器耦合**  
   文件在应用机本地磁盘，容器重建/多实例部署时不易共享，无 CDN。

3. **无旧文件清理**  
   用户更换头像会生成新文件，历史 `uploads/avatars/*` 不自动删除。

4. **静态目录公开读**  
   知道 URL 即可访问（头像场景一般可接受）。

### OSS 改造时建议一并完成

- [ ] 头像上传改为 **OSS 直传（STS）** 或 **后端接收后转存 OSS**（与后台封面、资料文件统一链路）
- [ ] `users.avatar` 统一存 **OSS 公网 HTTPS URL** 或 **objectKey**（展示层拼 CDN 域名）
- [ ] 微信公众平台配置 **downloadFile 合法域名**（OSS/CDN 域名）
- [ ] 更换头像时 **删除或覆盖旧 object**（可选）
- [ ] 下线或保留 `POST /user/avatar` + `r.Static("/uploads")` 仅作本地开发兼容（按环境开关）

### 相关代码

```
course_server/internal/handler/user_profile.go   # UploadUserAvatar、UpdateUserProfile
course_server/internal/router/router.go          # Static /uploads、POST /user/avatar
miniApp/utils/upload.js
miniApp/utils/media.js
miniApp/pages/mine/index.js          # 我的 Tab（无 profile UI；不调用 GET /mine）
miniApp/README.md                  # 我的 Tab 一期定稿说明
```

---

## 2. 其他上传（同期纳入 OSS 改造）

后台与内容侧当前多为 **URL 录入** 或 **STS 占位**，OSS 统一改造时建议同一批次处理：

- 文章封面、正文长图
- 首页入口图标、广告位素材
- 公众号二维码
- 下载资料文件（签名 URL，见 `handler/download.go`）

具体范围以 OSS 方案评审为准；本文档随改造推进更新勾选状态。
