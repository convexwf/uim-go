# IM 系统设计


## 需要的东西

**基础功能**：

1. **用户注册和登录**：允许用户创建帐户并登录以访问应用。
2. **个人资料管理**：用户可以设置个人资料、头像和状态消息。
3. **好友列表**：用户可以添加、删除和管理联系人，查看在线状态。
4. **即时消息**：实时聊天功能，包括文字消息、表情符号、图片和语音消息。
5. **群组聊天**：允许用户创建和管理群组，与多个联系人进行群聊。
6. **消息通知**：通知用户有新消息，包括声音、震动和推送通知。
7. **消息历史记录**：保存和查看以前的聊天记录。需要有搜索引擎的扩展。让用户能够按时间、联系人或关键词对消息进行排序和筛选。
8. **在线状态**：显示联系人的在线/离线状态。
9. **消息传送确认**：通知发送者消息已被接收和阅读。
10. **加密和安全性**：确保消息和用户数据的隐私和安全。
11. **消息发送后编辑**：允许用户在发送消息后编辑已发送的消息。
12. **消息引用**：允许用户引用其他消息。
13. **消息转发**：允许用户转发消息给其他联系人。
14. **免打扰模式**：允许用户设置免打扰模式，以免打扰。

**扩展功能**：

1. **语音和视频通话**：实现实时语音和视频通话功能。
2. **音频和视频消息留言**：用户可以在不进行实时通话的情况下留下音频和视频消息。
3. **语音消息转录**：自动将语音消息转录为文字，以便用户可以阅读消息而不必听取录音。
4. **文件传输**：允许用户发送和接收文件，如文档、图片和视频。
5. **位置共享**：用户可以分享自己的实时位置或特定位置。
6. **表情符号和贴纸**：提供更多的表情符号、贴纸和表情包。
7. **智能日历集成**：允许用户在应用内共享和协调日程安排。
8. **多平台支持**：支持多个操作系统和设备，如iOS、Android、Web等。
9. **多语言支持**：提供多种语言的本地化支持。
10. **多设备同步**：确保用户的聊天记录和设置在多个设备上同步。
11. **自动翻译**：自动翻译消息，使用户可以与非本地语言的人交流。
12. **消息撤回和删除**：允许用户在一定时间内删除或撤回发送的消息。
13. **主题和定制**：用户可以选择不同的主题和自定义应用外观。
14. **数据同步**：确保用户的聊天历史和设置在不同设备上同步。
15. **thread**：thread是一个会话，一个thread可以有多个message，一个message可以有多个thread。
16. **邮件系统**：用户可以发送邮件给其他用户。
17. **Feed流**：用户可以发布动态，其他用户可以评论和点赞。
18. **消息提醒设置**：用户可以自定义通知铃声、震动模式和通知类型。
19. **投票和决策功能**：使用户能够发起投票和进行决策，例如选择餐厅或电影。
20. **云存储**：允许用户在聊天中共享文件并将其保存到云存储服务。
21. **媒体分享和播放器**：用户可以在应用内分享音乐、视频和链接，同时也能够播放媒体。
22. **快捷回复**：提供一些常用的快捷回复选项，以便用户更快速地回复消息。
23. **应用内搜索**：快速搜索聊天记录、文件和链接。
24. **表情符号和GIF搜索和推荐**：内置表情符号和GIF搜索引擎，让用户能够更轻松地表达情感。

**创新功能**：

1. **AI 功能**：整合人工智能，如聊天机器人、自动回复和智能搜索。
2. **虚拟现实沟通**：实现虚拟现实环境中的沟通和互动。
3. **区块链认证**：使用区块链技术确保消息的真实性和来源。
4. **社交媒体整合**：连接其他社交媒体平台，允许用户在应用内分享内容。
5. **AR 交互**：使用增强现实技术进行有趣的互动和游戏。
6. **身份验证和生物识别**：更强的用户身份验证和生物识别功能，如指纹和面部识别。
7. **虚拟助手**：提供虚拟助手来执行任务，如预订餐厅或购物。
8. **情感分析**：分析用户消息的情感，并提供情感支持或建议。
9. **即时支付**：允许用户在应用内发送和接收资金。
10. **社交游戏**：整合社交游戏，让用户在聊天中玩游戏。比如在聊天中玩象棋、扑克牌、五子棋或其他游戏。
11. **社交音乐合作**：允许用户在应用中共同创作音乐、演奏音乐或进行音乐表演。

