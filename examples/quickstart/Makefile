SHELL := /bin/sh

.PHONY: help up down ps logs run run-http run-gin env bot-up bot-logs restart open-pg open-redis webhook-test seed clean

help:
	@echo "QuickStart 常用命令:"
	@echo "  make up        # 启动依赖(Postgres/Redis/面板/可选botserver)"
	@echo "  make down      # 停止并移除容器"
	@echo "  make ps        # 查看容器状态"
	@echo "  make logs      # 跟随查看所有服务日志"
	@echo "  make run       # 运行基础示例 (examples/quickstart)"
	@echo "  make run-http  # 运行 httpServer 示例 (examples/quickstart-httpserver)"
	@echo "  make run-gin   # 运行 Gin 示例 (examples/quickstart-gin)"
	@echo "  make env       # 复制 .env.example -> .env (若不存在)"
	@echo "  make bot-up    # 启动 botserver 服务"
	@echo "  make bot-logs  # 查看 botserver 日志"
	@echo "  make restart   # 重启所有服务"
	@echo "  make open-pg   # 打开 pgAdmin 面板"
	@echo "  make open-redis# 打开 RedisInsight 面板"
	@echo "  make webhook-test # 快速测试 botserver webhook"
	@echo "  make seed      # 执行初始种子数据写入"
	@echo "  make clean     # 清理数据卷(慎用)"

up:
	docker compose up -d

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f

env:
	@test -f .env || cp .env.example .env
	@echo ".env 已就绪 (如需请自行修改)"

run: env
	REDIS_URI=$${REDIS_URI:-127.0.0.1:6379?db=1} \
	PG_URL=$${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
	go run . --listen :8080

run-http: env
	REDIS_URI=$${REDIS_URI:-127.0.0.1:6379?db=1} \
	PG_URL=$${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
	go run ../quickstart-httpserver --listen :8081

run-gin: env
	REDIS_URI=$${REDIS_URI:-127.0.0.1:6379?db=1} \
	PG_URL=$${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
	go run ../quickstart-gin --listen :8082

bot-up: env
	docker compose up -d botserver

bot-logs:
	docker compose logs -f botserver

restart:
	docker compose restart

open-pg:
	@python3 - <<'PY'
import webbrowser; webbrowser.open('http://localhost:5050')
PY

open-redis:
	@python3 - <<'PY'
import webbrowser; webbrowser.open('http://localhost:5540')
PY

webhook-test:
	@echo "GET webhook:"
	@curl -fsS "http://localhost:8082/webhook?msg=hello" || true
	@echo "\nPOST webhook:"
	@curl -fsS -X POST "http://localhost:8082/webhook" -H 'Content-Type: application/json' -d '{"service":"demo","type":"test","message":"Hello from webhook"}' || true

seed:
	PG_URL=$${PG_URL:-postgresql://dev:123@127.0.0.1:5432/base} \
	go run ./seeds

clean:
	docker compose down -v
	@echo "已清理容器与数据卷"
