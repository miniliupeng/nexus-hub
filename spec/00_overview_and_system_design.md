# Nexus-Hub 规格说明书 00：全景概述与系统架构设计 (Overview & Architecture)

## 📌 1. 项目愿景与业务价值

`nexus-hub` 是一个专为现代化研发团队打造的**企业级智能知识资产与团队协作协同平台后端**。
本项目旨在打破传统小作坊式“Gin + GORM 简单增删改查”的简易 Demo 范式，完全按照一线大厂（字节跳动、美团、腾讯、阿里）核心业务中台的技术标准与规范实施落地。

### 核心设计原则：
1. **大厂工程规范严谨性**：严禁出现 `global.DB`、`global.Redis` 等反模式全局变量，强制推行 **Google Wire** 编译期显式依赖注入与洋葱分层架构（Clean Architecture / DDD）；
2. **数据库基准与规范演化**：以 **MySQL 8.0** 为生产基准主库，深度落地 InnoDB 联合索引设计、ACID 事务与行锁机制；通过 GORM 抽象兼顾 **SQLite 本地免安装零依赖双模式**；基于 **golang-migrate** 维护版本化数据库迁移脚本，拒绝手工运维改表；
3. **前后端深度闭环亲和性**：严格遵循契约优先（Contract-First）标准，开箱即用支持生产级跨域（CORS）、JWT 双 Token 无感静默刷新、统一标准响应与 TypeScript SDK / OpenAPI 3.0 契约；
4. **对象存储与预签名直传（S3/MinIO）**：颠覆传统低效的“前端传后端、后端存磁盘”模式，采用大厂标准的 **Pre-signed URL** 预签名机制，前端直接把文件推送到对象存储，节省 90% 后端带宽；
5. **异步任务与长耗时解耦（Asynq）**：集成基于 Redis 的分布式后台异步任务队列，解耦内容发布后的敏感词审核、热度计算、站内信通知等长耗时逻辑；
6. **分布式协同与高性能**：整合 **Redis Cache-Aside** 双写一致性机制，以及 **ETCD v3** 租约心跳注册与 Watch 毫秒级动态配置热更新；
7. **一键容器化编排**：提供工业级 `docker-compose.yml`，本地秒级启动 MySQL、Redis、ETCD、MinIO 完整基建。

---

## 🏛️ 2. 全景系统架构拓扑图

