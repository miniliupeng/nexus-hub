# Nexus-Hub 模块化实现过程与技术复习手册 (Implementation & Dev-Log)

> **导读**：
> 本手册是 `nexus-hub` 项目的**全程实操记录与纯正后端技术演进宝典**。
> 整个开发旅程严格遵循业界标准的大厂工程初始化次序：**【先确立工程身份与最小骨架 ➔ 铺设基础设施与环境 ➔ 逐层实现领域核心业务 ➔ 编译期装配与优雅停机】**。
> 每一个模块在编码落地时，均会严格按照 [DEVELOPMENT_CONVENTION.md](./DEVELOPMENT_CONVENTION.md) 记录其**技术痛点**、**选型权衡 (Trade-offs)**、**底层核心机制与避坑点**以及**纯命令行验证方式**，方便随时系统复盘与求职答辩。

---

## 📑 模块演进与开发进度总览表

| 模块序号 | 模块名称 | 核心技术组件 | 解决的核心后端问题 | 状态 |
| :---: | :--- | :--- | :--- | :---: |
| **Module 00** | 工程身份确立与最小可运行骨架 | `go.mod`, Standard Go Layout, `Makefile`, `main.go` | 建立正规工程底座、验证本地编译与基础工具链畅通 | **已完成** ✅ |
| **Module 01** | 基础设施容器编排与强类型配置引擎 | Docker Compose, Viper/YAML, 强类型配置校验 | 生产级环境参数隔离、防配置漏配崩溃 | **已完成** ✅ |
| **Module 02** | 通用响应契约与领域业务错误码 | Generic API Response, Domain Errors | 统一前后端交互协议、错误精准溯源 | 待开始 |
| **Module 03** | 数据库连接池与版本化 SQL 迁移 | MySQL 8.0, SQLite 双模, golang-migrate | 连接复用防耗尽、数据库版本演化可追溯 | 待开始 |
| **Module 04** | 洋葱中间件链路与安全防线 | Recovery, CORS 预检, W3C TraceID, 接口幂等性 | 进程崩溃兜底、跨域预检、弱网防重复提交 | 待开始 |
| **Module 05** | 用户认证领域与双 Token 状态机 | Bcrypt 哈希, JWT 双 Token, Redis 白名单 | 密码加盐散列、登录态安全与无感刷新 | 待开始 |
| **Module 06** | 对象存储服务与预签名直传凭证 | MinIO / S3 SDK, Pre-signed URL 直传 | 避免大文件穿透后端、节约 90% 网卡带宽 | 待开始 |
| **Module 07** | 知识资产 CRUD、状态机与游标深分页 | GORM 联合索引, Keyset Pagination, 软删除 | 消除 Offset 慢查询、严谨文档全生命周期 | 待开始 |
| **Module 08** | 高频互动防刷与分布式异步任务队列 | Redis Set 去重, Asynq 延时/重试/死信队列 | 业务长耗时解耦、读写热点合并削峰 | 待开始 |
| **Module 09** | ETCD 动态配置中心与毫秒级热更新 | ETCD v3 Lease, clientv3.Watch, atomic.Pointer | 业务运行参数免重启毫秒级动态生效 | 待开始 |
| **Module 10** | Google Wire 依赖静态装配与平滑停机 | Wire 编译期注入, SIGTERM 优雅注销 | 彻底根除全局变量、防止发版在途请求被掐断 | 待开始 |

---

## 🛠️ Module 00：工程身份确立与最小可运行骨架

### 1. 业务背景与技术痛点 (Problem & Context)
- **痛点**：若跳过工程初始化直接编写业务逻辑，会导致没有根包名（import 依赖错乱）、无统一构建入口（团队成员构建参数不一致）、缺乏目录分层（代码混乱揉在根目录）；
- **解法**：依据业界公认的 `golang-standards/project-layout`，首先确立 `nexus-hub` 独立模块身份，建立 `cmd/server`、`internal/`、`pkg/`、`configs/` 物理目录，并编写工业级 `Makefile` 统领全流程工程动作。

### 2. 核心架构与设计选型 (Architecture & Rationale)
- **模块初始化**：基于当前官方支持的最新版 Go 1.27.1 执行 `go mod init nexus-hub`；
- **目录规划**：
  - `cmd/server/`：唯一可编译 `main` 包入口，严格保持轻薄，不塞业务逻辑；
  - `internal/`：受 Go 编译器强制访问控制保护，避免私有业务逻辑被外部代码越权导入；
  - `pkg/`：纯技术组件库，无业务属性，支持未来多项目跨工程复用；
  - `Makefile`：工程指令枢纽，封装编译、测试、迁移、Docker 编排。

### 3. 关键代码机制与底层避坑 (Key Implementation & Gotchas)
- **动态链接版本注入 (-ldflags)**：
  在 `Makefile` 编译阶段通过 `-ldflags "-X main.Version=... -X main.BuildTime=..."` 在编译期向二进制注入元数据，无需在代码里写死版本号；
