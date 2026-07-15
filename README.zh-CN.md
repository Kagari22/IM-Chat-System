# IM Chat System

一个基于 Go 的即时通讯示例项目，目标是把常见的后端基础设施整合进一个小型但可运行的聊天系统。

当前项目已经接入：

- MySQL：用户和消息持久化
- Redis：在线状态、未读计数、Token 黑名单、限流
- RabbitMQ：异步消息投递
- MinIO：文件和图片存储
- Elasticsearch：消息搜索
- WebSocket：实时通信

## 功能特性

- 用户注册与登录
- 基于 JWT 的身份认证
- 单聊聊天
- 基于 WebSocket 的实时消息收发
- 离线消息拉取
- 未读消息计数
- 文件上传与图片预览
- 消息关键词搜索
- 长会话滚动消息窗口

## 技术栈

- Go
- MySQL
- Redis
- RabbitMQ
- MinIO
- Elasticsearch
- Docker Compose
- 原生 HTML / JavaScript 前端

## 架构概览

```text
Client
  -> HTTP API / WebSocket
  -> Handler Layer
  -> Service Layer
  -> Repository / Infrastructure Layer

Infrastructure:
- MySQL: persistent data
- Redis: state and counters
- RabbitMQ: async events
- MinIO: object storage
- Elasticsearch: message indexing and search
```

## 项目结构

```text
cmd/server                    程序入口
internal/app                  依赖装配
internal/auth                 JWT 鉴权辅助
internal/chat                 WebSocket Hub
internal/config               环境变量配置
internal/handler              HTTP 处理器
internal/httpx                HTTP 响应辅助
internal/model                领域模型
internal/mq                   RabbitMQ 发布与消费
internal/presence/redis       在线状态
internal/ratelimit/redis      简单限流
internal/repository           仓储接口
internal/repository/mysql     MySQL 实现
internal/search/elasticsearch Elasticsearch 搜索实现
internal/service              业务逻辑
internal/storage/minio        MinIO 上传
internal/tokenblacklist/redis Token 黑名单
internal/unread/redis         未读计数
db/schema.sql                 数据库结构脚本
web/index.html                前端页面
scripts                       本地开发脚本
```

## 快速开始

### 前置要求

- Go 1.22 或更高版本
- Docker Desktop
- Windows PowerShell

### 启动完整环境

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
.\scripts\Start-Dev.ps1 -WithInfra
```

这个命令会：

- 启动 MySQL、Redis、RabbitMQ、MinIO、Elasticsearch
- 加载本地开发环境变量
- 运行 Go 服务端

然后打开：

- [http://127.0.0.1:8080](http://127.0.0.1:8080)

## 本地开发

### 只启动应用

如果基础设施已经启动：

```powershell
.\scripts\Start-Dev.ps1
```

### 手动加载环境变量

如果你想自己执行 `go run`：

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
. .\scripts\Set-DevEnv.ps1
go run .\cmd\server
```

### 重置数据库

```powershell
.\scripts\Reset-Db.ps1
```

这个脚本会重建 `im_chat` 数据库，并重新执行 [db/schema.sql](C:/Users/pc/Documents/Codex/2026-07-14/new-chat/work/im-chat-phase1/db/schema.sql)。

## 默认端口

为了避免和本机常见服务冲突，项目使用了非默认主机端口：

- App: `127.0.0.1:8080`
- MySQL: `127.0.0.1:13306`
- Redis: `127.0.0.1:16379`
- RabbitMQ AMQP: `127.0.0.1:15673`
- RabbitMQ UI: `http://127.0.0.1:15674`
- MinIO API: `127.0.0.1:19000`
- MinIO Console: `http://127.0.0.1:19001`
- Elasticsearch: `http://127.0.0.1:19200`

## 环境变量

开发脚本当前会加载以下默认值：

```env
IM_ADDR=:8080
IM_NODE_ID=node-1
IM_JWT_SECRET=change-me
IM_TOKEN_TTL_HOURS=168

IM_MYSQL_DSN=root:123456@tcp(127.0.0.1:13306)/im_chat?parseTime=true&charset=utf8mb4&loc=Local

IM_REDIS_ADDR=127.0.0.1:16379
IM_REDIS_PASSWORD=
IM_REDIS_DB=0

IM_ENABLE_RABBITMQ=true
IM_RABBITMQ_URL=amqp://guest:guest@127.0.0.1:15673/

IM_ENABLE_MINIO=true
IM_MINIO_ENDPOINT=127.0.0.1:19000
IM_MINIO_ACCESS_KEY=minioadmin
IM_MINIO_SECRET_KEY=minioadmin
IM_MINIO_BUCKET=im-chat
IM_MINIO_USE_SSL=false
IM_MINIO_PUBLIC_BASE_URL=http://127.0.0.1:19000

IM_ENABLE_ELASTICSEARCH=true
IM_ELASTICSEARCH_URL=http://127.0.0.1:19200
IM_ELASTICSEARCH_INDEX=messages
```

推荐同时保留本地 Go 缓存配置：

```powershell
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
```

## 基础设施在项目中的作用

### MySQL

- 存储用户
- 存储文本消息
- 存储文件和图片消息元数据

### Redis

- 跟踪在线状态
- 存储未读消息计数
- 存储被拉黑的 Token
- 支持简单的消息发送限流

### RabbitMQ

- 在消息写入后发布 `message.created` 事件
- 异步消费消息事件
- 驱动异步实时投递

### MinIO

- 存储上传的文件和图片
- 返回对象访问 URL
- 支持聊天中的图片预览

### Elasticsearch

- 为消息建立搜索索引
- 提供关键词搜索 API
- 支持类似在 `1234` 中搜索 `123` 的子串场景

## API 概览

### 认证

- `POST /api/register`
- `POST /api/login`
- `GET /api/me`

### 用户与消息接口

- `GET /api/users`
- `GET /api/messages?peer_id=2`
- `GET /api/offline`

### 搜索

- `GET /api/search/messages?q=keyword&peer_id=2`

### 媒体上传

- `POST /api/media/upload`

### WebSocket

- `GET /ws?token=...`

## 前端能力

当前页面支持：

- 注册和登录
- 选择聊天对象
- 发送文本消息
- 上传文件
- 预览图片
- 搜索当前会话消息
- 搜索后恢复完整会话

## 常见问题

### `localhost:3306` 上的 MySQL 认证失败

如果你看到：

```text
Error 1045 (28000): Access denied for user 'root'@'localhost'
```

通常原因是开发环境变量没有加载，应用退回连接到了你本机另一个运行在 `3306` 的 MySQL。

请使用：

```powershell
.\scripts\Start-Dev.ps1
```

或者：

```powershell
. .\scripts\Set-DevEnv.ps1
go run .\cmd\server
```

### 数据库字段不匹配

如果你看到：

```text
Unknown column 'object_key' in 'field list'
```

请重置数据库：

```powershell
.\scripts\Reset-Db.ps1
```

### 用户名已存在

如果你看到：

```text
Duplicate entry 'admin' for key 'users.username'
```

说明这个用户名已经在数据库里存在了。换一个用户名，或者重置数据库。

### Docker 端口冲突

除非你确认本机端口空闲，否则不要把项目改回默认端口 `3306`、`6379`、`5672`。

## 测试

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
go test ./...
```

## 后续可扩展方向

- 群聊
- 消息撤回
- 已读回执
- 文件类型校验
- 搜索结果高亮
- 后台管理面板

## License

当前仓库还没有显式的许可证文件。如果要公开分发，建议补充合适的 License。
