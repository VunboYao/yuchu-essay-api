# 后端接口自测清单

> 配套：[api-reference.md](./api-reference.md)、[download-permission-boundary.md](./download-permission-boundary.md)、[postman/essay-mvp-api.postman_collection.json](./postman/essay-mvp-api.postman_collection.json)

本清单用于本地联调、上线前回归和二期改造前基线检查。接口范围以 `internal/router/router.go` 为准。

## 前置条件

- [ ] MySQL 已建库 `essay_mvp`，`config.yaml` 配置正确。
- [ ] 服务已启动，`GET /health` 返回 `ok`。
- [ ] 默认管理员 `admin / admin123` 可登录；生产环境须已改密。
- [ ] 已导入 Postman 集合与环境，或在 Apifox 中导入 Collection v2.1。
- [ ] 若测试 OSS，`oss` 配置完整，`POST /admin/api/v1/upload/sts` 返回 `mode=oss`。

推荐顺序：先 Admin 登录，再创建分类、标签、文章、优秀作文、下载资料，最后测试小程序公开接口和用户接口。

## 一、健康检查与认证

| # | 模块 | Method | Path | 鉴权 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|
| 1 | 健康 | GET | `/health` | 无 | `status=ok` | [ ] |
| 2 | 健康 | HEAD | `/health` | 无 | HTTP 200 | [ ] |
| 3 | Admin 登录 | POST | `/admin/api/v1/auth/login` | 无 | 返回 `token` | [ ] |

## 二、Admin 统计、审计、反馈

| # | 模块 | Method | Path | 鉴权 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|
| 4 | 统计 | GET | `/admin/api/v1/stats/overview` | Admin | 含 today/total 指标 | [ ] |
| 5 | 统计 | GET | `/admin/api/v1/stats/hot-articles` | Admin | 返回数组 | [ ] |
| 6 | 统计 | GET | `/admin/api/v1/stats/hot-searches` | Admin | 返回数组 | [ ] |
| 7 | 统计 | GET | `/admin/api/v1/stats/hot-downloads` | Admin | 返回数组 | [ ] |
| 8 | 统计 | GET | `/admin/api/v1/stats/ad-events` | Admin | 返回广告统计 | [ ] |
| 9 | 审计 | GET | `/admin/api/v1/operation-logs` | Admin | 分页结构正确 | [ ] |
| 10 | 反馈 | GET | `/admin/api/v1/feedbacks` | Admin | 分页结构正确 | [ ] |
| 11 | 反馈 | PUT | `/admin/api/v1/feedbacks/:id/reply` | Admin | 回复后状态更新 | [ ] |

## 三、Admin 内容与基础配置

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 12 | 分类 | GET | `/admin/api/v1/taxonomy-units` | Admin | 登录 | 分页列表 | [ ] |
| 13 | 分类 | POST | `/admin/api/v1/taxonomy-units` | Admin | 登录 | 创建成功，记下 `unitId` | [ ] |
| 14 | 分类 | PUT | `/admin/api/v1/taxonomy-units/:id` | Admin | #13 | 更新成功 | [ ] |
| 14b | 分类 | DELETE | `/admin/api/v1/taxonomy-units/:id` | Admin | 已禁用且无文章占用 | 删除成功或约束报错符合预期 | [ ] |
| 15 | 标签 | GET | `/admin/api/v1/tags` | Admin | 登录 | 分页列表 | [ ] |
| 16 | 标签 | POST | `/admin/api/v1/tags` | Admin | 登录 | 创建「免费」标签 | [ ] |
| 17 | 标签 | PUT | `/admin/api/v1/tags/:id` | Admin | #16 | 更新成功 | [ ] |
| 18 | 标签 | DELETE | `/admin/api/v1/tags/:id` | Admin | 已禁用且无关联 | 删除成功或约束报错符合预期 | [ ] |

