# Nexus-Hub 终端常用命令速查手册 (Commands Cheat-Sheet)

> **导读**：
> 本手册汇总了 `nexus-hub` 项目在开发、编译、数据库迁移、依赖装配、测试及容器运维全流程中使用的**所有核心命令行语句**。
> 所有复杂长命令均已在根目录的 `Makefile` 中完成封装，支持“原生完整命令”与“Makefile 简写命令”双重对照。

---

## 🛠️ 1. Go 基础工具链与依赖治理

| 功能操作 | Makefile 简写 | 原生完整命令 | 适用场景与说明 |
| :--- | :--- | :--- | :--- |
| **模块初始化** | 无 | `go mod init nexus-hub` | 仅在项目创建之初执行一次 |
| **依赖整理与修剪** | `make tidy` | `go mod tidy` | 增加新依赖或删除无用依赖后同步更新 `go.mod` 与 `go.sum` |
| **依赖下载预热** | `make download` | `go mod download` | CI/CD 构建前预热本地依赖缓存 |
| **本地极速编译** | `make build` | `go build -o bin/server ./cmd/server` | 编译生成二进制执行文件到 `bin/` 目录 |
| **本地直接启动** | `make run` | `go run ./cmd/server/main.go` | 本地开发时快速单步调试运行服务 |
| **清理编译产物** | `make clean` | `rm -rf bin/ coverage.out` | 清理历史二进制与测试覆盖率文件 |

---

## 🐳 2. Docker 基础设施编排 (MySQL, Redis, ETCD, MinIO)

| 功能操作 | Makefile 简写 | 原生完整命令 | 适用场景与说明 |
| :--- | :--- | :--- | :--- |
| **一键后台启动所有容器** | `make docker-up` | `docker compose up -d` | 30 秒在本地拉起全套生产级中间件 |
| **查看容器运行与健康状态** | `make docker-status` | `docker compose ps` | 查看 MySQL、Redis、ETCD、MinIO 端口与运行状态 |
| **查看特定容器运行日志** | 无 | `docker compose logs -f <service_name>` | 查看指定中间件（如 mysql 或 redis）的标准输出日志 |
| **一键停止并卸载容器** | `make docker-down` | `docker compose down` | 释放宿主机内存与端口资源 |

---

## 🗄️ 3. 数据库版本化迁移 (golang-migrate)

| 功能操作 | Makefile 简写 | 原生完整命令 | 适用场景与说明 |
| :--- | :--- | :--- | :--- |
| **安装迁移 CLI 工具** | 无 | `go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` | 初次使用时安装全局工具 |
| **创建新迁移脚本对** | 无 | `migrate create -ext sql -dir migrations -seq <migration_name>` | 生成成对的 `00000x_xxx.up.sql` 与 `00000x_xxx.down.sql` |
| **执行全量增量迁移** | `make migrate-up` | `migrate -path migrations -database "mysql://root:nexus_root_2026@tcp(127.0.0.1:3306)/nexus_hub" up` | 自动化将表结构演进至最新版本 |
| **安全回滚上一个版本** | `make migrate-down` | `migrate -path migrations -database "mysql://root:nexus_root_2026@tcp(127.0.0.1:3306)/nexus_hub" down 1` | 生产紧急回滚表结构 |

---

## ⚡ 4. 依赖注入与 API 契约代码生成

| 功能操作 | Makefile 简写 | 原生完整命令 | 适用场景与说明 |
| :--- | :--- | :--- | :--- |
| **安装 Wire 编译工具** | 无 | `go install github.com/google/wire/cmd/wire@latest` | 初次使用时安装 Google Wire CLI |
| **编译期依赖自动组装** | `make wire` | `wire ./cmd/server` | 静态分析 Provider 依赖图谱并自动生成 `wire_gen.go` |
| **生成 OpenAPI/Swagger 契约** | `make swagger` | `swag init -g cmd/server/main.go -o api/openapi` | 从 Handler 注释一键生成标准 OpenAPI 3.0 / Swagger JSON |

---

## 🧪 5. 测试、并发竞争检测与质量分析

| 功能操作 | Makefile 简写 | 原生完整命令 | 适用场景与说明 |
| :--- | :--- | :--- | :--- |
| **全库单元测试** | `make test` | `go test -v ./...` | 跑通全库所有单测并打印执行明细 |
| **并发数据竞争检测 (强制验收)** | `make test-race` | `go test -v -race ./...` | **DoD 强制门禁**：基于 Go ThreadSanitizer 扫描 100% 零数据竞争 |
| **生成代码测试覆盖率报告** | `make cover` | `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html` | 本地生成 HTML 可视化测试覆盖率报表 |
