# 下载权限边界测试

> 实现源码：[`internal/handler/download.go`](../internal/handler/download.go)、[`internal/model/model_shared.go`](../internal/model/model_shared.go)（`IsPublicFreeDownload`）  
> 自动化测试：[`test/handler/sign_download_url_test.go`](../test/handler/sign_download_url_test.go)、[`test/handler/unlock_download_ad_test.go`](../test/handler/unlock_download_ad_test.go)

---

## 一、权限模型说明

### 两条独立权限线

| 对象 | 字段 | 作用 |
|------|------|------|
| 文章 | `articles.view_access_type` | 控制「能不能看」文章详情 |
| 资料 | `download_files.access_type` | 控制「能不能下」资料文件 |

二者**互不影响**：文章可免费查看，资料仍可能因 `access_type` 或 `tags` 被拦截下载。

### 一期免费下载三条件

须**同时**满足才下发 `sign-url`：

1. `status = 1`（启用）
2. `access_type = "free"`
3. `tags` 字符串包含「免费」

`purchase_required` / `member_only` / `internal` 一期一律 **403**。

### tags 与 access_type 分工

- `access_type`：权限主判据（结构化枚举）
- `tags` 含「免费」：仅作用于 `free` 模式的运营软开关（临时下架可去掉「免费」，无需改 `access_type`）

---

## 二、`GET /api/v1/download-files/:id/sign-url` 判定矩阵

| # | status | access_type | tags 含「免费」 | gate 启用 | unlockToken | HTTP | message |
|---|--------|-------------|----------------|-----------|-------------|------|---------|
| 1 | 0 | free | 是 | 否 | — | 403 | download file is not free |
| 2 | 1 | free | 否 | 否 | — | 403 | download file is not free |
| 3 | 1 | purchase_required | 是 | 否 | — | 403 | download file is not free |
| 4 | 1 | member_only | 是 | 否 | — | 403 | download file is not free |
| 5 | 1 | internal | 是 | 否 | — | 403 | download file is not free |
| 6 | 1 | free | 是 | 否 | — | **200** | — |
| 7 | 1 | free | 是 | 是 | 缺失 | 403 | rewarded video required |
| 8 | 1 | free | 是 | 是 | 无效/过期 | 403 | invalid or expired unlock token |
| 9 | 1 | free | 是 | 是 | 有效 | **200** | — |

**gate 启用条件**（`ad_slots` 表 `slot_key=download_rewarded_video`）：

- `status = 1`
- `ad_type = rewarded_video`
- `ad_unit_id` 非空

---

## 三、`POST /api/v1/download-files/:id/ad-unlock` 判定矩阵

| # | 资料是否免费 | gate 启用 | HTTP | message |
|---|-------------|-----------|------|---------|
| 1 | 否 | — | 403 | download file is not free |
| 2 | 是 | 否 | 400 | download rewarded video gate is not enabled |
| 3 | 是 | 是 | **200** | 返回 `unlockToken`、`expiresIn`（300 秒） |

---

## 四、Apifox / Postman 手工测试步骤

### 前置

1. 服务已启动（`go run .` 或 air）
2. 已完成 Admin 登录，取得 `adminToken`
3. 已创建并发布一篇作文专题（记下 `articleId`）

### 步骤 A：创建可免费下载的测试资料

```http
POST /admin/api/v1/download-files
Authorization: Bearer {{adminToken}}
Content-Type: application/json

{
  "ownerType": "article",
  "ownerId": {{articleId}},
  "fileName": "边界测试资料.pdf",
  "tags": "免费,测试",
  "accessType": "free",
  "fileType": "pdf",
  "fileUrl": "https://cdn.vunbo.com/essay-mvp/boundary-test.pdf",
  "status": 1
}
```

保存响应中的 `id` 为 `downloadFileId`。

### 步骤 B：gate 关闭时直接下载（矩阵 #6）

```http
GET /api/v1/download-files/{{downloadFileId}}/sign-url
```

期望：`code=0`，`data.url` 非空，`download_count` 递增。

### 步骤 C：去掉「免费」标签（矩阵 #2）

```http
PUT /admin/api/v1/download-files/{{downloadFileId}}
Authorization: Bearer {{adminToken}}

{ "...": "...", "tags": "测试", "accessType": "free", ... }
```

再次 `GET sign-url` → 期望 **403**。

### 步骤 D：恢复免费标签并启用 gate

1. PUT 资料，`tags` 改回含「免费」
2. Upsert 广告位：

```http
POST /admin/api/v1/ad-slots
Authorization: Bearer {{adminToken}}

{
  "slotKey": "download_rewarded_video",
  "slotName": "下载激励视频",
  "page": "download",
  "adType": "rewarded_video",
  "adUnitId": "adunit-test-001",
  "status": 1
}
```

3. `GET sign-url` 不带 token → **403** `rewarded video required`（矩阵 #7）
4. `POST /api/v1/download-files/{{downloadFileId}}/ad-unlock` → 取得 `unlockToken`
5. `GET sign-url?unlockToken={{unlockToken}}` → **200**（矩阵 #9）

### 步骤 E：非 free access_type（矩阵 #3）

PUT 资料 `accessType=purchase_required`（保留 tags 含「免费」）→ `sign-url` **403**。

---

## 五、自动化回归

```bash
cd course_server
go test ./test/...
```

覆盖项：

| 测试文件 | 用例 |
|----------|------|
| `test/model/download_permission_test.go` | `IsPublicFreeDownload` 8 种组合 |
| `test/handler/sign_download_url_test.go` | 404、403 禁用/付费、403 gate、200 无 gate、200 有效 token |
| `test/handler/unlock_download_ad_test.go` | 403 非免费、400 gate 未启用、200 签发 token |
| `test/pkg/adunlock/adunlock_test.go` | unlockToken 签发与校验 |

---

## 六、二期说明（未实现）

`purchase_required` / `member_only` 在二期将按用户权益（`user_entitlements`）放行，**不看** `tags`。一期所有非 `free` 类型保持 403。
