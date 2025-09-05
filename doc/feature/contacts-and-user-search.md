# uim-go 联系人和用户搜索技术方案

**文档版本**: 0.1  
**更新时间**: 2026-04-23  
**适用项目**: `uim-go`  
**相关现状**: 当前仅有认证、会话、消息、在线状态接口；无联系人域

---

## Table of Contents

- [1. 背景与目标](#1-背景与目标)
  - [1.1 当前现状](#11-当前现状)
  - [1.2 本次目标](#12-本次目标)
  - [1.3 初期不做](#13-初期不做)
- [2. 与 roadmap-v1.0 的关系](#2-与-roadmap-v10-的关系)
- [3. 方案总览](#3-方案总览)
  - [3.1 Primary owner](#31-primary-owner)
  - [3.2 Likely affected projects](#32-likely-affected-projects)
  - [3.3 Files to inspect first](#33-files-to-inspect-first)
  - [3.4 Verification path](#34-verification-path)
- [4. 数据模型设计](#4-数据模型设计)
  - [4.1 新增表](#41-新增表)
  - [4.2 为什么初期不做 friendships](#42-为什么初期不做-friendships)
- [5. API 设计](#5-api-设计)
  - [5.1 GET /api/contacts](#51-get-apicontacts)
  - [5.2 POST /api/contacts](#52-post-apicontacts)
  - [5.3 GET /api/users/search](#53-get-apiuserssearch)
  - [5.4 可选的后续删除接口](#54-可选的后续删除接口)
- [6. 服务与仓储分层改造](#6-服务与仓储分层改造)
- [7. 鉴权、约束与错误语义](#7-鉴权约束与错误语义)
- [8. 实施步骤](#8-实施步骤)
- [9. 测试与验证](#9-测试与验证)
- [10. 风险与折中](#10-风险与折中)
- [11. 建议提交信息](#11-建议提交信息)

---

## 1. 背景与目标

### 1.1 当前现状

目前 `uim-go` 的 HTTP 路由只覆盖：

- 认证：`/api/auth/*`
- 会话：`/api/conversations*`
- 消息：`/api/conversations/:id/messages`
- 在线状态：`GET /api/users/:id/presence`

现有代码中没有联系人列表、用户搜索、好友关系增删改查等接口；Flutter 端“联系人”页仍是 seed 用户占位页，已经超出当前后端能力边界。

### 1.2 本次目标

为 `uim-go` 增加一个**轻量联系人域**，先满足客户端替换占位页所需的最小能力：

- 浏览我的联系人列表
- 按 `username` 精确搜索用户
- 将某个用户加入我的联系人
- 为联系人页提供可直接展示的基础资料
- 与现有 presence / conversation 创建能力平滑协作

### 1.3 初期不做

为了控制范围，初期明确不做：

- 双向好友关系确认
- 好友请求、黑名单、推荐、分组、备注、置顶
- 通讯录导入
- 全文检索引擎
- 联系人与会话自动强绑定

---

## 2. 与 roadmap-v1.0 的关系

`roadmap-v1.0` 的 Phase 1 明确将 `friendships table` 延后，不属于 MVP；Phase 2 则开始出现 search 类能力。基于这一点，本方案的定位是：

- **不回溯修改 v1.0 MVP 定义**：联系人仍不是原始 MVP 的一部分。
- **作为 v1.x 增量能力实现**：属于对 “Features & Optimization / search” 的前置补齐，而不是重做现有消息主线。
- **不直接落成完整 friendship 系统**：roadmap 里 deferred 的是更完整的社交关系；本方案先落“单向联系人收藏”模型，以更低成本支撑客户端联系人页和搜索入口。

换句话说，这次新增能力与 roadmap **不冲突**，但需要在文档中明确它是 `v1.0` 之后的增量能力，而不是补漏式修改原 MVP 验收口径。

---

## 3. 方案总览

### 3.1 Primary owner

- `uim-go`

### 3.2 Likely affected projects

- `uim-go`
- `uim-flutter`
- `uportal-flutter`（如果后续复用同一联系人页）
- `cloud-server`（仅在需要联调部署时受影响）

### 3.3 Files to inspect first

- `uim-go/internal/api/router.go`
- `uim-go/internal/repository/user_repository.go`
- `uim-go/internal/service/conversation_service.go`
- `uim-go/internal/model/user.go`
- `uim-go/doc/design/v1.0/roadmap-v1.0.md`

### 3.4 Verification path

- 单元测试：repository / service / handler
- 集成测试：联系人列表、用户搜索、添加联系人
- 联调验证：Flutter 联系人页拉取列表、搜索用户、添加后立即可见

---

## 4. 数据模型设计

### 4.1 新增表

建议新增 `user_contacts` 表，表达“某用户收藏了哪些联系人”。

```sql
CREATE TABLE user_contacts (
  owner_user_id UUID NOT NULL,
  contact_user_id UUID NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'manual',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,
  PRIMARY KEY (owner_user_id, contact_user_id),
  CONSTRAINT fk_user_contacts_owner
    FOREIGN KEY (owner_user_id) REFERENCES users(user_id),
  CONSTRAINT fk_user_contacts_contact
    FOREIGN KEY (contact_user_id) REFERENCES users(user_id),
  CONSTRAINT chk_user_contacts_no_self
    CHECK (owner_user_id <> contact_user_id)
);

CREATE INDEX idx_user_contacts_owner_created
  ON user_contacts(owner_user_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_user_contacts_contact
  ON user_contacts(contact_user_id)
  WHERE deleted_at IS NULL;
```

字段说明：

- `owner_user_id`: 联系人列表归属人
- `contact_user_id`: 被加入联系人列表的用户
- `source`: 预留来源，初期固定 `manual`
- `deleted_at`: 支持软删除，便于后续恢复与审计

### 4.2 为什么初期不做 friendships

如果直接引入 `friendships`，通常要同时定义：

- 单向/双向语义
- 请求态与接受态
- 重复申请与并发写入处理
- UI 上的 request / pending / accepted 状态流转

这会把后端和客户端范围从“联系人入口”迅速扩大成“社交关系系统”。对当前目标来说，`user_contacts` 更贴合：

- 能支撑浏览、新增、搜索后的加入
- 与当前 1:1 会话能力低耦合
- 未来若要升级为 friendship，可在其上叠加 `status` 或平滑迁移

---

## 5. API 设计

初期建议新增 3 个核心接口。

### 5.1 GET /api/contacts

用途：列出当前用户的联系人。

请求：

- Header: `Authorization: Bearer <token>`
- Query: `limit`, `offset`

响应建议：

```json
{
  "contacts": [
    {
      "user_id": "uuid",
      "username": "alice",
      "display_name": "Alice",
      "avatar_url": "",
      "created_at": "2026-04-23T12:00:00Z",
      "presence": {
        "status": "online",
        "last_seen": "2026-04-23T11:58:00Z"
      }
    }
  ]
}
```

说明：

- `presence` 建议作为**可选内联字段**返回，减少 Flutter 列表页 N 次 presence 请求。
- 若 `presenceStore == nil`，按现有语义回落到 `offline`。

### 5.2 POST /api/contacts

用途：把某个用户加入我的联系人。

请求体建议：

```json
{
  "contact_user_id": "uuid"
}
```

返回：

- `201 Created`: 新增成功
- `200 OK`: 已存在，返回已有记录也可接受

错误语义：

- `400`: 无效 UUID / 自己添加自己
- `404`: 目标用户不存在
- `409`: 若决定严格区分重复添加，可用冲突；更推荐幂等写入，直接 `200`

### 5.3 GET /api/users/search

用途：提供联系人添加前的用户搜索。

请求：

- Query: `q`, `limit`, `offset`

初期匹配规则建议：

- 仅按 `username` 做精确匹配：`WHERE username = ?`
- 排除当前登录用户
- 不支持 `user_id` 检索
- 不支持 `display_name` 模糊匹配
- 精确匹配前先做 `trim`

响应建议：

```json
{
  "users": [
    {
      "user_id": "uuid",
      "username": "bob",
      "display_name": "Bob",
      "avatar_url": "",
      "already_added": true
    }
  ]
}
```

说明：

- `already_added` 很重要，客户端可以直接禁用“添加”按钮，避免重复操作。
- 初期不建议暴露 `email` 到搜索结果，减少隐私扩散面。
- 当前接口虽然命名为 `search`，但实现语义应保持为“按 username 精确查找”。

### 5.4 可选的后续删除接口

本轮不是必做，但建议在文档中预留：

- `DELETE /api/contacts/:contact_user_id`

如果第一期只实现浏览、新增、搜索，删除可以留到下一轮，不影响总体模型。

---

## 6. 服务与仓储分层改造

建议沿用当前 `api -> service -> repository -> model` 分层：

1. `internal/model/contact.go`
   - 定义 `UserContact`
2. `internal/repository/contact_repository.go`
   - `ListByOwner`
   - `Add`
   - `Exists`
   - `Delete`（可选）
3. `internal/service/contact_service.go`
   - `ListContacts(ownerID, limit, offset)`
   - `AddContact(ownerID, contactUserID)`
   - `SearchUsers(ownerID, query, limit, offset)`
4. `internal/api/contact_handler.go`
   - `List`
   - `Create`
   - `SearchUsers`
5. `internal/api/router.go`
   - 注册 `/api/contacts`
   - 注册 `/api/users/search`

实现要点：

- `SearchUsers` 可先复用 `user_repository`，补充一个 `Search` 方法
- 也可以直接复用 `GetByUsername`，避免为了最简需求引入模糊搜索实现
- `ListContacts` 需要 join `users` 表拿到联系人资料
- 若要内联 `presence`，service 层可注入 `store.PresenceStore`

---

## 7. 鉴权、约束与错误语义

- 所有联系人接口都应复用现有 `AuthMiddleware`
- 服务层强校验 `owner_user_id != contact_user_id`
- `POST /api/contacts` 设计成**幂等**，便于客户端重试
- 搜索接口应限制 `limit` 上限，例如最大 50，避免扫库
- 关键词为空时：
  - 可以返回 `400`
  - 或定义为热门推荐，但当前不建议扩范围
- 联系人关系定义为**单向**：A 添加 B，只生成 `A -> B` 记录，不自动补 `B -> A`
- 联系人关系不作为会话权限前置条件：即使双方不是互为联系人，仍可按现有 1:1 会话能力交流

---

## 8. 实施步骤

建议按 4 个小阶段推进：

1. **数据库与模型**
   - 新增 migration
   - 新增 `UserContact` model
2. **仓储与服务**
   - repository + service + user search
3. **API 与文档**
   - handler、router、feature doc 更新
4. **测试与联调**
   - integration test
   - 与 `uim-flutter` 联系人页联调

流程如下：

```mermaid
flowchart LR
  A["Search users"] --> B["Add contact"]
  B --> C["List contacts"]
  C --> D["Start conversation from client"]
```

---

## 9. 测试与验证

建议最少覆盖以下场景：

- `GET /api/contacts` 空列表
- `POST /api/contacts` 成功添加
- 重复添加同一联系人
- 添加自己失败
- 添加不存在用户失败
- `GET /api/users/search?q=alice` 返回命中
- `GET /api/users/search` 不返回自己
- 已加入联系人时 `already_added = true`
- A 添加 B 后，B 的联系人列表默认不出现 A
- 非互相联系人场景下，现有 1:1 会话创建与交流流程不受影响

联调重点：

- Flutter 搜索结果点“添加”后，联系人列表立即刷新
- 联系人页进入会话仍复用现有 `POST /api/conversations`
- presence 缺失时不影响联系人列表主流程

---

## 10. 风险与折中

- **隐私风险**: 搜索接口若支持 email 或模糊搜索，信息暴露面会变大；初期只按 `username` 精确查
- **性能风险**: `%keyword%` 在用户量增大后会退化；初期可接受，后续再补 trigram / dedicated index
- **产品语义风险**: “联系人”与“好友”容易混淆；文档和接口命名都应坚持 contact，不叫 friend
- **可用性风险**: 只支持 `username` 精确搜索，对用户名记忆要求更高；这是当前“保持最简”的取舍
- **接口耦合风险**: 如果列表页继续逐个调用 presence，会放大请求数；因此建议联系人列表直接内联 presence

---

## 11. 建议提交信息

按逻辑拆分时建议：

- `docs: add uim-go contacts and user search design`
- `feat: add contacts data model and api`
- `test: cover contacts and user search flows`

若最终合并成一个提交，建议：

- `feat: add contacts and user search support`
