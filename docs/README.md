# 后端文档入口

本目录保存 `course_server` 已交付能力的接口、权限、OSS、域名和验收资料。新增文档前先判断能否合并到现有文档，避免再产生阶段过程文档。

## 文档索引

| 文档 | 用途 | 维护依据 |
|---|---|---|
| [api-reference.md](./api-reference.md) | Admin 与小程序接口清单 | 以 `internal/router/router.go` 为唯一事实来源 |
| [api-self-test-checklist.md](./api-self-test-checklist.md) | 手工/Apifox/Postman 自测清单 | 与接口文档同步 |
| [download-permission-boundary.md](./download-permission-boundary.md) | 下载权限和激励视频门槛矩阵 | 与 `IsPublicFreeDownload` 和下载 handler 同步 |
| [domain-vunbo.md](./domain-vunbo.md) | 生产域名、微信合法域名、DNS/证书口径 | 与部署现状同步 |
| [oss-setup-guide.md](./oss-setup-guide.md) | 阿里云 OSS、STS、CDN 配置和验收 | 与 `config.yaml` 的 `oss` 段同步 |
| [postman/essay-mvp-api.postman_collection.json](./postman/essay-mvp-api.postman_collection.json) | Postman / Apifox 集合 | 与接口文档同步 |
| [postman/local.postman_environment.json](./postman/local.postman_environment.json) | 本地环境变量 | 与本地端口和 token 变量同步 |

## 维护规则

- 接口变更先改代码，再同步 `api-reference.md` 和 `api-self-test-checklist.md`。
- 下载权限变更必须同步 `download-permission-boundary.md`。
- 部署命令放到 `../deploy/`；本目录只放配置口径、验收和依赖服务说明。
- 不再新增 Day 计划、临时待办、过程踩坑文档；有价值的踩坑合并到对应手册的“常见问题”。