- **启动元数据标准化输出**：
  在进程启动瞬间打印 Process PID、Go Runtime、OS/Arch 平台信息，确保在容器与 K8s 集群运维现场能第一眼确认实例健康度与身份。

### 4. 命令行验证与预期输出 (Verification & CLI)
- **测试命令**：
  ```bash
  make build && ./bin/server
  ```
- **实际终端输出**：
  ```text
  >> 正在编译 nexus-hub 服务...
  go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=2026-09-10" -o bin/server ./cmd/server
  >> 编译完成: bin/server

    _   _                     _   _       _     
   | \ | |                   | | | |     | |    
   |  \| | _____  ___   _ ___| |_| |_   _| |__  
   | . ` |/ _ \ \/ / | | / __|  _  | | | | '_ \ 
   | |\  |  __/>  <| |_| \__ \ | | | |_| | |_) |
   |_| \_|\___/_/\_\\__,_|___/_| |_|\__,_|_.__/ 
                                                
   Nexus-Hub Enterprise Business Platform [Go 1.27]

  >> [Bootstrap] Version       : v1.0.0
  >> [Bootstrap] Build Time    : 2026-09-10
  >> [Bootstrap] Go Runtime    : go1.27.1
  >> [Bootstrap] Platform      : darwin/arm64
  >> [Bootstrap] Process PID   : 29292
  >> [Bootstrap] Status        : Engine Core Initialized Successfully.
  ```

---

## 🛠️ Module 01：基础设施容器编排与强类型配置引擎

### 1. 业务背景与技术痛点 (Problem & Context)
- **痛点**：
  - 传统开发中环境搭建碎片化，新人接手需手动安装 MySQL/Redis/ETCD/MinIO，配置账号端口容易踩坑；
  - 很多项目将配置写死在代码中，或者使用没有类型约束的弱类型 `map[string]interface{}`，一旦漏配、错配某个字段，服务往往在运行数小时后因空指针（Nil Pointer）诡异崩溃；
- **解法**：
  - 编写 `docker-compose.yml` 声明标准化生产级中间件拓扑，一键 `make docker-up` 秒级拉起；
  - 采用 Go Struct 强类型映射 YAML，并在启动第一毫秒执行严格的 `Validate()` 边界校验（Fast-Fail 机制）。

### 2. 核心架构与设计选型 (Architecture & Rationale)
- **容器拓扑**：
  - MySQL 8.0.36（端口 3306，指定 `utf8mb4_unicode_ci` 字符集）；
  - Redis 7.2-alpine（端口 6379，开启 AOF 持久化）；
  - ETCD 3.5（端口 2379，单节点 Raft 状态机）；
  - MinIO（端口 9000 API，9001 控制台，S3 协议对象存储）。
- **配置模型设计**：
  - 模块化拆分子配置结构体：`AppConfig`, `DatabaseConfig`, `RedisConfig`, `EtcdConfig`, `StorageConfig`, `JWTConfig`；
  - 构造函数 `configs.Load(path string)` 统一加载，严禁使用全局变量 `global.Config`。

### 3. 关键代码机制与底层避坑 (Key Implementation & Gotchas)
- **Fast-Fail 启动前置熔断校验**：
  在 `Validate()` 中校验 TCP 端口区间（1~65535）、数据库驱动白名单（只允许 mysql/sqlite）、DSN 必填性以及 JWT 密钥长度防暴力破解（至少 16 字节）；
- **配置敏感信息脱敏**：
  主程序控制台打印配置元数据时，严禁输出明文数据库密码和 JWT Secret，保护系统安全合规。

### 4. 命令行验证与预期输出 (Verification & CLI)
- **测试命令**：
  ```bash
  make test-race && make run
  ```
- **实际终端输出**：
  ```text
  go test -v -race ./...
  === RUN   TestLoad_Success
  --- PASS: TestLoad_Success (0.00s)
  === RUN   TestValidate_InvalidCases
  --- PASS: TestValidate_InvalidCases (0.00s)
  PASS
  ok  	nexus-hub/configs	1.327s

  go run ./cmd/server/main.go
  >> [Bootstrap] Version       : v1.0.0-dev
  >> [Bootstrap] Process PID   : 35574
  >> [Config] App Name         : nexus-hub
  >> [Config] Environment      : development
  >> [Config] HTTP Port        : 8088
  >> [Config] Database Driver  : mysql (MaxOpen: 100, MaxIdle: 20)
  >> [Config] Redis Node       : 127.0.0.1:6379 (DB: 0)
  >> [Config] ETCD Endpoints   : [127.0.0.1:2379]
  >> [Config] Storage Endpoint : 127.0.0.1:9000 (Bucket: nexus-assets)
  >> [Bootstrap] Configuration Loaded & Verified Successfully.
  ```

---
