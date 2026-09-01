# API 参考文档

> 路由唯一事实来源：[`internal/router/router.go`](../internal/router/router.go)。
> 当前文档覆盖：健康检查 2 个 HTTP 方法、后台 Admin 46 个接口、小程序 28 个接口。

## 一、通用约定

### 1.1 路径前缀

| 类型 | 前缀 |
|---|---|
| 健康检查 | `/health` |
| 后台管理 | `/admin/api/v1` |
| 微信小程序 | `/api/v1` |

### 1.2 统一响应

```json
{ "code": 0, "message": "success", "data": {} }
```

- 成功：`code = 0`，HTTP 200。
- 失败：HTTP 状态码与 `code` 一致，常见为 400、401、403、404、500。

分页响应位于 `data` 内：

```json
{ "list": [], "total": 0, "page": 1, "pageSize": 20 }
```

### 1.3 鉴权

请求头：

```http
Authorization: Bearer <token>
```

| 鉴权 | 说明 | JWT `type` |
|---|---|---|
| 无 | 公开接口 | 无 |
| Admin | 后台除登录外全部接口 | `admin` |
| User | 小程序用户中心接口 | `user` |
| OptionalUser | 无 token 也可访问，有 token 则记录用户 ID | `user` |

## 二、健康检查

| Method | Path | 鉴权 | 说明 |
|---|---|---|---|
| GET | `/health` | 无 | 存活探测 |
| HEAD | `/health` | 无 | 负载均衡或 `curl -I` 探测 |

响应：

```json
{ "status": "ok" }
```

## 三、后台 Admin API

后台统一前缀：`/admin/api/v1`。除登录外均需 Admin 鉴权。

### 3.1 认证

