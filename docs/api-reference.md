# API 参考文档（内容型作文 MVP 一期）

> 路由源码唯一入口：[`internal/router/router.go`](../internal/router/router.go)  
> 通用约定详见 [`README.md`](../README.md)。本文档覆盖 **65** 个 HTTP 端点。

---

## 一、通用约定

### 1.1 路径前缀

| 客户端 | 前缀 |
|--------|------|
| 后台管理 | `/admin/api/v1` |
| 微信小程序 | `/api/v1` |
| 健康检查 | `GET /health` |

### 1.2 统一响应

```json
{ "code": 0, "message": "success", "data": {} }
```

- 成功：`code = 0`，HTTP 200
- 错误：HTTP 状态码与 `code` 一致（400/401/403/404/500）

分页 `data` 结构：

```json
{ "list": [], "total": 0, "page": 1, "pageSize": 20 }
```

- `page` 默认 1，`pageSize` 默认 20，最大 100

### 1.3 鉴权

Header：`Authorization: Bearer <token>`（兼容裸 token）

| 类型 | 说明 | JWT `type` |
|------|------|------------|
| 无 | 公开接口 | — |
| Admin | 后台除 login 外全部 | `admin` |
| User | 我的/收藏/历史/下载记录/意见反馈 | `user` |
| OptionalUser | 下载签名、埋点；有 token 则解析 userId，无 token 仍放行 | 可选 `user` |

---

## 二、健康检查（1）

### GET `/health`

- **鉴权**：无
- **说明**：存活探测，不访问数据库
- **响应**：`{ "status": "ok" }`

---

## 三、后台 Admin API（41）

### 3.1 认证（1）

#### POST `/admin/api/v1/auth/login`

- **鉴权**：无
- **Body**：

```json
{ "username": "admin", "password": "admin123" }
```

- **成功 data**：`token`、`adminId`、`username`、`role`
- **错误**：401 用户名或密码错误

---

### 3.2 统计 Dashboard（5）

均需 **AdminAuth**。

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/stats/overview` | 今日/累计：用户、PV、UV、文章、下载、收藏等 |
| 2 | GET | `/stats/hot-articles` | 热门文章 Top 10（按 view_count） |
| 3 | GET | `/stats/hot-searches` | 热门搜索词 |
| 4 | GET | `/stats/hot-downloads` | 免费资料下载排行 |
| 5 | GET | `/stats/ad-events` | 广告曝光/点击/激励完成统计 |

---

### 3.3 操作日志（1）

#### GET `/admin/api/v1/operation-logs`

- **鉴权**：Admin
- **Query**：`page`、`pageSize`、`adminId`、`targetType`
- **说明**：后台 POST/PUT/DELETE 成功写操作自动记入（GET 不记）

---

### 3.4 教材分类单元（3）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/taxonomy-units` | 列表；Query：`grade`、`semester`、`status`、分页 |
| 2 | POST | `/taxonomy-units` | 创建 |
| 3 | PUT | `/taxonomy-units/:id` | 更新 |

**POST/PUT Body 示例**：

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

---

### 3.5 标签（4）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/tags` | 列表；Query：`keyword`、`status`、分页 |
| 2 | POST | `/tags` | 创建 |
| 3 | PUT | `/tags/:id` | 更新；启用→禁用时须无文章/作文关联 |
| 4 | DELETE | `/tags/:id` | 删除；须已禁用且无文章/作文关联 |

**POST/PUT Body**：`{ "name": "免费", "sortOrder": 0, "status": 1 }`

**禁用/删除约束**：

- 启用→禁用：若 `article_tags` 或 `essay_example_tags` 仍引用该标签，返回 400
- 物理删除：仅 `status=0` 且无上述关联时可删

---

### 3.6 作文专题 / 文章（5）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/articles` | 列表；Query：`grade`、`semester`、`unit`、`contentType`、`status`、`keyword`、`tagId`、分页；省略 `bodyImages`；每项含 `tags` 数组 |
| 2 | POST | `/articles` | 创建（含 `bodyImages`、`tagIds`） |
| 3 | GET | `/articles/:id` | 详情（含标签、正文图片） |
| 4 | PUT | `/articles/:id` | 更新 |
| 5 | PUT | `/articles/:id/status` | 上下架；Body：`{ "status": 0 \| 1 }`；发布时维护 `publishedAt` |

**创建 Body 关键字段**：`title`、`slug`、`grade`、`semester`、`unit`、`essayTopic`、`contentType`、`summary`、`coverImg`；可选 `bodyImages`（最多 20 张）、`tagIds`、`viewAccessType`（一期固定 `free`）、`isRecommended`、`isHot`

---

