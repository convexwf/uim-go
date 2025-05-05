# UIM Flutter 方案选型说明

**文档版本：** 1.0  
**最后更新：** 2026-02-24  
**作者：** convexwf@gmail.com  
**关联文档：** [UIM Flutter 客户端重构说明](uim-flutter-refactor.zh-cn.md)

本文档为 uim-flutter 重构实施过程中的**技术方案选型备份**，记录各决策点的选定方案及选用理由，便于后续维护与审计。

---

## 目录

- [1. 文档目的](#1-文档目的)
- [2. 选型汇总](#2-选型汇总)
- [3. 状态管理](#3-状态管理)
- [4. HTTP 客户端](#4-http-客户端)
- [5. Web 端 Token 存储](#5-web-端-token-存储)
- [6. WebSocket 库](#6-websocket-库)
- [7. Web 端 IndexedDB 封装](#7-web-端-indexeddb-封装)
- [8. JSON 序列化与模型](#8-json-序列化与模型)
- [9. 实施顺序](#9-实施顺序)
- [参考资料](#参考资料)

---

## 1. 文档目的

- 将「Flutter 重构实施计划」中的**拍板结果**固化为设计文档，避免选型依据丢失。
- 新成员或后续迭代时可据此理解「为何用 A 而非 B」。
- 若需调整某选型，应在本文档中更新并注明原因与日期。

---

## 2. 选型汇总

| 决策项 | 选定方案 | 说明 |
|--------|----------|------|
| 状态管理 | Provider | 官方推荐、结构清晰、与规范兼容，便于 Web/Native 共用 |
| HTTP 客户端 | dio | 拦截器统一处理 Token、401 刷新与重试，联调友好 |
| Web Token 存储 | 统一接口 + 两实现 | Native 用 flutter_secure_storage，Web 用 localStorage/sessionStorage 等，与存储抽象思路一致 |
| WebSocket | web_socket_channel | Flutter 团队维护、多平台支持，满足规范协议即可 |
| Web IndexedDB | indexed_db | 与浏览器 IndexedDB 对应清晰，薄封装，由业务层实现存储接口 |
| JSON / 模型 | json_serializable | 与 uim-go 字段（snake_case）一一对应，可维护性好 |
| 实施顺序 | 严格按阶段 1→2→3→4 | 与重构说明文档一致，依赖关系清晰，便于验收 |

---

## 3. 状态管理

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. 保留 GetX** | 沿用现有 GetX（Controller + 响应式） | 无需重写现有 Controller；学习成本为零 | 社区争议大；与官方推荐不一致；大项目易乱；与「存储抽象」等分层风格难统一 | 否 |
| **B. Provider** | 使用 `provider` + `ChangeNotifier` | 官方推荐、简单、易测、无强约定；与规范兼容；便于 Web/Native 共用 | 需重写当前 GetX 为 ChangeNotifier；异步与错误需自行封装 | **是** |
| **C. Riverpod** | 使用 `flutter_riverpod` | 强类型、可测试性好、无 BuildContext 依赖、适合多平台 | 学习曲线略陡；需重写状态层；当前规模下略重 | 否 |

**选定方案：** B. Provider（`provider` + `ChangeNotifier`）

**选用理由：** 与重构说明兼容且不引入额外强约定；官方文档与生态成熟，团队易上手；结构清晰，后续若需更强类型可再评估 Riverpod。

---

## 4. HTTP 客户端

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. dio** | 使用 `dio` 发 REST | 拦截器可统一加 Token、捕获 401、refresh、重试；取消请求、超时、FormData；联调调试友好 | 依赖体积略大 | **是** |
| **B. http** | 使用 `http` 包 | 轻量、无额外能力 | 无拦截器，Token 注入与 refresh 需在每个调用点手写，易遗漏、难统一 | 否 |

**选定方案：** A. dio

**选用理由：** 规范要求统一注入 Token 与 401 时 refresh 再重试，dio 拦截器可集中实现，与 uim-go 联调更清晰。

---

## 5. Web 端 Token 存储

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. 平台分支：Native secure_storage，Web shared_preferences** | Native 用 `flutter_secure_storage`；Web 用 `shared_preferences_web` 等 | 实现简单；Web 能持久化 | Web 端 XSS 时 Token 可能被读；非真安全存储 | 否 |
| **B. 平台分支：Native secure_storage，Web 仅内存** | Web 不持久化，刷新即需重新登录 | 实现简单；Web 不落盘更安全 | 联调时每次刷新都要登录，体验差 | 否 |
| **C. 统一接口 + 两实现** | 抽象 `TokenStorage`；Native 实现用 secure_storage，Web 实现用 localStorage/sessionStorage | 业务层统一；平台差异收敛在实现层；与「存储抽象」思路一致；Web 可注明仅开发/联调用 | 需多写一层抽象与一个 Web 实现 | **是** |

**选定方案：** C. 统一 Token 存储接口 + 两实现（Native：flutter_secure_storage；Web：localStorage / sessionStorage 或等价）

**选用理由：** 与重构说明中统一存储抽象一致；业务层不散落 `kIsWeb` 分支；Native 安全、Web 可仅用于联调并后续单独评估。

---

## 6. WebSocket 库

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. web_socket_channel** | 使用 `web_socket_channel` | Flutter 团队维护、多平台（含 Web）；API 简单；满足规范 send_message/new_message | ping/pong、重连需应用层实现 | **是** |
| **B. 原生 dart:io / dart:html WebSocket** | 按平台分支使用原生 API | 无额外依赖 | 需自行处理平台差异与序列化，代码重复、维护成本高 | 否 |

**选定方案：** A. web_socket_channel

**选用理由：** 与规范「带 Token 连接、收发 JSON」完全匹配，实现成本低，无需更重封装。

---

## 7. Web 端 IndexedDB 封装

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. idb_shim** | 使用 `idb_shim` 等 idb 系包 | 部分场景可复用与 Node/VM 的 API | Dart 3 与 Web 兼容性需确认；维护活跃度与 API 稳定性需核实 | 否 |
| **B. indexed_db** | 使用 `indexed_db`（Dart 对浏览器 IndexedDB 的封装） | 与浏览器 IndexedDB 对应清晰；薄封装；业务层实现存储接口即可 | API 偏底层，需自封装异步方法 | **是** |
| **C. 手写 dart:html / package:web** | 直接用 `dart:html` 或 `package:web` 的 IndexedDB | 无第三方依赖 | 需手写事务与 ObjectStore，易出错、维护成本高 | 否 |

**选定方案：** B. indexed_db

**选用理由：** 规范要求 Web 直接使用 IndexedDB 并实现同一套存储接口；indexed_db 作为薄封装足够，schema 与 uim-go 字段对齐在实现层完成即可。

---

## 8. JSON 序列化与模型

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. json_serializable** | 使用 `json_serializable` + `@JsonKey(name: 'snake_case')` | 与 uim-go 字段一一对应；生成代码可读；字段变更需重新生成即可 | 需 build_runner；模型变更要跑 codegen | **是** |
| **B. freezed + json_serializable** | 模型用 `freezed`，JSON 用 `json_serializable` | 不可变模型、copyWith、union、类型安全 | 依赖与 codegen 更多；当前规模略重 | 否 |
| **C. 手写 fromJson / toJson** | 手写 Map 转换 | 无 codegen | 易错、字段变更易漏改 | 否 |

**选定方案：** A. json_serializable

**选用理由：** 规范要求与 uim-go 字段一致（snake_case），codegen 可避免手写遗漏；当前规模不引入 freezed，保持简单。

---

## 9. 实施顺序

| 方案 | 说明 | 优点 | 缺点 | 选用 |
|------|------|------|------|------|
| **A. 严格按阶段 1→2→3→4** | 基础（去 Hive、抽象+Isar+IndexedDB、HTTP）→ 认证 → 会话与消息 → UI | 与重构说明一致，便于验收与排期；依赖关系清晰，减少返工 | 阶段 1 完成前无法联调接口 | **是** |
| **B. 先 API + 认证再接存储** | 先实现 dio + 登录/注册/刷新 + Token，再做存储抽象与两实现 | 尽早打通「登录 → 调接口」，联调体验好 | 与文档阶段描述不一致，需单独标注例外 | 否 |

**选定方案：** A. 严格按重构说明的阶段 1 → 2 → 3 → 4 执行

**选用理由：** 与 [UIM Flutter 客户端重构说明 – 6. 实施阶段](uim-flutter-refactor.zh-cn.md#6-实施阶段) 一致，步骤可追溯、便于验收。

---

## 参考资料

- [UIM Flutter 客户端重构说明](uim-flutter-refactor.zh-cn.md) – 原则、API、存储、Seed、阶段
- [初始化](../feature/initialization.md) – uim-go 认证与 API
- [核心消息](../feature/core-messaging.md) – uim-go 会话、消息与 WebSocket