## 1.0 计划

实现 IM 系统的基础功能，包括：

1. OAuth2 认证
2. rust 服务端
3. react 网页前端
4. websocket 通信
5. 实现基本的 IM 功能，包括即时消息、好友列表、在线状态、消息传送确认

## 消息类型

```proto
syntax = "proto3";

message ChatMessage {
  string message_id = 1;
  string sender_id = 2;
  string receiver_id = 3;
  string content = 4;
  MessageType type = 5;
  int64 timestamp = 6;
  MessageStatus status = 7;
  map<string, string> additional_info = 8;
  string stream_id = 9;
  MessagePriority priority = 10;
}

enum MessageType {
  TEXT = 0;
  IMAGE = 1;
  VOICE = 2;
  // ...
}

enum MessageStatus {
  SENT = 0;
  RECEIVED = 1;
  READ = 2;
  // ...
}

enum MessagePriority {
  NORMAL = 0;
  URGENT = 1;
  // ...
}
```

- 消息 ID：每条消息的唯一标识符，可以方便地进行查找和管理。
- 发送者 ID：发送消息的用户 ID。
- 接收者 ID：接收消息的用户 ID。
- 消息内容：消息的具体内容，可能是文本、图片、语音等。
- 消息类型：消息的类型，例如文本消息、图片消息、语音消息等。
- 消息时间戳：消息发送的时间戳，用于排序和显示。
- 消息状态：消息的状态，例如已发送、已接收、已读等。
- 消息附加信息：消息的附加信息，例如发送方和接收方的设备信息、发送方和接收方的位置信息等。
- 消息流 ID：用于标识消息所属的消息流，例如一个会话的 ID。
- 消息优先级：用于标识消息的优先级，例如紧急消息和普通消息。

