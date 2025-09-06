# uim-go 群聊功能技术方案

**文档版本**: 0.1  
**更新时间**: 2026-04-23  
**适用项目**: `uim-go`  
**相关 roadmap**: `doc/design/v1.0/roadmap-v1.0.md` Phase 4

---

## 1. 背景与目标

当前 `uim-go` 的会话主干已经支持：

- 认证
- 1:1 会话创建与列表
- 消息发送、拉取、离线同步
- 在线状态
- 轻量联系人与精确用户搜索

但“群聊”还停留在数据模型预留阶段：

- `Conversation.Type` 已支持 `group`
- `ConversationParticipant.Role` 已支持 `owner / admin / member`
- 现有 service / handler / repository 仍只真正实现了 1:1 语义

本方案目标是补齐**第一阶段可用群聊能力**：

- 创建群聊
- 查看群信息与成员列表
- 添加成员 / 移除成员
- 在现有消息链路上发送群消息
- 区分“删除 1:1 会话”和“离开群聊”的语义

初期不做：

- 群邀请链接 / 邀请码
- 群公告、群置顶消息、群禁言
- 群头像上传
- 超大群优化
- read receipt / typing indicator

---

## 2. 与 roadmap-v1.0 的关系

`roadmap-v1.0` 的 Phase 4 已经把群聊列为下一阶段高优任务，因此这次方案属于：

- 对 roadmap 的正向落实，不是额外插入的新方向
- 在当前 1:1 会话主线稳定后最自然的能力扩展

需要注意的一点是：当前系统已经实现了联系人功能，这部分并不在原 roadmap v1.0 MVP 里；而群聊则是 roadmap 原本就计划中的下一块核心能力。因此从优先级上，群聊应高于消息搜索、typing indicator、read receipt。

---

## 3. 现状与主要差距

### 3.1 已有基础

- `conversations` 表已有 `type`、`name`、`created_by`
- `conversation_participants` 表已有 `role`
- WebSocket fanout 已按 `conversation_id -> participant user_ids` 广播，天然可复用到群聊
- `GET /api/conversations`、`GET /api/conversations/:id/messages` 已具备会话和消息主线

### 3.2 当前缺口

- 没有 `POST /api/conversations/group`
- 没有成员管理 API
- `ListByUserIDWithMeta` 只对 1:1 补 `other_user`
- `DELETE /api/conversations/:id` 当前语义是“整会话硬删除”，不适合群聊
- 没有群信息获取接口

---

## 4. 第一阶段范围定义

建议把群聊第一阶段限定为下面 6 个接口：

- `POST /api/conversations/group`
- `GET /api/conversations/:id`
- `GET /api/conversations/:id/members`
- `POST /api/conversations/:id/members`
- `DELETE /api/conversations/:id/members/:user_id`
- `POST /api/conversations/:id/leave`

同时保留现有：

- `GET /api/conversations`
- `GET /api/conversations/:id/messages`
- WebSocket `send_message`

并调整：

- `DELETE /api/conversations/:id`

第一阶段建议把它收紧为：

- 只允许删除 `one_on_one`
- 若目标是 `group`，返回 `409` 或 `400`，提示使用 `leave`

这样可以避免现有“参与者任意删掉整个会话”的危险语义被直接带入群聊。

---

## 5. 数据模型与迁移建议

### 5.1 现有表是否足够

第一阶段基本足够，不必新开群专属表。

已有字段可直接承载：

- `conversations.type = group`
- `conversations.name` 作为群名称
- `conversations.created_by` 作为群主
- `conversation_participants.role` 作为成员角色

### 5.2 建议补充的约束与索引

建议追加 migration：

1. 为群聊成员查询补索引
2. 为角色变更补约束
3. 如有必要，为群名称补长度限制说明

建议 SQL 方向：

```sql
CREATE INDEX idx_conv_participants_conversation_role
  ON conversation_participants(conversation_id, role);
```

如果当前库里还没有 `joined_at` 默认值，也建议确认落库行为一致。

### 5.3 是否需要独立 group_metadata 表

第一阶段不需要。

原因：

- 当前只需要 `name`
- `description / avatar / settings` 暂未进入范围
- 单表可以减少 API 和 ORM 改造面

后续如果有：

- 群公告
- 群简介
- 群头像
- 多种配置项

再拆 `group_metadata` 更合适。

---

## 6. API 设计

### 6.1 POST /api/conversations/group

用途：创建群聊。

请求体建议：

```json
{
  "name": "项目讨论组",
  "member_user_ids": [
    "uuid-a",
    "uuid-b"
  ]
}
```

规则：

- 创建者自动入群
- `member_user_ids` 去重
- 不允许空群名
- 初期限制成员总数，例如 `2 <= total_members <= 50`

响应：

- `201 Created`
- 返回群会话对象和初始成员摘要

### 6.2 GET /api/conversations/:id

用途：获取会话详情。

对 1:1 和 group 都适用，但 group 需返回更多元信息：

```json
{
  "conversation_id": "uuid",
  "type": "group",
  "name": "项目讨论组",
  "created_by": "uuid",
  "created_at": "...",
  "updated_at": "...",
  "member_count": 4,
  "my_role": "owner"
}
```

### 6.3 GET /api/conversations/:id/members

用途：获取群成员列表。

只允许群成员访问。