### 3.7 优秀作文（5）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/essay-examples` | Query：`grade`、`articleId`、`status`、`keyword`、`tagId`、分页；每项含 `tags` 数组 |
| 2 | POST | `/essay-examples` | 创建 |
| 3 | GET | `/essay-examples/:id` | 详情 |
| 4 | PUT | `/essay-examples/:id` | 更新 |
| 5 | DELETE | `/essay-examples/:id` | 删除 |

**Body 关键字段**：`title`、`grade`、`body`、`comment`；可选 `articleId`、`tagIds`。`wordCount` 由服务端根据 `body` 自动统计（不计空白符），无需传入

---

### 3.8 下载资料（4）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/download-files` | Query：`ownerType`、`ownerId`、`accessType`、分页 |
| 2 | POST | `/download-files` | 创建 |
| 3 | PUT | `/download-files/:id` | 更新 |
| 4 | DELETE | `/download-files/:id` | 删除；有下载记录时需 `?confirm=true` |

**Body 关键字段**：`ownerType`（一期仅 `article`）、`ownerId`、`fileName`、`accessType`、`fileUrl`；可选 `tags`（含「免费」为运营软开关）、`fileType`、`fileSize`、`needLogin`、`sortOrder`、`status`

---

### 3.9 运营配置（11）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/home-blocks` | 首页入口列表 |
| 2 | POST | `/home-blocks` | 创建首页入口 |
| 3 | PUT | `/home-blocks/:id` | 更新首页入口 |
| 4 | DELETE | `/home-blocks/:id` | 删除首页入口；须已禁用 |
| 5 | GET | `/ad-slots` | 广告位列表 |
| 6 | POST | `/ad-slots` | 按 `slotKey` Upsert 广告位 |
| 7 | DELETE | `/ad-slots/:id` | 删除广告位；须已禁用 |
| 8 | GET | `/official-account-configs` | 公众号导流配置列表 |
| 9 | POST | `/official-account-configs` | 按 `scene` Upsert 公众号配置 |
| 10 | DELETE | `/official-account-configs/:id` | 删除公众号导流配置；须已禁用 |
| 11 | POST | `/upload/sts` | OSS STS 凭证（开发占位） |

**首页入口 Body**：`title`、`icon`、`navType`（`navigateTo`/`switchTab`/`redirectTo`/`reLaunch`）、`path`（须以 `/` 开头）；可选 `sortOrder`、`status`

**首页入口 / 广告位 / 公众号导流删除约束**：

- 物理删除：仅 `status=0`（已禁用）时可删；启用中返回 400
- 首页入口：删除后不再出现在 `GET /api/v1/home` 的 `blocks` 中
- 广告位：PRD 预置 `slotKey`（如 `home_feed_mid`、`mine_top`）删除后，服务重启时 `SeedAdSlots` 会按缺失补回；预置位默认禁用且 `adUnitId` 为空

**广告位 Upsert Body**：`slotKey`、`slotName`、`page`、`position`、`adType`（`banner`/`rewarded_video`）、`adUnitId`、`sortOrder`、`status`

**下载激励视频**：`slotKey=download_rewarded_video`，`page=download`，`adType=rewarded_video`

---

### 3.10 意见反馈（2）

均需 **AdminAuth**。

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/feedbacks` | 分页列表；Query：`page`、`pageSize`、`status`（`pending`/`replied`）、`keyword`（匹配正文或回复） |
| 2 | PUT | `/feedbacks/:id/reply` | 管理员回复；回复后 `status=replied` |

**列表/详情项字段**（`FeedbackListItem`）：`id`、`userId`、`userNickname`、`content`、`status`、`reply`、`repliedAt`、`createdAt`

**回复 Body**：

```json
{ "reply": "感谢您的反馈，已记录。" }
```

- `reply` 必填；trim 后非空；最长 **1000** 字（按 rune 计）
- 成功返回更新后的反馈对象（含 `userNickname`）
- 404：反馈不存在

---

## 四、小程序 API（23）

### 4.1 登录（1）

#### POST `/api/v1/wx/login`

- **鉴权**：无
- **Body**：`{ "code": "任意字符串", "sourceScene": "", "nickname": "", "avatar": "" }`
- **说明**：开发态 `openid = "dev_" + code`；配置 wechat 后替换为 code2session
- **成功 data**：`token`、`user`

---

### 4.2 公开内容（8）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/home` | 推荐/热门文章、优秀作文、首页入口 |
| 2 | GET | `/categories` | 分类树 |
| 3 | GET | `/articles` | 已发布文章列表；Query：`grade`、`semester`、`keyword` 等 |
| 4 | GET | `/articles/:id` | 已发布详情；递增 `view_count`；返回可免费下载资料 + `downloadGate` |
| 5 | GET | `/essays` | 优秀作文列表 |
| 6 | GET | `/essays/:id` | 优秀作文详情；可带关联文章精简信息 |
| 7 | GET | `/search` | 全文搜索；Query：`keyword`（必填）、`page`、`pageSize`；搜索埋点见 `POST /track/search` |
| 8 | GET | `/ad-slots` | 启用广告位；Query：可选 `page`（`home` / `article_detail` / `essay_detail` / `search` / **`mine`**） |

