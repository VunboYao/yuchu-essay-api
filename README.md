# course_server

内容型作文小程序 MVP 后端。

Agent 核心规则见 [AGENTS.md](./AGENTS.md)（MVP 边界、开发流程、注释要求）。

## 当前已实现

- Go + Gin 工程骨架
- MySQL/GORM 模型（按共用/后台/小程序分文件）与自动迁移、表/列注释
- Redis 连接占位
- 管理员登录与 JWT
- 文章、正文图片、优秀作文、分类、标签、下载资料基础 API
- 小程序首页、分类、文章、优秀作文、搜索、收藏、历史、下载记录接口骨架
- 免费下载权限校验：`access_type=free` 且标签包含「免费」

## 启动

1. 创建数据库：

```sql
CREATE DATABASE essay_mvp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 修改 `config.yaml` 中的 MySQL/Redis 配置。

3. 启动：

```bash
go mod tidy
go run .
```

开发期热重载（需先安装 [air](https://github.com/air-verse/air)）：

```bash
go install github.com/air-verse/air@latest
air   # 在 course_server 目录执行，配置见 .air.toml
```

默认后台管理员：

```text
admin / admin123
```

Redis 连接失败时仅打印警告，不阻断服务启动。

## 重要边界

第一期不包含订单、支付、视频、会员、分销、付费下载。完整边界与 Agent 开发约束见 [AGENTS.md](./AGENTS.md)。

---

## 目录结构

```text
course_server/
├── main.go                 # 启动：config → DB/Redis → migrate/seed → router
├── config.yaml             # 运行时配置（勿提交生产密钥）
└── internal/
    ├── config/             # YAML 配置加载
    ├── database/           # MySQL/Redis 连接、AutoMigrate、Seed
    ├── model/              # GORM 实体、常量与持久化辅助
    │   ├── models.go           # 业务常量
    │   ├── validate.go         # NextSortOrder 等 DB 辅助
    │   ├── model_write.go      # UpdateMap、唯一性检查、按业务键查询
    │   ├── model_shared.go     # 【共用】内容/配置
    │   ├── model_admin.go      # 【后台】
    │   ├── model_miniprogram.go  # 【小程序】
    │   ├── model_order.go      # 【二期】订单占位
    │   └── db_comments.go      # 表级 COMMENT
    ├── dto/                # HTTP 入参/出参 DTO
    │   ├── validate.go         # ResolveStatus 等入参解析
    │   └── admin/              # Admin 写操作 Request DTO 与 Validate
    ├── handler/            # HTTP 处理（按领域分文件）
    │   ├── helpers.go          # 分页、鉴权上下文
    │   └── helpers_write.go    # UpdateAndReload 等写操作辅助
    ├── middleware/         # CORS、AdminAuth、UserAuth
    ├── router/             # 路由注册（唯一入口）
    └── pkg/
        ├── jwt/            # JWT 签发/解析
        └── response/       # 统一响应
```

**新增代码放置规则：**

- 新 API 路由 → `internal/router/router.go`
- 新表/字段 → 按使用端写入 `internal/model/` 对应文件（见「数据模型文件结构」），并同步 `internal/database/migrate.go` 与 `db_comments.go`
- Admin 写操作 Request DTO → `internal/dto/admin/<domain>.go`
- Handler 方法挂在 `Server` 结构体上（`internal/handler/server.go`）
- 不在 `main.go` 写业务逻辑

### 为什么使用 `internal/` 目录

Go 规定：路径中包含 `internal` 的包，**只能被该 `internal` 目录的父目录及其子目录下的代码 import**；项目外的其他模块无法引用。这是 Go 官方推荐的封装方式（见 [Go 1.4 Release Notes — Internal packages](https://go.dev/doc/go1.4#internalpackages)）。

在本项目中：

```text
course_server/              ← internal 的「父目录」
├── main.go                 ✅ 可以 import course_server/internal/...
└── internal/
    ├── config/             ❌ 外部项目不能 import
    ├── handler/
    └── ...
