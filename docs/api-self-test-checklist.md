# 后端全量自测清单

> 配套：[`api-reference.md`](./api-reference.md)、[`postman/essay-mvp-api.postman_collection.json`](./postman/essay-mvp-api.postman_collection.json)  
> 下载边界专项：[`download-permission-boundary.md`](./download-permission-boundary.md)

---

## 前置条件

- [ ] MySQL 已建库 `essay_mvp`，`config.yaml` 配置正确
- [ ] 服务已启动（`go run .` 或 `air`），`GET /health` 返回 ok
- [ ] 默认管理员 `admin / admin123` 可登录
- [ ] 已导入 Postman 集合与环境，或 Apifox 导入 Collection v2.1

**推荐顺序**：按表中 `#` 递增执行，确保数据依赖（文章 → 下载资料 → 小程序详情/下载）。

---

## 一、健康检查与认证（2）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 1 | 健康 | GET | `/health` | 无 | — | `status=ok` | [ ] |
| 2 | Admin | POST | `/admin/api/v1/auth/login` | 无 | — | `code=0`，返回 `token` | [ ] |

---

## 二、Admin 统计与审计（6）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 3 | 统计 | GET | `/admin/api/v1/stats/overview` | Admin | 登录 | `code=0`，含 today/total 字段 | [ ] |
| 4 | 统计 | GET | `/admin/api/v1/stats/hot-articles` | Admin | 登录 | 返回数组 | [ ] |
| 5 | 统计 | GET | `/admin/api/v1/stats/hot-searches` | Admin | 登录 | 返回数组 | [ ] |
| 6 | 统计 | GET | `/admin/api/v1/stats/hot-downloads` | Admin | 登录 | 返回数组 | [ ] |
| 7 | 统计 | GET | `/admin/api/v1/stats/ad-events` | Admin | 登录 | 返回统计对象 | [ ] |
| 8 | 审计 | GET | `/admin/api/v1/operation-logs` | Admin | 登录 | 分页 `list/total` | [ ] |

---

## 三、Admin 分类与标签（6）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 9 | 分类 | GET | `/admin/api/v1/taxonomy-units` | Admin | 登录 | 分页列表 | [ ] |
| 10 | 分类 | POST | `/admin/api/v1/taxonomy-units` | Admin | 登录 | 创建成功，返回 `id` | [ ] |
| 11 | 分类 | PUT | `/admin/api/v1/taxonomy-units/:id` | Admin | #10 | 更新成功 | [ ] |
| 12 | 标签 | GET | `/admin/api/v1/tags` | Admin | 登录 | 分页列表 | [ ] |
| 13 | 标签 | POST | `/admin/api/v1/tags` | Admin | 登录 | 创建「免费」标签 | [ ] |
| 14 | 标签 | PUT | `/admin/api/v1/tags/:id` | Admin | #13 | 更新成功 | [ ] |

---

## 四、Admin 文章（5）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 15 | 文章 | GET | `/admin/api/v1/articles` | Admin | 登录 | 分页列表 | [ ] |
| 16 | 文章 | POST | `/admin/api/v1/articles` | Admin | 登录 | 创建成功，记下 `articleId` | [ ] |
| 17 | 文章 | GET | `/admin/api/v1/articles/:id` | Admin | #16 | 含 `bodyImages`、标签 | [ ] |
| 18 | 文章 | PUT | `/admin/api/v1/articles/:id` | Admin | #16 | 更新成功 | [ ] |
| 19 | 文章 | PUT | `/admin/api/v1/articles/:id/status` | Admin | #16 | `status=1` 发布成功 | [ ] |

---

## 五、Admin 优秀作文（5）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 20 | 范文 | GET | `/admin/api/v1/essay-examples` | Admin | 登录 | 分页列表 | [ ] |
| 21 | 范文 | POST | `/admin/api/v1/essay-examples` | Admin | 登录 | 创建成功 | [ ] |
| 22 | 范文 | GET | `/admin/api/v1/essay-examples/:id` | Admin | #21 | 详情含标签 | [ ] |
| 23 | 范文 | PUT | `/admin/api/v1/essay-examples/:id` | Admin | #21 | 更新成功 | [ ] |
| 24 | 范文 | DELETE | `/admin/api/v1/essay-examples/:id` | Admin | #21 | 删除成功 | [ ] |

---

## 六、Admin 下载资料（4）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 25 | 资料 | GET | `/admin/api/v1/download-files` | Admin | 登录 | 分页列表 | [ ] |
| 26 | 资料 | POST | `/admin/api/v1/download-files` | Admin | #16 已发布 | `accessType=free`，`tags` 含「免费」 | [ ] |
| 27 | 资料 | PUT | `/admin/api/v1/download-files/:id` | Admin | #26 | 更新成功 | [ ] |
| 28 | 资料 | DELETE | `/admin/api/v1/download-files/:id` | Admin | 无下载记录或 `confirm=true` | 删除成功 | [ ] |

---