---

### 4.3 下载（2，OptionalUser）

#### GET `/api/v1/download-files/:id/sign-url`

- **Query**：`unlockToken`（激励视频门槛启用时必填）、`sourceScene`
- **成功 data**：`url`、`expiresIn`（300）、`file`
- **权限**：须 `status=1` + `access_type=free` + `tags` 含「免费」；否则 403
- **副作用**：写 `download_logs`、递增 `download_count`

#### POST `/api/v1/download-files/:id/ad-unlock`

- **Query**：可选 `pagePath`
- **说明**：激励视频完成后签发 `unlockToken`（300 秒）
- **错误**：403 非免费资料；400 gate 未启用

详见 [`download-permission-boundary.md`](./download-permission-boundary.md)。

---

### 4.4 埋点（4，OptionalUser）

`userId` 仅服务端从 JWT 解析，**不接受 body 传入**；未登录为 0。

| # | Method | Path | Body 要点 |
|---|--------|------|-----------|
| 1 | POST | `/track/view` | `targetType`、`targetId`、`pagePath`、`sourceScene` |
| 2 | POST | `/track/share` | 同上 + `shareChannel`；文章分享递增 `share_count` |
| 3 | POST | `/track/search` | `keyword`（必填）、`resultCount`、`sourceScene` |
| 4 | `/track/ad` | `slotKey`、`eventType`（`expose`/`click`）；`rewarded_complete` 须走 ad-unlock |

---

### 4.5 用户中心（6，UserAuth）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | GET | `/mine` | 收藏/浏览/下载计数概览；**一期我的 Tab 不调用**，接口保留 |
| 2 | GET | `/favorites` | 我的收藏分页 |
| 3 | POST | `/favorites` | Body：`targetType`（`article`/`essay`）、`targetId`；去重 |
| 4 | DELETE | `/favorites` | Query 或 Body：`targetType`、`targetId` |
| 5 | GET | `/history` | 浏览历史分页 |
| 6 | GET | `/downloads` | 下载记录分页；`available` 反映当前是否仍可下载 |

无 token 或 admin token → **401**。

---

### 4.6 意见反馈（2，UserAuth）

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | POST | `/feedbacks` | 提交反馈 |
| 2 | GET | `/feedbacks` | 当前用户反馈分页列表 |

**POST Body**：

```json
{ "content": "希望增加四年级下册范文" }
```

- `content` 必填；trim 后非空；最长 **300** 字（按 rune 计）
- 新建 `status=pending`；成功返回单条 `FeedbackListItem`（不含 `userNickname`）

**GET Query**：`page`、`pageSize`（默认 1 / 20，最大 100）

**列表项字段**：`id`、`userId`、`content`、`status`（`pending` / `replied`）、`reply`、`repliedAt`、`createdAt`

- 仅返回**当前登录用户**自己的反馈
- 管理员回复后，用户在小程序「我的反馈」列表可见 `reply` 与 `repliedAt`

---

## 五、一期 MVP 边界附录

### 5.1 不存在的路径（旧课件方案已废弃）

以下路径**一期未实现**，请勿在客户端调用：

| 路径模式 | 说明 |
|----------|------|
| `/admin/api/v1/courses/*` | 旧课件 CRUD |
| `/admin/api/v1/orders/*` | 订单管理 |
| `/admin/api/v1/users/*` | 用户封禁等 |
| `/admin/api/v1/settings` | 全局设置 |
| `/api/v1/orders/*` | 小程序订单 |
| `/api/v1/pay/*` | 支付回调 |
| `/api/v1/courses/*` | 旧课件列表 |
| 视频号 / 讲解视频相关 | 一期禁止 |

### 5.2 鉴权混用边界

| 场景 | 期望 |
|------|------|
| 无 token 访问 `/mine` | 401 |
| 无 token 访问 `/feedbacks` | 401 |
| admin token 访问 `/favorites` | 401 |
| 无 token 访问 `/download-files/:id/sign-url` | 200/403（视资料权限，不强制登录） |
| 无 token 访问 `/track/view` | 200，`userId=0` |

---

## 六、关联文档

| 文档 | 说明 |
|------|------|
| [`api-self-test-checklist.md`](./api-self-test-checklist.md) | 全量手工自测勾选表 |
| [`download-permission-boundary.md`](./download-permission-boundary.md) | 下载权限边界矩阵 |
| [`postman/essay-mvp-api.postman_collection.json`](./postman/essay-mvp-api.postman_collection.json) | Postman / Apifox 接口集合 |
