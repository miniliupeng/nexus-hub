# Nexus-Hub 模块化实现过程与技术复习手册 (Implementation & Dev-Log)

> **导读**：
> 本手册是 `nexus-hub` 项目的**全程实操记录与纯正后端技术演进宝典**。
> 整个开发旅程严格遵循业界标准的大厂工程初始化次序：**【先确立工程身份与最小骨架 ➔ 铺设基础设施与环境 ➔ 逐层实现领域核心业务 ➔ 编译期装配与优雅停机】**。
> 每一个模块在编码落地时，均会严格按照 [DEVELOPMENT_CONVENTION.md](./DEVELOPMENT_CONVENTION.md) 记录其**技术痛点**、**选型权衡 (Trade-offs)**、**底层核心机制与避坑点**以及**纯命令行验证方式**，方便随时系统复盘与求职答辩。

---

## 📑 模块演进与开发进度总览表

| 模块序号 | 模块名称 | 核心技术组件 | 解决的核心后端问题 | 状态 |
| :---: | :--- | :--- | :--- | :---: |
| **Module 00** | 工程身份确立与最小可运行骨架 | `go.mod`, Standard Go Layout, `Makefile`, `main.go` | 建立正规工程底座、验证本地编译与基础工具链畅通 | 待开始 |
| **Module 01** | 基础设施容器编排与强类型配置引擎 | Docker Compose, Viper, 强类型配置校验 | 生产级环境参数隔离、防配置漏配崩溃 | 待开始 |
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

*(后续每个模块的详细实现过程将在此持续追加记录)*
