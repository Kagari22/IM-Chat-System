$root = Split-Path -Parent $PSScriptRoot

$env:IM_ADDR = ":8080"
$env:IM_NODE_ID = "node-1"
$env:IM_JWT_SECRET = "change-me"
$env:IM_TOKEN_TTL_HOURS = "168"

$env:IM_MYSQL_DSN = "root:123456@tcp(127.0.0.1:13306)/im_chat?parseTime=true&charset=utf8mb4&loc=Local"

$env:IM_REDIS_ADDR = "127.0.0.1:16379"
$env:IM_REDIS_PASSWORD = ""
$env:IM_REDIS_DB = "0"

$env:IM_ENABLE_RABBITMQ = "true"
$env:IM_RABBITMQ_URL = "amqp://guest:guest@127.0.0.1:15673/"

$env:IM_ENABLE_MINIO = "true"
$env:IM_MINIO_ENDPOINT = "127.0.0.1:19000"
$env:IM_MINIO_ACCESS_KEY = "minioadmin"
$env:IM_MINIO_SECRET_KEY = "minioadmin"
$env:IM_MINIO_BUCKET = "im-chat"
$env:IM_MINIO_USE_SSL = "false"
$env:IM_MINIO_PUBLIC_BASE_URL = "http://127.0.0.1:19000"

$env:IM_ENABLE_ELASTICSEARCH = "true"
$env:IM_ELASTICSEARCH_URL = "http://127.0.0.1:19200"
$env:IM_ELASTICSEARCH_INDEX = "messages"

$env:GOCACHE = Join-Path $root ".gocache"
$env:GOMODCACHE = Join-Path $root ".gomodcache"

Write-Host "Development environment variables loaded."
