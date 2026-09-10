# Nexus-Hub 现代化企业级业务中台系统（Go + Gin + GORM + 微服务基础设施）

本项目是与底层的语言级实验代码库 `lp-go` 严格解耦的**独立企业级商业化高性能业务中台**。

它的诞生旨在建立大厂高标的**现代业务后端工程体系**，打破传统仅掌握“Gin + GORM 简单增删改查”的初级认知，全面引入一线大厂（字节跳动、美团、腾讯、阿里等）的高阶工程实践与技术栈闭环：

1. **数据库标准与深度性能优化**：以 **MySQL 8.0** 为生产基准主库（深度落地 InnoDB 联合索引设计与事务锁机制），通过 GORM 驱动抽象兼顾本地 **SQLite 免安装双模**；采用 **golang-migrate** 进行版本化 SQL 迁移；支持**游标深分页（Keyset Pagination）**，避免海量数据下慢查询 I/O 尖刺。
2. **完整 RBAC 权限与资产状态机**：内置企业级基于角色的权限控制（`admin` / `editor` / `viewer`）；实现严谨的文档全生命周期状态机：`Draft` (草稿) ➔ `Pending` (待审核) ➔ `Published` (发布) / `Rejected` (驳回) ➔ `Archived` (归档)。
3. **接口幂等性防刷体系**：洋葱中间件内置基于 `X-Idempotency-Key` 与 Redis `SetNX` 的幂等性防重锁，彻底消除弱网重试或并发重复提交导致的重复写入。
4. **对象存储与预签名直传（S3/MinIO）**：接入 **MinIO / AWS S3** 协议，支持大厂标准的 **Pre-signed URL 预签名直传**，客户端直接将文件推送到对象存储，节约业务服务器 90% 网卡带宽。
5. **异步任务与业务解耦（Asynq）**：基于 **Asynq / Redis** 实现分布式延时与异步任务队列（敏感词风控审查、阅读量与点赞数批量异步合并刷盘）。
6. **前后端深度协同机制**：支持 **JWT 双 Token（Access + Refresh）无感静默刷新**、跨域标准处理、统一强类型**领域业务错误码字典**，提供开箱即用的 Axios 拦截器与 TypeScript DTO。
7. **架构严谨性与现代依赖注入**：采用 **Google Wire** 编译期无反射依赖注入，杜绝全局变量反模式，代码完全遵循 Clean Architecture（整洁架构）。
8. **分布式配置中心**：整合 **ETCD v3** 租约心跳与 Watch 机制，实现系统业务配置毫秒级热更新无需发版。
9. **一键容器化编排**：提供完整 `docker-compose.yml`，一键拉起 MySQL 8.0、Redis 7.0、ETCD v3、MinIO 等基础设施。

---

## 📖 架构设计规范与工程效能指引

为保障企业级中台研发的高可用性、可维护性与代码严肃性，本项目在根目录设立了严苛的工程规范与演进纪要：

- 📘 **[DEVELOPMENT_CONVENTION.md](./DEVELOPMENT_CONVENTION.md)**：**《企业级工程开发与架构约定规范》**，明确了底层设计决策注释（Why & Rationale）、无全局变量设计准则与单向依赖分层架构；
- 📝 **[DEV_LOG.md](./DEV_LOG.md)**：**《核心模块演进与技术架构实操手册》**，系统化沉淀 10 大核心模块的技术痛点、选型权衡 (Trade-offs)、底层避坑指南与纯命令行自动化验证；
- 🕹️ **[playbook/](./playbook/)**：**《全流程实操剧本与命令手册》**，按模块拆分记录每一步敲下的原生命令、真实避坑排查与终端输出；
 - ⌨️ **[COMMANDS_CHEAT_SHEET.md](./COMMANDS_CHEAT_SHEET.md)**：**《终端常用命令速查手册》**，汇总全流程 Go 工具链、Docker 编排、数据库迁移、Wire 组装与并发测试命令，支持 Makefile 简写对照；
- 🎯 **[QA.md](./QA.md)**：**《高并发与微服务架构核心追问深度解析》**，深度剖析分布式事务、InnoDB 锁机制、缓存双写一致性与编译期依赖注入。