响应建议：

```json
{
  "members": [
    {
      "user_id": "uuid",
      "username": "alice",
      "display_name": "Alice",
      "avatar_url": "",
      "role": "owner",
      "joined_at": "..."
    }
  ]
}
```

### 6.4 POST /api/conversations/:id/members

用途：拉人进群。

请求体：

```json
{
  "user_ids": ["uuid-a", "uuid-b"]
}
```

权限建议：

- 第一阶段仅 `owner` 可添加
- `admin` 体系先保留字段，不先开放复杂规则

语义：

- 已在群中的用户跳过
- 不存在的用户返回 `404` 或整体 `400`
- 推荐做“部分幂等，整体失败前校验”

### 6.5 DELETE /api/conversations/:id/members/:user_id

用途：移除群成员。

权限建议：

- `owner` 可移除普通成员
- 不允许移除自己，自己退出用 `leave`
- 第一阶段不开放 owner 转移，所以不允许移除 owner

### 6.6 POST /api/conversations/:id/leave

用途：当前用户退出群聊。

规则建议：

- 普通成员可直接退出
- owner 若群内还有其他成员，第一阶段不允许直接退出，返回错误提示先转移或解散
- owner 且仅剩自己时，可把群会话删除

### 6.7 DELETE /api/conversations/:id

建议重定义为：

- `one_on_one`: 保持现有删除语义
- `group`: 第一阶段不支持，返回错误

否则会与 `leave`、`remove member` 语义冲突。

---

## 7. Service / Repository 改造建议

### 7.1 Service 层新增能力

建议在 `ConversationService` 扩展：

- `CreateGroup(creatorID uuid.UUID, name string, memberIDs []uuid.UUID) (*model.Conversation, error)`
- `GetConversationDetail(conversationID, userID uuid.UUID) (...)`
- `ListMembers(conversationID, userID uuid.UUID) (...)`
- `AddMembers(conversationID, operatorID uuid.UUID, userIDs []uuid.UUID) error`
- `RemoveMember(conversationID, operatorID, targetUserID uuid.UUID) error`
- `LeaveConversation(conversationID, userID uuid.UUID) error`

并新增错误类型：

- `ErrGroupOnly`
- `ErrPermissionDenied`
- `ErrConversationTypeMismatch`
- `ErrInvalidGroupSize`
- `ErrCannotRemoveOwner`
- `ErrOwnerCannotLeave`

### 7.2 Repository 层新增能力

建议补：

- `AddParticipants([]*ConversationParticipant) error`
- `ListParticipants(conversationID uuid.UUID) ([]*ConversationParticipant, error)`
- `GetParticipant(conversationID, userID uuid.UUID) (*ConversationParticipant, error)`
- `DeleteParticipant(conversationID, userID uuid.UUID) error`
- `CountParticipants(conversationID uuid.UUID) (int64, error)`
- `UpdateConversation(...)`

### 7.3 列表元数据兼容策略

`ListByUserIDWithMeta` 需要区分：

- 1:1: 继续返回 `other_user`
- group: 返回 `name`、`member_count`，不再依赖 `other_user`

即列表项的 API shape 需要从“以 1:1 为中心”升级成“兼容 1:1 / group 双态”。

---

## 8. 删除语义与权限语义

这是群聊方案里最需要先定死的部分。

### 8.1 1:1 删除

维持当前简单语义：

- 删除整个会话

### 8.2 group 删除

第一阶段不做“参与者删除整个群”。

建议拆为：

- 退出群：`leave`
- 管理员移除别人：`remove member`
- 解散群：后续再做单独 API，例如 `DELETE /api/conversations/:id/group`

### 8.3 权限最小集

第一阶段建议只落实：

- `owner`：添加成员、移除成员、查看成员
- `member`：查看成员、发送消息、退出群

`admin` 只保留字段，不在第一阶段真的开放权限差异。

---

## 9. Flutter / WebSocket 影响

后端实现时要同步考虑客户端消费形态：

- 列表接口里 group 需要明确 `type=group`
- 群聊消息仍走现有 `send_message`
- `ChatScreen` 不应再假设只有 `otherUserId`
- presence 在 group 聊天页不再是单一顶部状态，而是可不展示

也就是说，后端返回 shape 需要对客户端足够友好，避免客户端为区分 group 再做多次补查。

---

## 10. 分阶段实施建议

### Phase A：后端最小可用群聊

- create group
- list/detail supports group
- list members
- add members
- leave group
- group messages over existing WebSocket

### Phase B：群成员管理完善

- remove member
- owner rule enforcement
- richer group metadata

### Phase C：后续增强

- group settings
- owner transfer
- disband group
- admin role

---

## 11. 测试与验证

后端建议至少补以下测试：

- 创建群聊成功 / 非法成员 / 重复成员
- owner 添加成员
- member 添加成员被拒绝
- 成员退出群
- owner 退出群被拒绝
- 群消息 fanout 到多个连接
- 群列表返回正确 `type/name/member_count`

联调验证建议：

- Flutter 创建群聊
- 进入群聊页发送消息
- 第二个成员实时收到
- 退出群后不再收到群消息

---

## 12. 建议提交边界

建议拆成两步提交：

1. `uim-go` 群聊后端能力与测试
2. `uim-flutter` 群聊 UI 与客户端 API