[gRPC从入门到放弃之好家伙，双向流! | 朝·闻·道](http://wuwenliang.net/2022/03/17/gRPC%E4%BB%8E%E5%85%A5%E9%97%A8%E5%88%B0%E6%94%BE%E5%BC%83%E4%B9%8B%E5%A5%BD%E5%AE%B6%E4%BC%99%EF%BC%8C%E5%8F%8C%E5%90%91%E6%B5%81/)

## Referecne

[Rust的GRPC实现Tonic - 张小凯的博客](https://jasonkayzk.github.io/2022/12/03/Rust%E7%9A%84GRPC%E5%AE%9E%E7%8E%B0Tonic/)
[A simple demo: The basics of gRPC using a Flutter client and a Rust server | by Matthäus Pordzik | Medium](https://medium.com/@matthaeus.pordzik/a-simple-demo-the-basics-of-grpc-using-a-flutter-client-and-a-rust-server-cba3fb736f59)
[Flutter Windows安装、使用ProtoBuf - 简书](https://www.jianshu.com/p/830fa4b6933c)
[Mattiusz/flutter-grpc-client-demo: A simple implementation of a gRPC client using unary and streaming service methods in Flutter](https://github.com/Mattiusz/flutter-grpc-client-demo)
[woodylan/go-websocket: 基于Golang实现的分布式WebSocket服务、IM服务，仅依赖Etcd，简单易部署，支持高并发、单发、群发、广播，其它项目可以通过http与本项目通信。](https://github.com/woodylan/go-websocket)

### 科普

- [知识科普：IM聊天应用是如何将消息发送给对方的？（非技术篇）](http://www.52im.net/thread-2433-1-1.html)


# How to Design a Websocket-based Chat Server

Description: Inspired by [ByteByteGo | design-a-chat-system](https://bytebytego.com/courses/system-design-interview/design-a-chat-system), this project is a simple implementation of a chat server using Websockets.

## Reference

- [ByteByteGo | design-a-chat-system](https://bytebytego.com/courses/system-design-interview/design-a-chat-system)
- [tinode/chat: Instant messaging platform. Backend in Go. Clients: Swift iOS, Java Android, JS webapp, scriptable command line; chatbots](https://github.com/tinode/chat)
- [tinode/webapp: Tinode web chat using React](https://github.com/tinode/webapp/)
- [Actor model 的理解与 protoactor-go 的分析 - 机智的小小帅 - 博客园](https://www.cnblogs.com/XiaoXiaoShuai-/p/16001285.html)


postgreSQL
gin
gorm

## Reference

- [woodylan/go-websocket](https://github.com/woodylan/go-websocket/tree/master)
- [初识 WebSocket 以及 Golang 实现一、 WebSocket 介绍 1.1 WebSocket 的诞生背景 - 掘金](https://juejin.cn/post/7141311208451211278)
- [go - create unit test for ws in golang - Stack Overflow](https://stackoverflow.com/questions/47637308/create-unit-test-for-ws-in-golang)
https://www.reddit.com/r/golang/comments/ndoyan/testing_for_websockets_using_gorilla_and_gingonic/
http://123.56.139.157:8082/article/15/400298/detail.html
https://stackoverflow.com/questions/47637308/create-unit-test-for-ws-in-golang
https://medium.com/@abhishekranjandev/building-a-production-grade-websocket-for-notifications-with-golang-and-gin-a-detailed-guide-5b676dcfbd5a#id_token=eyJhbGciOiJSUzI1NiIsImtpZCI6ImE0OTM5MWJmNTJiNThjMWQ1NjAyNTVjMmYyYTA0ZTU5ZTIyYTdiNjUiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiIyMTYyOTYwMzU4MzQtazFrNnFlMDYwczJ0cDJhMmphbTRsamRjbXMwMHN0dGcuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiIyMTYyOTYwMzU4MzQtazFrNnFlMDYwczJ0cDJhMmphbTRsamRjbXMwMHN0dGcuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDE4MTE1NzA4ODM3MzM1Mzc0NTgiLCJoZCI6ImF1dG94LmFpIiwiZW1haWwiOiJ3ZWlmZW5ncWl1QGF1dG94LmFpIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5iZiI6MTcyNDIxNjQzMSwibmFtZSI6IldlaWZlbmcgUWl1IiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hL0FDZzhvY0lka29kRmFzV1ZmRVRqOE1UUTJRNy0zLUF5LXdDcFEwSVE2Q2F5NEJCc1ZDd195dz1zOTYtYyIsImdpdmVuX25hbWUiOiJXZWlmZW5nIiwiZmFtaWx5X25hbWUiOiJRaXUiLCJpYXQiOjE3MjQyMTY3MzEsImV4cCI6MTcyNDIyMDMzMSwianRpIjoiYTk1MTQ1MmRlMTc5ZmExZjY5YWE4OWFlMjk0ZDQyOGFjZTk4NzUwZiJ9.cBS5fySyQhEM8H6_CvC2PYm-DuAbT4DLjicDMHsh6qjLXlWdnryoQ7-jnuPpOwsCL0aYm83ndHYNrv0n4MpjRgbX7OmH2JZ7TkK46EZOH4Lu1LVBLmZD7cPvX3i1UGdoyOD_PRvQHuL_GP1_RWDt0iGO5xrek1EkbM_jX4PUdPU0Oz_Z73uU40mIvJey8QrrJxqVx791POt4QWErgYj7xn5ZxVqnsvoO6XCEdMN5Ue0vM7fVBOIqn8OVmA6OO07ukxKQHOnpTwCIo4-wVC8clz5aPoZ_Isga6GLiPBZEUMnqZQY3ivyx4RTWgwaXdaQJzdC9Hga_Wz8g9mEmSJhk1g
https://github.com/woodylan/go-websocket/blob/master/docs/api.md
https://golangbot.com/go-websocket-server/
https://www.levenx.com/issues/what-is-the-difference-between-lucene-and-elasticsearch-4qy2zf