| Method | Path | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/auth/login` | 无 | 管理员登录 |

请求示例：

```json
{ "username": "admin", "password": "admin123" }
```

成功返回：`token`、`adminId`、`username`、`role`。

### 3.2 统计与审计

| Method | Path | 说明 |
|---|---|---|
| GET | `/stats/overview` | 今日/累计用户、PV、UV、文章、下载等概览 |
| GET | `/stats/hot-articles` | 热门文章 Top 10 |
| GET | `/stats/hot-searches` | 热门搜索词 |
| GET | `/stats/hot-downloads` | 免费资料下载排行 |
| GET | `/stats/ad-events` | 广告曝光、点击、激励完成统计 |
| GET | `/operation-logs` | 后台写操作审计日志 |

### 3.3 意见反馈

| Method | Path | 说明 |
|---|---|---|
| GET | `/feedbacks` | 分页查询用户反馈 |
| PUT | `/feedbacks/:id/reply` | 管理员回复反馈 |

回复请求：

```json
{ "reply": "感谢反馈，已记录。" }
```

### 3.4 教材分类单元

| Method | Path | 说明 |
|---|---|---|
| GET | `/taxonomy-units` | 分类列表；支持分页和筛选 |
| POST | `/taxonomy-units` | 创建分类 |
| PUT | `/taxonomy-units/:id` | 更新分类 |
| DELETE | `/taxonomy-units/:id` | 删除分类；须已禁用且无文章占用同一四元组 |

创建/更新关键字段：

```json
{
  "textbookVersion": "统编版",
  "grade": 3,
  "semester": "上册",
  "unit": 1,
  "unitTitle": "第一单元",
  "sortOrder": 0,
  "status": 1
}
```

### 3.5 标签

| Method | Path | 说明 |
|---|---|---|
| GET | `/tags` | 标签列表 |
| POST | `/tags` | 创建标签 |
| PUT | `/tags/:id` | 更新标签 |
| DELETE | `/tags/:id` | 删除标签；须已禁用且无内容关联 |

请求示例：

```json
{ "name": "免费", "sortOrder": 0, "status": 1 }
```

### 3.6 作文专题文章

| Method | Path | 说明 |
|---|---|---|
| GET | `/articles` | 文章列表；省略正文长图 |
| POST | `/articles` | 创建文章 |
| GET | `/articles/:id` | 文章详情；含标签和正文长图 |
| PUT | `/articles/:id` | 更新文章 |
| PUT | `/articles/:id/status` | 上下架 |

创建/更新关键字段：`title`、`slug`、`grade`、`semester`、`unit`、`essayTopic`、`contentType`、`summary`、`coverImg`、`bodyImages`、`tagIds`、`viewAccessType`、`isRecommended`、`isHot`。

### 3.7 优秀作文

| Method | Path | 说明 |
|---|---|---|
| GET | `/essay-examples` | 优秀作文列表 |
| POST | `/essay-examples` | 创建优秀作文 |
| GET | `/essay-examples/:id` | 优秀作文详情 |
| PUT | `/essay-examples/:id` | 更新优秀作文 |
| DELETE | `/essay-examples/:id` | 删除优秀作文 |

关键字段：`title`、`grade`、`body`、`comment`、`articleId`、`tagIds`。`wordCount` 由服务端统计。

### 3.8 下载资料

| Method | Path | 说明 |
|---|---|---|
| GET | `/download-files` | 下载资料列表 |
| POST | `/download-files` | 创建下载资料 |
| PUT | `/download-files/:id` | 更新下载资料 |
| DELETE | `/download-files/:id` | 删除下载资料；有下载记录时需 `?confirm=true` |

关键字段：`ownerType`、`ownerId`、`fileName`、`accessType`、`fileUrl`、`tags`、`fileType`、`fileSize`、`needLogin`、`sortOrder`、`status`。

一期公开下载要求：`status=1` + `accessType=free` + `tags` 含「免费」。

### 3.9 首页入口

| Method | Path | 说明 |
|---|---|---|
| GET | `/home-blocks` | 首页入口列表 |
| POST | `/home-blocks` | 创建首页入口 |
| PUT | `/home-blocks/:id` | 更新首页入口 |
| DELETE | `/home-blocks/:id` | 删除首页入口；须已禁用 |

请求示例：

```json
{
  "title": "写作素材库",
  "icon": "https://cdn.vunbo.com/icons/material.png",
  "navType": "navigateTo",
  "path": "/pages/material/index",
  "sortOrder": 40,
  "status": 1
}
```

### 3.10 搜索热门词

| Method | Path | 说明 |
|---|---|---|
| GET | `/hot-search-keywords` | 搜索热门词列表（含禁用项） |
| POST | `/hot-search-keywords` | 创建搜索热门词；`keyword` 唯一 |
| PUT | `/hot-search-keywords/:id` | 更新搜索热门词 |
| DELETE | `/hot-search-keywords/:id` | 删除搜索热门词；须已禁用 |

请求示例：

```json
{
  "keyword": "写人作文",
  "sortOrder": 10,
  "status": 1
}
```

与 Dashboard `GET /stats/hot-searches`（`search_logs` 聚合）无关。

### 3.11 广告位

| Method | Path | 说明 |
|---|---|---|
| GET | `/ad-slots` | 广告位列表 |
| POST | `/ad-slots` | 按 `slotKey` 创建或更新广告位 |
| DELETE | `/ad-slots/:id` | 删除广告位；须已禁用 |

关键字段：`slotKey`、`slotName`、`page`、`position`、`adType`、`adUnitId`、`sortOrder`、`status`。

下载激励视频固定使用：

```json
{
  "slotKey": "download_rewarded_video",
  "page": "download",
  "adType": "rewarded_video"
}
```

### 3.12 公众号导流

| Method | Path | 说明 |
|---|---|---|
| GET | `/official-account-configs` | 公众号导流配置列表 |
| POST | `/official-account-configs` | 按 `scene` 创建或更新配置 |
| DELETE | `/official-account-configs/:id` | 删除配置；须已禁用 |

### 3.13 上传凭证

| Method | Path | 说明 |
|---|---|---|
| POST | `/upload/sts` | 返回 OSS STS 临时凭证 |

配置完整时返回 `mode=oss`；OSS 配置缺失时返回占位信息，便于本地开发降级。

## 四、小程序 API

小程序统一前缀：`/api/v1`。

### 4.1 登录

| Method | Path | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/wx/login` | 无 | 微信登录；生产配置后走 code2session |