```

**作用：**

| 目的 | 说明 |
|---|---|
| 封装实现细节 | 对外不暴露 handler、database 等内部实现，后续重构不影响外部 |
| 防止误用 | 其他仓库无法 `import "course_server/internal/handler"` 当公共库使用 |
| 明确边界 | 一眼可辨：`internal/` 下为项目私有代码，根目录 `main.go` 为对外入口 |

**与 `internal/pkg/` 的区别：** 本项目的 `internal/pkg/jwt`、`internal/pkg/response` 仍在 `internal/` 下，同样不可被外部引用。这里的 `pkg` 仅表示「项目内可复用的工具包」，并非对外公开的公共库。若将来需要被其他项目 import，应放在模块根目录的 `pkg/`（且不在 `internal/` 内）。

## API 规范

**路由前缀：**

| 客户端 | 前缀 |
|---|---|
| 后台管理 | `/admin/api/v1` |
| 微信小程序 | `/api/v1` |
| 健康检查 | `GET /health` |

**统一响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

**分页响应**（`data` 内）：

```json
{
  "list": [],
  "total": 0,
  "page": 1,
  "pageSize": 20
}
```

使用 `internal/pkg/response` 包：

- 成功：`response.Success()` / `response.Page()`
- 错误：`response.BadRequest` / `Unauthorized` / `Forbidden` / `NotFound` / `ServerError`
- HTTP 状态码与 body 中的 `code` 保持一致

**分页默认值**（`internal/handler/helpers.go`）：

- `page` 默认 1，`pageSize` 默认 20，最大 100
- 列表默认排序：`sort_order DESC, id DESC`

**JSON 字段命名：** 模型使用 camelCase JSON tag（如 `adminId`、`coverImg`），新增字段保持一致。

## 鉴权

**Header：** `Authorization: Bearer <token>`

JWT 区分 `type`：`admin` / `user`，不可混用。中间件注入 `auth_id`、`auth_subject`、`auth_role`；handler 通过 `currentAuthID(c)` 读取当前用户 ID。

### Authorization 请求头说明

HTTP 的 `Authorization` 不是「专门放 token 的字段」，而是「认证方式 + 凭证」：

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

| 写法 | 示例 | 说明 |
|---|---|---|
| 标准（推荐） | `Authorization: Bearer <token>` | OAuth/JWT 通用约定，服务端一眼可知按 Bearer Token 解析 |
| 兼容 | `Authorization: <token>` | 后端 `bearerToken` 已兼容裸 token，便于 Apifox 等工具直接粘贴 |

**为什么推荐带 `Bearer` 前缀？**

- `Authorization` 还支持 `Basic`、`Digest` 等 scheme，前缀标明解析方式
- 网关、第三方 API、监控工具均默认识别 `Bearer <token>`
- 前端多写 7 个字符（`Bearer `），换来规范与跨服务兼容

**前端写法（推荐）：**

```javascript
// axios 拦截器
config.headers.Authorization = `Bearer ${token}`