## 四、Admin 文章、范文、下载资料

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 19 | 文章 | GET | `/admin/api/v1/articles` | Admin | 登录 | 分页列表，不含完整 `bodyImages` | [ ] |
| 20 | 文章 | POST | `/admin/api/v1/articles` | Admin | #13,#16 | 创建成功，记下 `articleId` | [ ] |
| 21 | 文章 | GET | `/admin/api/v1/articles/:id` | Admin | #20 | 含标签、正文长图 | [ ] |
| 22 | 文章 | PUT | `/admin/api/v1/articles/:id` | Admin | #20 | 更新成功 | [ ] |
| 23 | 文章 | PUT | `/admin/api/v1/articles/:id/status` | Admin | #20 | `status=1` 发布成功 | [ ] |
| 24 | 范文 | GET | `/admin/api/v1/essay-examples` | Admin | 登录 | 分页列表 | [ ] |
| 25 | 范文 | POST | `/admin/api/v1/essay-examples` | Admin | #20 | 创建成功，记下 `essayId` | [ ] |
| 26 | 范文 | GET | `/admin/api/v1/essay-examples/:id` | Admin | #25 | 详情含标签 | [ ] |
| 27 | 范文 | PUT | `/admin/api/v1/essay-examples/:id` | Admin | #25 | 更新成功 | [ ] |
| 28 | 范文 | DELETE | `/admin/api/v1/essay-examples/:id` | Admin | #25 | 删除成功 | [ ] |
| 29 | 资料 | GET | `/admin/api/v1/download-files` | Admin | 登录 | 分页列表 | [ ] |
| 30 | 资料 | POST | `/admin/api/v1/download-files` | Admin | #23 | `accessType=free` 且 tags 含「免费」，记下 `fileId` | [ ] |
| 31 | 资料 | PUT | `/admin/api/v1/download-files/:id` | Admin | #30 | 更新成功 | [ ] |
| 32 | 资料 | DELETE | `/admin/api/v1/download-files/:id` | Admin | 无下载记录或 `confirm=true` | 删除成功或约束报错符合预期 | [ ] |

## 五、Admin 运营配置与上传

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 33 | 首页入口 | GET | `/admin/api/v1/home-blocks` | Admin | 登录 | 列表 | [ ] |
| 34 | 首页入口 | POST | `/admin/api/v1/home-blocks` | Admin | 登录 | 创建成功，记下 `blockId` | [ ] |
| 35 | 首页入口 | PUT | `/admin/api/v1/home-blocks/:id` | Admin | #34 | 更新成功 | [ ] |
| 36 | 首页入口 | DELETE | `/admin/api/v1/home-blocks/:id` | Admin | 已禁用 | 删除成功 | [ ] |
| 36b | 搜索热门词 | GET | `/admin/api/v1/hot-search-keywords` | Admin | 登录 | 列表 | [ ] |
| 36c | 搜索热门词 | POST | `/admin/api/v1/hot-search-keywords` | Admin | 登录 | 创建成功，记下 `keywordId` | [ ] |
| 36d | 搜索热门词 | PUT | `/admin/api/v1/hot-search-keywords/:id` | Admin | #36c | 更新成功 | [ ] |
| 36e | 搜索热门词 | DELETE | `/admin/api/v1/hot-search-keywords/:id` | Admin | 已禁用 | 删除成功 | [ ] |
| 37 | 广告 | GET | `/admin/api/v1/ad-slots` | Admin | 登录 | 列表 | [ ] |
| 38 | 广告 | POST | `/admin/api/v1/ad-slots` | Admin | 登录 | Upsert 成功 | [ ] |
| 39 | 广告 | DELETE | `/admin/api/v1/ad-slots/:id` | Admin | 已禁用 | 删除成功或预置位后续可补回 | [ ] |
| 40 | 公众号 | GET | `/admin/api/v1/official-account-configs` | Admin | 登录 | 列表 | [ ] |
| 41 | 公众号 | POST | `/admin/api/v1/official-account-configs` | Admin | 登录 | Upsert 成功 | [ ] |
| 42 | 公众号 | DELETE | `/admin/api/v1/official-account-configs/:id` | Admin | 已禁用 | 删除成功 | [ ] |
| 43 | 上传 | POST | `/admin/api/v1/upload/sts` | Admin | 登录 | OSS 配置完整时返回 `mode=oss` | [ ] |

## 六、小程序公开接口

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 44 | 登录 | POST | `/api/v1/wx/login` | 无 | 微信 code 或开发 code | 返回 `token`、`user` | [ ] |
| 45 | 首页 | GET | `/api/v1/home` | 无 | #23,#25,#34 | 含 `recommendedArticles`、`hotArticles`、`essayExamples`、`homeBlocks` | [ ] |
| 46 | 分类 | GET | `/api/v1/categories` | 无 | #13 | 返回启用中的分类单元扁平列表 | [ ] |
| 47 | 文章 | GET | `/api/v1/articles` | 无 | #23 | 仅返回已发布文章 | [ ] |
| 48 | 文章 | GET | `/api/v1/articles/:id` | 无 | #23 | 详情可读，`view_count` 递增 | [ ] |
| 49 | 范文 | GET | `/api/v1/essays` | 无 | #25 | 返回列表 | [ ] |
| 50 | 范文 | GET | `/api/v1/essays/:id` | 无 | #25 | 返回详情 | [ ] |
| 51 | 搜索 | GET | `/api/v1/search?keyword=测试` | 无 | #23 | 返回结果列表 | [ ] |
| 51b | 搜索热门词 | GET | `/api/v1/hot-search-keywords` | 无 | #36c | 仅返回启用项的 `string[]` | [ ] |
| 52 | 广告 | GET | `/api/v1/ad-slots?page=mine` | 无 | #38 | 仅返回启用项 | [ ] |
| 53 | 公众号 | GET | `/api/v1/official-account-configs?scene=mine` | 无 | #41 | 返回启用配置或空对象 | [ ] |

