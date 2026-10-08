# 实操剧本 03：数据库连接池与版本化 SQL 迁移 (Database & Migration Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 2 (Module 03) 阶段基于**方案 A（纯正 MySQL 8.0 生产标准）**从零到一敲下的**每一行命令、代码实现细节、真实避坑排查与最终终端输出**。
> 本手册采用自包含设计，即便脱离项目源码，亦可作为 Go 生产级 MySQL 8.0 连接池管理与 `golang-migrate` 版本化 DDL 迁移落地的实操宝典。

---

## 🎯 一、 核心目标与技术背景

- **核心目标**：
  1. 引入 GORM 核心库、官方 `gorm.io/driver/mysql` 驱动与 `github.com/DATA-DOG/go-sqlmock`；
  2. 编写 `pkg/database/db.go`：支持 MySQL 8.0 连接工厂、连接池调优（`MaxOpenConns`、`MaxIdleConns`、`ConnMaxLifetimeSec`）与 Ping/Close 生命周期；
  3. 编写单元测试 `pkg/database/db_test.go`：基于 sqlmock 模拟底层驱动验证连接池与生命周期；
  4. 编写 `migrations/` 下成对的 DDL 迁移文件（`users` 表与 `articles` 表），严格遵循 MySQL 8.0 InnoDB 标准；
  5. 在 `Makefile` 中集成 `migrate-up` 与 `migrate-down` 自动化迁移命令；
  6. 编写 GORM 实体模型 `internal/model/user.go` 与 `internal/model/article.go`，建立 RBAC 角色与状态机枚举；
  7. 全库通过 `make test-race` 零数据竞争验收。
- **前置环境**：
  - 已完成 Module 00 ~ 02。

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：引入 GORM 与 MySQL 官方驱动依赖
```bash
go get gorm.io/gorm@latest gorm.io/driver/mysql@latest github.com/DATA-DOG/go-sqlmock@latest
go mod tidy
```
- **技术解析**：安装 GORM v1.31+、MySQL v1.6+ 驱动与工业级单测模拟库 `go-sqlmock`。

---

### 步骤 2：编写数据库连接工厂与连接池管理器 `pkg/database/db.go`
```go
// 核心逻辑：显式构造函数传递依赖，调优三项关键连接池参数
func ConfigurePool(sqlDB *sql.DB, cfg configs.DatabaseConfig) {
    if sqlDB == nil {
        return
    }
    if cfg.MaxOpenConns > 0 {
        sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    }
    if cfg.MaxIdleConns > 0 {
        sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    }
    if cfg.ConnMaxLifetimeSec > 0 {
        sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSec) * time.Second)
    }
}
```

---

### 步骤 3：编写版本化 SQL 迁移脚本对 (`migrations/`)
严格遵循 MySQL 8.0 InnoDB 标准建立表结构与索引：
- `000001_create_users_table.up.sql` / `.down.sql`
- `000002_create_articles_table.up.sql` / `.down.sql`

并在 `Makefile` 中封装：
```makefile
migrate-up:
	migrate -path migrations -database "mysql://root:nexus_root_2026@tcp(127.0.0.1:3306)/nexus_hub" up

migrate-down:
	migrate -path migrations -database "mysql://root:nexus_root_2026@tcp(127.0.0.1:3306)/nexus_hub" down 1
```

---

### 步骤 4：编写 GORM 实体模型 (`internal/model/`)
- `internal/model/user.go`：包含 `RoleAdmin`、`RoleEditor`、`RoleViewer` RBAC 角色常量，`PasswordHash` 字段标注 `json:"-"` 安全脱敏；
- `internal/model/article.go`：包含 `draft` / `pending` / `published` / `rejected` / `archived` 全生命周期状态机常量与软删除支持。

---

### 步骤 5：运行并发数据竞争测试
```bash
make test-race
```
- **实际终端输出**：
```text
go test -v -race ./...
?   	nexus-hub/cmd/server	[no test files]
=== RUN   TestLoad_Success
--- PASS: TestLoad_Success (0.00s)
=== RUN   TestValidate_InvalidCases
--- PASS: TestValidate_InvalidCases (0.00s)
PASS
ok  	nexus-hub/configs	(cached)
?   	nexus-hub/internal/model	[no test files]
=== RUN   TestNew_UnsupportedDriver
--- PASS: TestNew_UnsupportedDriver (0.00s)
=== RUN   TestDatabaseLifecycle_WithMock
--- PASS: TestDatabaseLifecycle_WithMock (0.00s)
=== RUN   TestDatabase_NilGuards
--- PASS: TestDatabase_NilGuards (0.00s)
PASS
ok  	nexus-hub/pkg/database	(cached)
=== RUN   TestCode_MsgAndHTTPStatus
--- PASS: TestCode_MsgAndHTTPStatus (0.00s)
=== RUN   TestPageResult_NilSliceGuard
--- PASS: TestPageResult_NilSliceGuard (0.00s)
=== RUN   TestCursorResult_NilSliceGuard
--- PASS: TestCursorResult_NilSliceGuard (0.00s)
=== RUN   TestGinHelpers
--- PASS: TestGinHelpers (0.00s)
PASS
ok  	nexus-hub/pkg/response	(cached)
```

---

## 避坑与故障排查记录 (Troubleshooting)

### 踩坑 1：GORM 内部初始探测与 sqlmock 期望不匹配
- **现象**：单测调用 `NewWithDialector` 时抛出 `call to database Ping was not expected` 导致初始化失败。
- **根因**：GORM 在首次通过 Dialector 初始化连接时，会在内部主动触发一次 `db.DB().Ping()` 探活探测。
- **解决**：在调用 `NewWithDialector` 之前显式注册 `mock.ExpectPing()` 捕获 GORM 内部的探活，在后续调用 `Close(db)` 前显式添加 `mock.ExpectClose()`，满足 sqlmock 的全链路期望。