---

## 🏛️ 技术栈全景矩阵（2026 大厂生产高标）

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                   Nexus-Hub 生产级核心技术矩阵                                   │
├───────────────────┬──────────────────────────────────────────────────────────────────────────────┤
│ 1. Web 接入层     │ Gin 框架 + 洋葱中间件 (CORS, JWT, 接口幂等性防重, RBAC 权限守卫, Recovery)   │
│ 2. 依赖注入与架构 │ Google Wire (编译期静态依赖注入，严格遵循 Clean Architecture，无全局单例)    │
│ 3. 核心持久化     │ MySQL 8.0 (主库，联合索引/事务) + GORM + SQLite (本地开箱即用免装双模式)     │
│ 4. 数据库演进性能 │ golang-migrate (版本化 SQL 迁移) + 游标深分页 (Keyset Pagination, O(1) 性能) │
│ 5. 缓存与防击穿   │ Redis 7.x (go-redis v9) + Cache-Aside 双写 + 分布式防重锁 + 点赞高频去重 Set │
│ 6. 对象存储 (OSS) │ MinIO / AWS S3 + 预签名直传 (Pre-signed URL，前端直传不占用后端带宽)         │
│ 7. 异步与延时队列 │ Asynq (基于 Redis 的分布式后台异步与延时任务队列，支持自动重试与死信机制)    │
│ 8. 分布式协调配置 │ ETCD v3 (租约服务保持 + Watch 机制实现业务运行参数毫秒级动态热更新)          │
│ 9. 前端协同契约   │ 强类型领域业务错误码字典 + 统一 ApiResponse + 开箱即用 Axios 双 Token 拦截器 │
│ 10. 安全鉴权体系  │ JWT 双 Token (15分钟 Access Token + 7天 Refresh Token) + Bcrypt 盐值散列      │
│ 11. 运维与环境编排│ Docker Compose 一键编排 (MySQL, Redis, ETCD, MinIO 本地极速开箱)             │
└───────────────────┴──────────────────────────────────────────────────────────────────────────────┘
```

---

## 📂 工程目录与规格划分

```text
nexus-hub/
├── README.md                   # 架构全局纲领与实战指南
├── DEVELOPMENT_CONVENTION.md   # 【工程规范】架构约定与代码设计注释指南
├── DEV_LOG.md                  # 【技术手册】10大功能模块全流程演进与底层避坑实操
├── QA.md                       # 【核心解析】架构师高频 10 大硬核追问与深度原理解析
├── Makefile                    # 效能脚本：wire 编译、migrate 迁移、swagger 生成、本地启动
├── docker-compose.yml          # 本地一键拉起 MySQL, Redis, ETCD, MinIO
├── go.mod                      # 独立 Go 模块 (module nexus-hub)
│
├── cmd/
│   └── server/
│       ├── main.go             # 服务主入口（监听系统信号、平滑停机、ETCD 注册）
│       ├── wire.go             # Google Wire 注入声明
│       └── wire_gen.go         # Wire 编译期自动生成的真实依赖装配代码
│
├── configs/
│   ├── config.go               # 强类型配置结构体（支持环境变量与校验）
│   └── config.yaml             # 运行环境配置参数（数据库 DSN、Redis、S3/MinIO、ETCD）
│
├── migrations/                 # 数据库版本迁移脚本（严格 MySQL 8.0 DDL）
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_articles_table.up.sql
│   └── 000002_create_articles_table.down.sql
│
├── internal/                   # 核心业务内部包（受 Go 编译器严格保护，禁止外部越权导入）
│   ├── handler/                # 控制器表现层（参数校验、反序列化、统一错误码映射）
│   │   ├── auth_handler.go     # 用户注册、登录、双 Token 无感静默刷新
│   │   ├── article_handler.go  # 资产 CRUD、多维分页与游标深分页、点赞/取消点赞
│   │   ├── upload_handler.go   # 获取 MinIO/S3 预签名直传 URL
│   │   └── config_handler.go   # 查询与修改 ETCD 动态热更配置
│   │
│   ├── service/                # 领域业务逻辑层（Bcrypt 加密、状态机编排、RBAC 校验）
│   │   ├── auth_service.go
│   │   ├── article_service.go
│   │   └── upload_service.go
│   │
│   ├── repository/             # 数据持久化与缓存仓储层（GORM 数据库交互、Redis 缓存双写）
│   │   ├── user_repo.go
│   │   ├── article_repo.go
│   │   └── cache_repo.go
│   │
│   ├── task/                   # 异步后台任务工作者 (Worker)
│   │   ├── article_audit.go    # 异步敏感词风控审查
│   │   └── stat_sync.go        # 高并发点赞与阅读量批量合并回写
│   │
│   ├── middleware/             # Gin 洋葱中间件
│   │   ├── cors.go             # 跨域安全配置（支持 Preflight OPTIONS）
│   │   ├── jwt.go              # JWT Bearer Token 拦截与 Context 上下文注入
│   │   ├── rbac.go             # RBAC 权限等级守卫
│   │   ├── idempotency.go      # X-Idempotency-Key 接口幂等性防重中间件
│   │   ├── recovery.go         # Panic 优雅捕获与兜底
│   │   └── logger.go           # 结构化请求耗时追踪日志
│   │
│   └── model/                  # 数据库实体与传输 DTO
│       ├── user.go             # 用户模型定义 (GORM Tags, JSON Tags)
│       └── article.go          # 文章模型定义 (状态机、游标支持)
│
├── pkg/                        # 可复用公共基础设施（纯技术工具，零业务字段，可直接跨项目复用）
│   ├── database/               # MySQL / SQLite 数据库初始化与连接池配置
│   ├── storage/                # MinIO / S3 客户端封装与 Pre-signed URL 生成
│   ├── queue/                  # Asynq 分布式异步任务客户端与服务配置
│   ├── etcd/                   # ETCD 租约注册与 Watch 动态配置监听器
│   ├── jwt/                    # 双 Token 签发与声明解析
│   ├── response/               # 统一前后端交互结构体与领域错误码映射
│   └── validator/              # 业务参数验证器
│
└── spec/                       # 详细模块规格说明书 (Common 通用底座 + Business 领域业务)
    ├── 00_overview_and_system_design.md
    ├── common/
    │   ├── 01_docker_and_infrastructure_spec.md
    │   ├── 02_middleware_and_lifecycle_spec.md
    │   ├── 03_database_and_migration_spec.md
    │   ├── 04_storage_presigned_spec.md
    │   ├── 05_async_task_engine_spec.md
    │   └── 06_frontend_contract_and_dod.md
    └── business/
        ├── 01_auth_and_user_spec.md
        ├── 02_article_and_asset_spec.md
        ├── 03_business_async_tasks_spec.md
        └── 04_etcd_dynamic_config_spec.md

---

## 🗓️ 工业级工程演进与实施路线

整个项目的研发落地严格遵循业界标准的大厂工程初始化次序：

1. **【Phase 0】工程身份确立与最小骨架**：`go.mod` 模块初始化、Standard Go Layout 目录结构落地、最小 `main.go` 与 `Makefile` 编译验证；
2. **【Phase 1】基础设施与配置引擎**：`docker-compose.yml` 本地中间件一键拉起、`configs/` 强类型配置加载与环境隔离；
3. **【Phase 2】统一契约与持久化基建**：通用响应封装、领域错误码字典、MySQL 8.0/SQLite 连接池、`golang-migrate` 版本迁移脚本；
4. **【Phase 3】Web 接入与洋葱中间件**：Gin 引擎初始化、Recovery、CORS、W3C Trace、接口幂等性防刷与 RBAC 权限守卫；
5. **【Phase 4】领域业务垂直切片**：用户认证与双 Token 状态机、知识资产 CRUD 与游标深分页、Asynq 分布式异步任务队列、ETCD 配置热更；
6. **【Phase 5】依赖装配与交付验收**：Google Wire 编译期无反射依赖装配、系统信号平滑优雅停机、`go test -race` 零并发竞争验收。
