.PHONY: help build run tidy test test-race clean docker-up docker-down docker-status wire

# 默认展示帮助指南
help:
	@echo "=========================================================================="
	@echo "                   Nexus-Hub 工程管理效能指令中心 (Makefile)                "
	@echo "=========================================================================="
	@echo "  make run           - 运行服务主入口 (本地极速单步调试)"
	@echo "  make build         - 编译二进制文件至 bin/server"
	@echo "  make tidy          - 整理并修剪 go.mod 与 go.sum 依赖"
	@echo "  make test          - 运行全库所有单元测试"
	@echo "  make test-race     - 运行全库数据竞争并发扫描 (DoD 强制验收)"
	@echo "  make clean         - 清理本地编译产物与测试覆盖率文件"
	@echo "  make docker-up     - 一键在后台启动全套基础设施 (MySQL, Redis, ETCD, MinIO)"
	@echo "  make docker-status - 查看本地容器健康检查与端口映射状态"
	@echo "  make docker-down   - 一键安全停止并卸载所有基础设施容器"
	@echo "  make wire          - 触发 Google Wire 编译期自动依赖静态组装"
	@echo "=========================================================================="

# 编译生成二进制执行文件
build:
	@mkdir -p bin
	@echo ">> 正在编译 nexus-hub 服务..."
	go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=$$(date +'%Y-%m-%d')" -o bin/server ./cmd/server
	@echo ">> 编译完成: bin/server"

# 本地快速运行
run:
	go run ./cmd/server/main.go

# 整理依赖
tidy:
	go mod tidy

# 运行基础测试
test:
	go test -v ./...

# 运行数据竞争排查测试
test-race:
	go test -v -race ./...

# 清理构建产物
clean:
	rm -rf bin coverage.out coverage.html
	@echo ">> 清理完毕"

# 基础设施编排指令
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-status:
	docker compose ps

# Google Wire 代码自动装配生成
wire:
	wire ./cmd/server
