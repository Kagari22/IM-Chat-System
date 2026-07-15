# IM Chat System

A Go-based instant messaging demo focused on integrating common backend infrastructure into a small but runnable chat system.

This project currently includes:

- MySQL for user and message persistence
- Redis for presence, unread counts, token blacklist, and rate limiting
- RabbitMQ for asynchronous message delivery
- MinIO for file and image storage
- Elasticsearch for message search
- WebSocket for real-time chat

## Features

- User registration and login
- JWT-based authentication
- One-to-one chat
- Real-time messaging over WebSocket
- Offline message retrieval
- Unread message counters
- File upload and image preview
- Message keyword search
- Scrollable chat window for long conversations

## Tech Stack

- Go
- MySQL
- Redis
- RabbitMQ
- MinIO
- Elasticsearch
- Docker Compose
- Plain HTML / JavaScript frontend

## Architecture Overview

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

## Project Structure

```text
cmd/server                    application entrypoint
internal/app                  dependency wiring
internal/auth                 JWT auth helpers
internal/chat                 WebSocket hub
internal/config               environment config
internal/handler              HTTP handlers
internal/httpx                HTTP response helpers
internal/model                domain models
internal/mq                   RabbitMQ publisher / consumer
internal/presence/redis       online presence
internal/ratelimit/redis      simple rate limiting
internal/repository           repository interfaces
internal/repository/mysql     MySQL implementations
internal/search/elasticsearch Elasticsearch search implementation
internal/service              business logic
internal/storage/minio        MinIO uploader
internal/tokenblacklist/redis token blacklist
internal/unread/redis         unread counters
db/schema.sql                 database schema
web/index.html                frontend page
scripts                       local development scripts
```

## Quick Start

### Prerequisites

- Go 1.22 or newer
- Docker Desktop
- PowerShell on Windows

### Start Everything

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
.\scripts\Start-Dev.ps1 -WithInfra
```

This command:

- starts MySQL, Redis, RabbitMQ, MinIO, and Elasticsearch
- loads local development environment variables
- runs the Go server

Then open:

- [http://127.0.0.1:8080](http://127.0.0.1:8080)

## Local Development

### Start Only the App

If the infrastructure is already running:

```powershell
.\scripts\Start-Dev.ps1
```

### Load Environment Variables Manually

If you want to run `go run` yourself:

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
. .\scripts\Set-DevEnv.ps1
go run .\cmd\server
```

### Reset the Database

```powershell
.\scripts\Reset-Db.ps1
```

This will recreate the `im_chat` database and reapply [db/schema.sql](C:/Users/pc/Documents/Codex/2026-07-14/new-chat/work/im-chat-phase1/db/schema.sql).

## Default Ports

To avoid conflicts with common local services, this project uses non-default host ports:

- App: `127.0.0.1:8080`
- MySQL: `127.0.0.1:13306`
- Redis: `127.0.0.1:16379`
- RabbitMQ AMQP: `127.0.0.1:15673`
- RabbitMQ UI: `http://127.0.0.1:15674`
- MinIO API: `127.0.0.1:19000`
- MinIO Console: `http://127.0.0.1:19001`
- Elasticsearch: `http://127.0.0.1:19200`

## Environment Variables

The development scripts currently load these defaults:

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

Recommended local Go cache settings:

```powershell
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
```

## How the Infrastructure Is Used

### MySQL

- stores users
- stores text messages
- stores file and image message metadata

### Redis

- tracks online status
- stores unread message counters
- stores blacklisted tokens
- supports simple message rate limiting

### RabbitMQ

- publishes `message.created` after a message is saved
- consumes events asynchronously
- drives async real-time delivery

### MinIO

- stores uploaded files and images
- returns object URLs
- supports image preview in chat

### Elasticsearch

- indexes messages for search
- provides keyword search API
- supports substring-like search scenarios such as searching `123` in `1234`

## API Overview

### Auth

- `POST /api/register`
- `POST /api/login`
- `GET /api/me`

### User and Message APIs

- `GET /api/users`
- `GET /api/messages?peer_id=2`
- `GET /api/offline`

### Search

- `GET /api/search/messages?q=keyword&peer_id=2`

### Media

- `POST /api/media/upload`

### WebSocket

- `GET /ws?token=...`

## Frontend

The web page currently supports:

- registration and login
- selecting a chat peer
- sending text messages
- uploading files
- previewing images
- searching messages in the current conversation
- restoring the full conversation after search

## Common Issues

### MySQL access denied on `localhost:3306`

If you see:

```text
Error 1045 (28000): Access denied for user 'root'@'localhost'
```

the usual cause is that development environment variables were not loaded, so the app fell back to another local MySQL instance on `3306`.

Use:

```powershell
.\scripts\Start-Dev.ps1
```

or:

```powershell
. .\scripts\Set-DevEnv.ps1
go run .\cmd\server
```

### Database column mismatch

If you see:

```text
Unknown column 'object_key' in 'field list'
```

reset the database:

```powershell
.\scripts\Reset-Db.ps1
```

### Username already exists

If you see:

```text
Duplicate entry 'admin' for key 'users.username'
```

that means the username is already in the database. Use another username or reset the database.

### Docker port conflicts

Do not switch this project back to default ports like `3306`, `6379`, or `5672` unless you know those ports are free on your machine.

## Testing

```powershell
cd C:\Users\pc\Documents\Codex\2026-07-14\new-chat\work\im-chat-phase1
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
go test ./...
```

## Roadmap Ideas

- group chat
- message recall
- read receipts
- file type validation
- highlighted search results
- admin dashboard

## License

This repository currently has no explicit license file. Add one before public distribution if needed.