请求示例：

```json
{ "code": "wx-login-code", "sourceScene": "mine" }
```

成功返回：`token`、`user`。

### 4.2 公开内容

| Method | Path | 鉴权 | 说明 |
|---|---|---|---|
| GET | `/home` | 无 | 首页聚合：推荐、热门、优秀作文、首页入口 |
| GET | `/categories` | 无 | 启用中的分类单元扁平列表 |
| GET | `/articles` | 无 | 已发布文章列表 |
| GET | `/articles/:id` | 无 | 已发布文章详情；递增阅读数 |
| GET | `/essays` | 无 | 优秀作文列表 |
| GET | `/essays/:id` | 无 | 优秀作文详情 |
| GET | `/search` | 无 | 搜索；`keyword` 必填 |
| GET | `/hot-search-keywords` | 无 | 启用中的搜索热门词，返回 `string[]` |
| GET | `/ad-slots` | 无 | 启用广告位；可选 `?page=` |
| GET | `/official-account-configs` | 无 | 启用的公众号导流配置；须传 `?scene=` |

### 4.3 下载

以下接口为 OptionalUser：无 token 可访问，有 user token 时记录用户 ID。

| Method | Path | 说明 |
|---|---|---|
| GET | `/download-files/:id/sign-url` | 获取资料预签名下载 URL |
| POST | `/download-files/:id/ad-unlock` | 激励视频完成后换取 `unlockToken` |

`sign-url` Query：

- `unlockToken`：下载激励视频门槛启用时必填。
- `sourceScene`：来源场景。

成功返回：`url`、`expiresIn`、`file`。

权限边界见 [download-permission-boundary.md](./download-permission-boundary.md)。

### 4.4 埋点

以下接口为 OptionalUser，`userId` 仅由服务端从 JWT 解析，未登录记为 0。

| Method | Path | 说明 |
|---|---|---|
| POST | `/track/view` | 浏览埋点 |
| POST | `/track/share` | 分享埋点；文章分享递增 `share_count` |
| POST | `/track/search` | 搜索埋点 |
| POST | `/track/ad` | 广告曝光/点击埋点；激励完成走 `ad-unlock` |

### 4.5 用户中心

以下接口需要 User 鉴权。

| Method | Path | 说明 |
|---|---|---|
| GET | `/mine` | 我的概览；一期我的 Tab 不调用但接口保留 |
| GET | `/favorites` | 收藏列表 |
| POST | `/favorites` | 添加收藏 |
| DELETE | `/favorites` | 取消收藏 |
| GET | `/history` | 浏览历史 |
| GET | `/downloads` | 下载记录 |
| GET | `/user/profile` | 用户资料详情 |
| PUT | `/user/profile` | 更新用户资料 |
| POST | `/user/avatar` | 上传用户头像；一期仅本地兼容 |
| POST | `/feedbacks` | 提交意见反馈 |
| GET | `/feedbacks` | 当前用户反馈列表 |

收藏请求示例：

```json
{ "targetType": "article", "targetId": 1 }
```

反馈请求示例：

```json
{ "content": "希望增加四年级下册范文" }
```

## 五、一期不存在的接口

以下接口一期未实现，客户端不得调用：

| 路径模式 | 说明 |
|---|---|
| `/admin/api/v1/orders/*` | 订单管理 |
| `/admin/api/v1/products/*` | 商品管理 |
| `/admin/api/v1/pay/*` | 支付管理 |
| `/api/v1/orders/*` | 小程序订单 |
| `/api/v1/pay/*` | 小程序支付 |
| `/api/v1/courses/*` | 旧课件方案 |
| 视频号/讲解视频相关 | 一期禁止 |

二期支付、订单、权益规划见 [../../prd/004-二期支付订单登录规划.md](../../prd/004-二期支付订单登录规划.md)。

## 六、关联文档

- [api-self-test-checklist.md](./api-self-test-checklist.md)
- [download-permission-boundary.md](./download-permission-boundary.md)
- [postman/essay-mvp-api.postman_collection.json](./postman/essay-mvp-api.postman_collection.json)