## 七、Admin 运营配置（8）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 29 | 首页 | GET | `/admin/api/v1/home-blocks` | Admin | 登录 | 列表 | [ ] |
| 30 | 首页 | POST | `/admin/api/v1/home-blocks` | Admin | 登录 | 创建入口 | [ ] |
| 31 | 首页 | PUT | `/admin/api/v1/home-blocks/:id` | Admin | #30 | 更新成功 | [ ] |
| 32 | 广告 | GET | `/admin/api/v1/ad-slots` | Admin | 登录 | 列表 | [ ] |
| 33 | 广告 | POST | `/admin/api/v1/ad-slots` | Admin | 登录 | Upsert 成功 | [ ] |
| 34 | 公众号 | GET | `/admin/api/v1/official-account-configs` | Admin | 登录 | 列表 | [ ] |
| 35 | 公众号 | POST | `/admin/api/v1/official-account-configs` | Admin | 登录 | Upsert 成功 | [ ] |
| 36 | 上传 | POST | `/admin/api/v1/upload/sts` | Admin | 登录 | 返回 STS 占位数据 | [ ] |

---

## 八、小程序登录与公开接口（9）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 37 | 登录 | POST | `/api/v1/wx/login` | 无 | — | `code=0`，返回 `userToken` | [ ] |
| 38 | 首页 | GET | `/api/v1/home` | 无 | #16 已发布 | 含四个顶层字段 | [ ] |
| 39 | 分类 | GET | `/api/v1/categories` | 无 | — | 分类树 | [ ] |
| 40 | 文章 | GET | `/api/v1/articles` | 无 | #19 已发布 | 仅已发布 | [ ] |
| 41 | 文章 | GET | `/api/v1/articles/:id` | 无 | #19 | `view_count` 递增 | [ ] |
| 42 | 范文 | GET | `/api/v1/essays` | 无 | #21 | 列表 | [ ] |
| 43 | 范文 | GET | `/api/v1/essays/:id` | 无 | #21 | 详情 | [ ] |
| 44 | 搜索 | GET | `/api/v1/search?keyword=测试` | 无 | — | 返回结果列表 | [ ] |
| 45 | 广告 | GET | `/api/v1/ad-slots` | 无 | #33 | 仅启用项 | [ ] |

---

## 九、小程序下载（2）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 46 | 下载 | GET | `/api/v1/download-files/:id/sign-url` | 可选 | #26 免费资料 | `code=0`，返回 `url` | [ ] |
| 47 | 下载 | POST | `/api/v1/download-files/:id/ad-unlock` | 可选 | gate 已启用 | 返回 `unlockToken` 或 400 | [ ] |

下载边界矩阵见 [`download-permission-boundary.md`](./download-permission-boundary.md)。

---

## 十、小程序埋点（4）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 48 | 埋点 | POST | `/api/v1/track/view` | 可选 | #19 | `code=0` | [ ] |
| 49 | 埋点 | POST | `/api/v1/track/share` | 可选 | #19 | 文章 `share_count` 递增 | [ ] |
| 50 | 埋点 | POST | `/api/v1/track/search` | 可选 | — | `keyword` 必填 | [ ] |
| 51 | 埋点 | POST | `/api/v1/track/ad` | 可选 | #33 | `eventType=expose/click` | [ ] |

---

## 十一、小程序用户中心（6）

| # | 模块 | Method | Path | 鉴权 | 前置 | 关键断言 | 通过 |
|---|------|--------|------|------|------|----------|------|
| 52 | 我的 | GET | `/api/v1/mine` | User | #37 | 三项计数 | [ ] |
| 53 | 收藏 | GET | `/api/v1/favorites` | User | #37 | 分页列表 | [ ] |
| 54 | 收藏 | POST | `/api/v1/favorites` | User | #37,#19 | 收藏成功 | [ ] |
| 55 | 收藏 | DELETE | `/api/v1/favorites` | User | #54 | 取消成功 | [ ] |
| 56 | 历史 | GET | `/api/v1/history` | User | #37,#48 | 分页列表 | [ ] |
| 57 | 下载 | GET | `/api/v1/downloads` | User | #37,#46 | 含 `available` 字段 | [ ] |

---

## 十二、鉴权边界（5）

| # | 场景 | Method | Path | 期望 | 通过 |
|---|------|--------|------|------|------|
| 58 | 无 token 访问用户接口 | GET | `/api/v1/mine` | 401 | [ ] |
| 59 | admin token 访问用户接口 | GET | `/api/v1/favorites` | 401 | [ ] |
| 60 | 无 token 访问 Admin | GET | `/admin/api/v1/articles` | 401 | [ ] |
| 61 | user token 访问 Admin | GET | `/admin/api/v1/stats/overview` | 401 | [ ] |
| 62 | 无 token 埋点仍可写 | POST | `/api/v1/track/view` | 200，`userId=0` | [ ] |

---

## 十三、一期不存在路径审计

以下路径**不应存在**（调用应 404）：

| 路径 | 说明 | 已确认 404 |
|------|------|------------|
| `GET /api/v1/orders` | 订单 | [ ] |
| `POST /api/v1/pay/notify` | 支付回调 | [ ] |
| `GET /admin/api/v1/courses` | 旧课件 | [ ] |
| `GET /api/v1/courses` | 旧课件 | [ ] |

---

## 十四、自动化回归

```bash
cd course_server && go test ./test/...
```

- [ ] 全部测试通过

---

## 验收标准

- [ ] 上表 57 个主接口 + 5 条鉴权边界全部通过
- [ ] 下载权限边界文档矩阵已手工或 Postman 验证
- [ ] 无订单、支付、视频相关可调用接口
- [ ] 后端接口可支撑管理端（Day13+）与小程序端（Day20+）开发
