#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)

usage() {
  cat <<EOF
QuickStart 一键脚本

用法:
  $0 up            # 启动依赖容器(Postgres/Redis/面板/可选botserver)
  $0 down          # 停止并移除容器
  $0 run           # 运行基础示例 (标准 http.Server)
  $0 run-http      # 运行 httpServer 备用版本
  $0 run-gin       # 运行 Gin 集成版本
  $0 bot-up        # 仅启动 botserver
  $0 webhook-test  # 测试 botserver webhook

环境变量:
  LISTEN_ADDR (:8080), PG_URL, REDIS_URI, BOT_TOKEN, CHAT_ID, BOT_WEBHOOK_SECRET
EOF
}

cmd=${1:-help}
case "$cmd" in
  up)
    (cd "$ROOT_DIR" && docker compose up -d)
    ;;
  down)
    (cd "$ROOT_DIR" && docker compose down)
    ;;
  run)
    REDIS_URI=${REDIS_URI:-127.0.0.1:6379?db=1} \
    PG_URL=${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
    go run "$ROOT_DIR" --listen "${LISTEN_ADDR:-:8080}"
    ;;
  run-http)
    REDIS_URI=${REDIS_URI:-127.0.0.1:6379?db=1} \
    PG_URL=${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
    go run "$ROOT_DIR/../quickstart-httpserver" --listen "${LISTEN_ADDR:-:8081}"
    ;;
  run-gin)
    REDIS_URI=${REDIS_URI:-127.0.0.1:6379?db=1} \
    PG_URL=${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
    go run "$ROOT_DIR/../quickstart-gin" --listen "${LISTEN_ADDR:-:8082}"
    ;;
  bot-up)
    (cd "$ROOT_DIR" && docker compose up -d botserver)
    ;;
  webhook-test)
    curl -fsS "http://localhost:8082/webhook?msg=hello" || true
    ;;
  *)
    usage
    ;;
esac