## 七、小程序下载与埋点

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 54 | 下载 | GET | `/api/v1/download-files/:id/sign-url` | OptionalUser | #30 | 免费资料返回签名 URL | [ ] |
| 55 | 下载 | POST | `/api/v1/download-files/:id/ad-unlock` | OptionalUser | gate 已启用 | 返回 `unlockToken`；gate 未启用时 400 | [ ] |
| 56 | 埋点 | POST | `/api/v1/track/view` | OptionalUser | #23 | `code=0`，未登录 `userId=0` | [ ] |
| 57 | 埋点 | POST | `/api/v1/track/share` | OptionalUser | #23 | 文章 `share_count` 递增 | [ ] |
| 58 | 埋点 | POST | `/api/v1/track/search` | OptionalUser | — | `keyword` 必填 | [ ] |
| 59 | 埋点 | POST | `/api/v1/track/ad` | OptionalUser | #38 | `eventType=expose/click` 可写入 | [ ] |

下载边界专项必须再按 [download-permission-boundary.md](./download-permission-boundary.md) 覆盖 403、gate 和 token 过期场景。

## 八、小程序用户中心

以下接口均需 User token，可使用 #44 返回的 token。

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|---|---|---|---|---|---|---|
| 60 | 我的 | GET | `/api/v1/mine` | User | #44 | 返回收藏/浏览/下载计数 | [ ] |
| 61 | 收藏 | GET | `/api/v1/favorites` | User | #44 | 分页列表 | [ ] |
| 62 | 收藏 | POST | `/api/v1/favorites` | User | #44,#23 | 收藏成功，重复收藏幂等 | [ ] |
| 63 | 收藏 | DELETE | `/api/v1/favorites` | User | #62 | 取消成功 | [ ] |
| 64 | 历史 | GET | `/api/v1/history` | User | #56 | 分页列表 | [ ] |
| 65 | 下载 | GET | `/api/v1/downloads` | User | #54 | 含 `available` 字段 | [ ] |
| 66 | 资料 | GET | `/api/v1/user/profile` | User | #44 | 返回当前用户资料 | [ ] |
| 67 | 资料 | PUT | `/api/v1/user/profile` | User | #44 | 更新昵称/头像等字段 | [ ] |
| 68 | 头像 | POST | `/api/v1/user/avatar` | User | #44 | 本地上传成功或按环境限制返回明确错误 | [ ] |
| 69 | 反馈 | POST | `/api/v1/feedbacks` | User | #44 | 创建反馈成功 | [ ] |
| 70 | 反馈 | GET | `/api/v1/feedbacks` | User | #69 | 仅返回当前用户反馈 | [ ] |

## 九、鉴权边界

| # | 场景 | Method | Path | 期望 | 通过 |
|---|---|---|---|---|---|
| 71 | 无 token 访问用户接口 | GET | `/api/v1/mine` | 401 | [ ] |
| 72 | admin token 访问用户接口 | GET | `/api/v1/favorites` | 401 | [ ] |
| 73 | 无 token 访问 Admin | GET | `/admin/api/v1/articles` | 401 | [ ] |
| 74 | user token 访问 Admin | GET | `/admin/api/v1/stats/overview` | 401 | [ ] |
| 75 | 无 token 写埋点 | POST | `/api/v1/track/view` | 200，`userId=0` | [ ] |
| 76 | 付费资料请求 sign-url | GET | `/api/v1/download-files/:id/sign-url` | 403 | [ ] |

## 十、一期不存在路径审计

以下路径应返回 404：

| 路径 | 说明 | 已确认 |
|---|---|---|
| `GET /api/v1/orders` | 订单 | [ ] |
| `POST /api/v1/pay/notify` | 支付回调 | [ ] |
| `GET /admin/api/v1/orders` | 后台订单 | [ ] |
| `GET /admin/api/v1/products` | 后台商品 | [ ] |
| `GET /api/v1/courses` | 旧课件 | [ ] |

## 十一、自动化回归

```bash
cd course_server && go test ./test/...
```

- [ ] 全部测试通过。
- [ ] 下载权限相关测试通过。
- [ ] 无订单、支付、视频相关可调用接口。

## 十二、验收标准

- [ ] 健康检查、Admin、小程序公开接口、下载、埋点、用户中心均通过。
- [ ] 下载权限矩阵已手工或 Postman 验证。
- [ ] Postman / Apifox 集合与本文档核心路径一致。
- [ ] 二期未实现路径保持 404。