```text
                                  [ 前端 Web (React/Vue) / 移动端 / 自动化脚本 ]
                                    │                                      │
              (1) 获取预签名直传凭证 / 业务 API                             │ (2) 文件直接上传 (PUT)
                                    ▼                                      ▼
┌──────────────────────────────────────────────────────────────────┐ ┌─────────────────────────┐
│                    Nexus-Hub 企业级业务服务集群                   │ │ 对象存储集群 (MinIO/S3) │
│                                                                  │ └─────────────────────────┘
│   ┌──────────────────────────────────────────────────────────┐   │              ▲
│   │ 1. 洋葱中间件链路 (Onion Middleware Pipeline)            │   │              │
│   │   ├── Recovery: 捕获未知 Panic，堆栈全量落盘，保障永不崩溃│   │              │
│   │   ├── CORS: 生产级跨域处理器，响应 OPTIONS 预检请求      │   │              │
│   │   ├── Trace & Logger: 请求时延追踪，注入 W3C TraceID     │   │              │
│   │   └── JWT Guard: 拦截 Bearer Token，提取 CurrentUser 上下文│  │              │
│   └──────────────────────────────────────────────────────────┘   │              │
│                                 │                                │              │
│                                 ▼                                │              │
│   ┌──────────────────────────────────────────────────────────┐   │              │
│   │ 2. 控制器路由与契约接入层 (Handler Layer)                │   │              │
│   │   ├── Auth Handler: 用户注册、账号登录、双 Token 静默刷新│   │              │
│   │   ├── Article Handler: 资产多维组合查询、分页、软删除    │   │              │
│   │   ├── Upload Handler: 签发 MinIO/S3 Pre-signed 直传 URL ───┼──────────────┘
│   │   ├── ETCD Config Handler: 动态配置管理                  │   │
│   │   └── Response API: 统一标准 { code: 200, data, message }│   │
│   └──────────────────────────────────────────────────────────┘   │
│                                 │                                │
│                                 ▼                                │
│   ┌──────────────────────────────────────────────────────────┐   │
│   │ 3. 核心业务逻辑与领域服务层 (Service Layer - Wire 装配)  │   │
│   │   ├── AuthService: Bcrypt 密码哈希、双 Token 签发        │   │
│   │   ├── ArticleService: 复杂业务过滤、事务编排、缓存双写   │   │
│   │   ├── StorageService: MinIO 客户端封装与预签名 URL 签发  │   │
│   │   └── TaskProducer: 投递文章审核与热度计算异步任务       │   │
│   └──────────────────────────────────────────────────────────┘   │
│               │                         │              │         │
│               ▼                         ▼              ▼         │
│   ┌────────────────────────┐  ┌──────────────────┐  ┌────────┐   │
│   │ 4. 持久化仓储 (GORM)   │  │ 5. 缓存与分布式  │  │ 6. 异步│   │
│   │   ├── MySQL 8.0 (主库) │  │    协调层        │  │    任务│   │
│   │   ├── SQLite (免装双模)│  │   ├── Redis 7.x  │  │ (Asynq)│   │
│   │   ├── 事务控制与软删除 │  │   └── ETCD v3    │  │ Worker │   │
│   │   └── golang-migrate   │  │       Watch 热更 │  │ 消费者 │   │
│   └────────────────────────┘  └──────────────────┘  └────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

---

## 📂 3. 规格说明书（Spec）分类索引导航

Nexus-Hub 严格按照 **“通用工程基建 (Common)”** 与 **“领域业务逻辑 (Business)”** 实施物理隔离：

### 🛠️ 通用工程基建规格 (Common Specs) —— 跨项目 100% 可复用底座
1. **[01_docker_and_infrastructure_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/01_docker_and_infrastructure_spec.md)**：Docker Compose 本地编排全套环境（MySQL 8.0, Redis 7.0, ETCD v3, MinIO）与网络端口配置；
2. **[02_middleware_and_lifecycle_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/02_middleware_and_lifecycle_spec.md)**：服务优雅停机、Google Wire 编译期无反射注入、洋葱中间件链路（Recovery/CORS/Trace/JWT）；
3. **[03_database_and_migration_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/03_database_and_migration_spec.md)**：MySQL 8.0 连接池标准、SQLite 免安装双模抽象、golang-migrate 版本化演进；
4. **[04_storage_presigned_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/04_storage_presigned_spec.md)**：通用 MinIO/S3 预签名直传（Pre-signed URL）安全生成服务；
5. **[05_async_task_engine_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/05_async_task_engine_spec.md)**：基于 Asynq 的通用分布式异步任务引擎、工作协程池、重试退避与死信队列机制；
6. **[06_frontend_contract_and_dod.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/common/06_frontend_contract_and_dod.md)**：统一 ApiResponse 响应体、Axios 双 Token 静默刷新拦截器源码、全局交付验收标准 (DoD)。

---

### 💼 领域业务逻辑规格 (Business Specs) —— 专注业务流转与数据价值
1. **[01_auth_and_user_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/business/01_auth_and_user_spec.md)**：用户注册 Bcrypt 散列、双 Token 状态机流转、Redis 登录白名单与主动注销；
2. **[02_article_and_asset_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/business/02_article_and_asset_spec.md)**：知识资产 CRUD、MySQL 8.0 联合索引设计、复合动态分页查询、GORM 软删除与作者权限防护；
3. **[03_business_async_tasks_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/business/03_business_async_tasks_spec.md)**：文章内容敏感词合规风控审查异步任务、高并发阅读量合并延迟批量刷盘；
4. **[04_etcd_dynamic_config_spec.md](file:///Users/max/Desktop/lp/code/nexus-hub/spec/business/04_etcd_dynamic_config_spec.md)**：ETCD v3 业务参数（附件上传大小/审核开关）毫秒级热更、atomic.Pointer 内存安全指针设计。
