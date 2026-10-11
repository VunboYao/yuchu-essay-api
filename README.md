# yuchu-essay-api

Backend for **予初作文 (Yuchu Essay)** — a content platform that turns a Chinese primary-school
writing WeChat account into a searchable, categorised library of writing samples and
downloadable teaching materials.

**Go · Gin · GORM · MySQL · Redis · Alibaba Cloud OSS**

**Live:** https://api.vunbo.com/health

Sibling repositories:

| Repo | What | Stack |
|---|---|---|
| [`yuchu-essay-admin`](https://github.com/VunboYao/yuchu-essay-admin) | Operations admin panel | React 19 · Vite · TanStack Query |
| [`yuchu-essay-web`](https://github.com/VunboYao/yuchu-essay-web) | Public content website | Nuxt 4 · SSR/ISR |

---

## What it does

Teachers, students and parents use a WeChat mini program to browse writing samples by
textbook unit, search them, and download teaching materials. This service is the whole
backend: content management, mini-program API, authentication for two separate audiences,
object-storage upload and signed downloads.

83 route registrations across three groups:

| Client | Prefix | Auth |
|---|---|---|
| Admin panel | `/admin/api/v1` | `AdminAuth` — JWT with `type: admin` |
| Mini program | `/api/v1` | public, plus `UserAuth` on `/mine`, `/favorites`, `/history`, `/downloads` |
| Health | `GET /health` | none |

---

## Stack

| | |
|---|---|
| Language | Go 1.25 |
| HTTP | Gin 1.10 |
| ORM | GORM 1.25 (MySQL in production, SQLite in tests) |
| Cache / counters | Redis 9 |
| Auth | `golang-jwt/jwt` v5, bcrypt password hashes |
| Storage | Alibaba Cloud OSS — STS credentials for direct upload, presigned URLs for private files |
| WeChat | `code2session` login, with a dev fallback |

---

## Layout

```text
.
├── main.go                  # config → DB/Redis → migrate/seed → router
├── cmd/reset-admin-password # small operational CLI
└── internal/
    ├── config/              # YAML config loading
    ├── database/            # connections, AutoMigrate, seed data
    ├── model/               # GORM entities
    │   ├── models.go        # business constants — no structs
    │   ├── model_write.go   # update-map and uniqueness helpers
    │   ├── model_shared.go       # content & config
    │   ├── model_admin.go        # admin-side
    │   ├── model_miniprogram.go  # mini-program-side
    │   └── model_order.go        # phase-2 placeholder, no logic
    ├── dto/                 # HTTP request DTOs, including dto/admin/*.go
    ├── handler/             # HTTP handlers, one file per domain
    ├── middleware/          # CORS, AdminAuth, UserAuth
    ├── router/              # single place where routes are registered
    └── pkg/                 # jwt, response, ossutil, wechat, jsontime, …
```

Two conventions worth knowing before you add code:

- **Routes are registered only in `internal/router/router.go`.** Nothing else wires HTTP.
- **Model files are split by *which client uses the table*, not by topic.** Phase-2
  order/payment tables go in `model_order.go` only — the content tables are not to be
  extended with commerce fields.

---

## Quickstart

Requires Go 1.25+ and MySQL. Redis is optional (a failed connection logs a warning and
does not block startup).

```bash
# 1. Database
mysql -e "CREATE DATABASE essay_mvp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. Config
cp config.yaml.example config.yaml
$EDITOR config.yaml          # point database.dsn at your MySQL

# 3. Run
go mod download
go run .                     # http://localhost:8080, /health should return 200
```

The schema is created automatically: `AutoMigrate` syncs the models, then
`ApplyTableComments` writes the MySQL table comments. Seed data creates one admin —
`admin / admin123`.

Hot reload during development needs [air](https://github.com/air-verse/air):

```bash
go install github.com/air-verse/air@latest
air                          # config in .air.toml
```

### Tests

```bash
go test ./...
```

**No database or Redis required.** `test/testutil.UseTestDB` swaps the global `database.DB`
for an in-memory SQLite instance and restores it when the test ends, so handler tests run
through real `httptest` requests against a real (in-memory) schema. 97 test functions across
33 files.

---

## API conventions

Single response envelope for everything:

```json
{ "code": 0, "message": "success", "data": {} }
```

`code === 0` means success. Non-zero is a business failure — the HTTP status and the body
`code` are kept consistent, and clients branch on HTTP status for transport failures
(notably `401`) and on `code` for business failures. Paginated `data` is
`{ list, total, page, pageSize }`, `pageSize` defaults to 20 and caps at 100.

Auth: `Authorization: Bearer <jwt>`. Tokens carry a `type` claim (`admin` | `user`) and the
two are not interchangeable.

Full endpoint list: [`docs/api-reference.md`](./docs/api-reference.md). A Postman
collection and a manual self-test checklist live in [`docs/`](./docs/).

---

## Three decisions worth explaining

### 1. Admin writes bind a request DTO, never a GORM entity

Binding `c.ShouldBindJSON(&model.Article{})` looks convenient and causes three specific
bugs, all of which showed up in this codebase:

| Symptom | Cause |
|---|---|
| Clients can set `id`, `createdAt`, or overwrite `viewCount` | The request body is bound to the persistence struct |
| `{}` inserts a row with empty required fields | GORM `not null` is not business validation; `int` zero-values pass |
| `status: 0` cannot be set on create, or cleared on update | `if status == 0 { status = 1 }` conflates "field missing" with "explicitly zero"; `Updates(struct)` skips zero values |

So every admin write goes through `internal/dto/admin/*.go`: a DTO that has no `id`, no
timestamps, no counters, plus a `ValidateForCreate` / `ValidateForUpdate` method. Optional
`status` is `*int`, because `0` is meaningful for this field — it means *disabled* — so the
code must distinguish `nil` from a deliberate `0`. Updates use
`map[string]interface{}` + `UpdateAndReload` rather than `Updates(struct)`.

### 2. Viewing and downloading are two independent permission lines

An article has a `view_access_type`; a downloadable file has an `access_type`. They never
consult each other. Phase two adds paid downloads and membership tiers — that change lands
entirely on the download line and does not touch the content tables.

### 3. A free download passes three conditions, and only three

`GET /api/v1/download-files/:id/sign-url` returns a URL only when all of these hold, and
returns `403` otherwise:

1. `status = 1`
2. `access_type = 'free'`
3. the `tags` string contains the word 免费 ("free")

Condition 3 is not redundant with condition 2 — it is the *operations* switch. Pulling the
word 免费 out of the tags takes the file offline for users without touching `access_type`,
so the permission model stays clean while operations keeps a one-field override. The
matrix and the manual steps are in
[`docs/download-permission-boundary.md`](./docs/download-permission-boundary.md).

---

## Scope

This is phase one of the product, and the boundary is deliberate: **no payments, no orders,
no SKUs, no memberships, no user submissions.** `model_order.go` exists as a placeholder so
phase two has somewhere to land, and it contains no logic. A few permission fields reserve
values (`purchase_required`, `member_only`, `internal`) so phase two can be added without a
migration — none of them are granted in phase one.

---

## License

MIT — see [LICENSE](./LICENSE).

The original Chinese README from the private monorepo is kept at
[`README.zh-CN.md`](./README.zh-CN.md); some links in it point at files that stay private.