// fetch
headers: { Authorization: `Bearer ${token}` }
```

登录接口返回的 `data.token` 只存 token 本身；发请求时再拼 `Bearer ` 前缀。Apifox/Postman 选 **Bearer Token** 鉴权类型时，工具会自动拼接，无需手写前缀。

**常见 401 原因：** Header 名写对但值格式不对（如只传裸 token 且后端未兼容、或 `Bearer` 与 token 之间缺少空格）。开发调试可先用裸 token，联调与上线建议统一 `Bearer ${token}`。

| 场景 | 入口 | 中间件 |
|---|---|---|
| 后台登录 | `POST /admin/api/v1/auth/login` | 无 |
| 后台接口 | `/admin/api/v1/*`（除 login） | `AdminAuth` |
| 小程序登录 | `POST /api/v1/wx/login` | 无 |
| 需登录的小程序接口 | `/api/v1/mine`、`/favorites`、`/history`、`/downloads` | `UserAuth` |
| 公开小程序接口 | `/api/v1/home`、`/articles`、`/download-files/:id/sign-url` 等 | 无 |

- 管理员密码：bcrypt 哈希
- 微信登录：当前为 dev fallback（`openid = "dev_" + code`）；配置 `wechat.app_id/app_secret` 后应替换为 code2session

## 下载权限

公开下载签名接口：`GET /api/v1/download-files/:id/sign-url`

须同时满足以下三条件才下发 URL，否则返回 403：

1. `status = 1`（`StatusEnabled`）
2. `access_type = "free"`（`AccessTypeFree`）
3. `tags` 字符串包含「免费」

`purchase_required` / `member_only` / `internal` 一期一律拒绝。

下载成功时：写入 `download_logs`、递增 `download_count`；若 `owner_type=article`，同步递增关联文章的下载数。

## 数据模型约定

### 数据模型文件结构

`internal/model/` 按**使用端**分文件，避免一期内容与二期订单/支付混写：

| 文件 | 归属 | 表（一期） |
|------|------|------------|
| `models.go` | 公共 | 业务常量（`StatusEnabled`、`AccessTypeFree` 等），**不放 struct** |
| `model_shared.go` | 【共用】后台写、小程序读 | `taxonomy_units`、`tags`、`articles`、`article_tags`、`essay_examples`、`essay_example_tags`、`download_files`、`ad_slots`、`home_blocks`、`official_account_configs` |
| `model_admin.go` | 【后台】 | `admins`、`admin_operation_logs`、`system_settings` |
| `model_miniprogram.go` | 【小程序】 | `users`、`download_logs`、`favorites`、`view_logs`、`search_logs`、`share_logs`、`ad_events` |
| `model_order.go` | 【二期】占位 | 订单/支付表**仅写在此文件**（一期禁止实现逻辑） |
| `db_comments.go` | 迁移辅助 | MySQL 表级 `COMMENT`（`ApplyTableComments`） |

**新增/修改表时：**

1. 在对应 `model_*.go` 增加 struct，字段加 `gorm:"comment:..."` 列注释
2. 在 `internal/database/migrate.go` 的 `AutoMigrate` 中按分组追加（顺序：**共用 → 后台 → 小程序 → 二期订单**）
3. 在 `db_comments.go` 的 `tableCommentEntries` 中追加表注释（与 migrate 分组一致）

**二期订单/支付约定：**

- 新表只放在 `model_order.go`（或 `model_order_*.go`），**不改** `model_shared.go` 中 `Article` 等内容主表
- 权限继续通过 `download_files.owner_type` / `owner_id` / `access_type` 接入，不在内容表上堆订单字段

### 其它约定

- API 时间字段（`createdAt`、`updatedAt`、`publishedAt` 等）统一返回东八区字符串 `yyyy-MM-dd HH:mm:ss`；存储仍为 MySQL `datetime`（类型见 `internal/pkg/jsontime`）
- 状态与权限常量定义在 `models.go`，禁止魔法数字/字符串散落在 handler
- 查询字段须加 index：`openid`、`slug`、`target_type/target_id`、`created_at` 及分类筛选字段
- 种子数据：默认管理员 `admin/admin123`（`internal/database/seed.go`）
- 启动时 `AutoMigrate` 后执行 `ApplyTableComments`，库表可见【后台】/【小程序】/【共用】前缀

**文章正文图片（`bodyImages`）：**

- 存储于 `articles.body_images`，JSON 数组，按序渲染
- 创建/更新文章时在请求体中传入；列表接口省略该字段，详情接口返回完整数组
- 单篇最多 20 张，URL 不可为空

```json
{
  "bodyImages": [
    "https://oss.example.com/article/1-part1.jpg",
    "https://oss.example.com/article/1-part2.jpg"
  ]
}
```

> 历史 `article_sections` 表已废弃，一期不再使用。

## Admin 写操作约定

后台 Create / Update / Upsert 曾直接使用 `ShouldBindJSON(&model.Article)` 等方式绑定 GORM 实体，存在三类常见问题。现已统一改为 **Request DTO（dto/admin）+ 入参校验 + model 持久化辅助 + 通用写库辅助**，新接口必须遵循；Agent 约束见 [AGENTS.md §6](./AGENTS.md#6-admin-写操作约定)。

### 常见问题

| 类型 | 表现 | 根因 |
|------|------|------|
| **脏数据** | 客户端可传 `id`、`createdAt`，或篡改 `viewCount` 等计数 | 请求体直接绑定数据库 struct |
| **必输缺失** | `{}` 或空字符串能入库（如 `grade=0`、`semester=""`） | GORM `not null` 不等于业务必填；`int` 零值为 `0` |
| **status / 零值更新** | 创建时无法设 `status=0`；更新时禁用、清零、关推荐改不动 | `if status==0 { status=1 }` 混淆「未传」与「传 0」；`Updates(struct)` 跳过零值 |

其中 **status** 字段因 `0` 有业务含义（禁用/草稿），须用 `*int` 区分「JSON 未传字段（`nil`）」与「显式传 `0`」。

### 解决方案

**1. Request DTO（`internal/dto/admin/*.go`）**

- **禁止**绑定 GORM 实体（如 `model.Tag`、`model.Article`）
- **允许**绑定 Request DTO（如 `admindto.CreateTagRequest`）——DTO 不含 `id`、时间戳、冗余计数
- 已落地文件示例：`tag.go`、`taxonomy_unit.go`、`article.go`、`download_file.go`、`essay_example.go`、`config.go`

**2. 校验集中在 dto/admin 包**

```go
tag, err := req.ValidateForCreate()              // 创建
tag, err := req.ValidateForUpdate(existing)      // 更新
status, err := dto.ResolveStatus(ptr, default)   // 可选 status
```

**3. 写库辅助（`internal/handler/helpers_write.go`）**

- 更新使用 `map[string]interface{}` + `UpdateAndReload`，避免零值被跳过、响应返回旧数据
- 运营配置 Upsert 按唯一键（`blockKey` / `slotKey` / `scene`）查找后 Create 或 Update，不依赖客户端传 `id`

**4. handler 模板（薄层）**

```go
var req admindto.CreateTagRequest
if err := c.ShouldBindJSON(&req); err != nil { ... }
tag, err := req.ValidateForCreate()
if err != nil { response.BadRequest(c, err.Error()); return }
if err := model.CheckTagNameDuplicate(database.DB, tag.Name, 0); err != nil { ... }
database.DB.Create(&tag)
```

### 已迁移接口

| 模块 | Handler | Request 文件 |
|------|---------|--------------|
| 标签 | `CreateTag` / `UpdateTag` | `dto/admin/tag.go` |
| 分类单元 | `CreateTaxonomyUnit` / `UpdateTaxonomyUnit` | `dto/admin/taxonomy_unit.go` |
| 作文专题 | `CreateArticle` / `UpdateArticle` | `dto/admin/article.go` |
| 下载资料 | `CreateDownloadFile` / `UpdateDownloadFile` | `dto/admin/download_file.go` |
| 优秀作文 | `CreateEssayExample` / `UpdateEssayExample` | `dto/admin/essay_example.go` |
| 运营配置 | `UpsertHomeBlock` / `UpsertAdSlot` / `UpsertOfficialAccountConfig` | `dto/admin/config.go` |

未迁移：`track.go` 埋点、`auth.go` 登录（服务端构造数据，风险较低）。`UpdateArticleStatus` 等专用接口保持独立 struct。

## 代码风格

- 参数绑定：`c.ShouldBindJSON` + `binding:"required"`；**Admin 写操作**须用 Request DTO，见上文 [Admin 写操作约定](#admin-写操作约定)
- DB 访问：全局 `database.DB`，复杂写操作用 `Transaction`
- 错误消息：英文简短描述（如 `"failed to list articles"`）
- 复用 `pageParams` / `paginate` / `response.*`，不新增 service/repository 层除非明确要求重构
- **注释规范详见 [AGENTS.md](./AGENTS.md)**

## 配置说明

`config.yaml` 各段含义：

| 段 | 用途 |
|---|---|
| `server.addr` | 监听地址，默认 `:8080` |
| `database.dsn` | MySQL 连接串 |
| `redis` | Redis 地址、密码、DB |
| `jwt` | 签名密钥与 issuer |
| `oss` | 对象存储（封面上传、资料文件） |
| `wechat` | 小程序 AppID / AppSecret |

## 关联文档

| 文档 | 说明 |
|---|---|
| [AGENTS.md](./AGENTS.md) | Agent 核心规则（边界、开发流程、注释） |
| [prd/001-产品整体结构.md](../prd/001-产品整体结构.md) | 产品整体结构 |
| [prd/003-技术文档介绍.md](../prd/003-技术文档介绍.md) | 技术方案与接口规范 |
| [prd/004](../prd/004-开发计划-第一阶段：后端基础与数据模型（Day1-7）.md) ~ [005](../prd/005-开发计划-第二阶段：后端业务接口与统计（Day8-12）.md) | 后端开发计划 |
| [study_doc/](../study_doc/) | Go/Gin 学习笔记（参考，非权威契约） |
| [Admin/](../Admin/) | 后台管理前端 |
