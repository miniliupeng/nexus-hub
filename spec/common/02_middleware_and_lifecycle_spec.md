# 通用规格 02：服务生命周期、依赖注入与洋葱中间件 (Lifecycle & Middleware)

## 📌 1. 服务生命周期与优雅停机 (Graceful Shutdown)
- **目标**：杜绝强制 kill 进程导致用户请求中途失败或事务破损。
- **机制**：
  1. `cmd/server/main.go` 监听操作系统信号 `SIGINT`, `SIGTERM`；
  2. 收到终止信号后，停止接收新连接，设置 10 秒超时 Context；
  3. 调用 `httpServer.Shutdown(ctx)` 等待现有活跃在途 HTTP 请求处理完毕；
  4. 顺序关闭 Redis、数据库连接池、ETCD Lease 租约注销以及 Asynq Worker 协程池。

---

## 💉 2. 依赖注入体系 (Google Wire)
- **目标**：100% 杜绝 `global.DB` 等包级全局变量，实现代码完全松耦合。
- **机制**：
  1. 在 `cmd/server/wire.go` 中声明 Provider 集合（配置、DB、Redis、ETCD、Repo、Service、Handler）；
  2. 执行 `wire ./cmd/server`，在编译期由静态分析自动生成装配代码 `wire_gen.go`；
  3. 任何组件只依赖接口或构造函数参数传入的依赖项，单测时可轻松 Mock。

---

## 🧅 3. 洋葱中间件流水线 (Middleware Pipeline)
所有进入系统的 HTTP 请求严格依次穿透以下 6 层通用中间件：

```text
[HTTP Request]
   │
   ▼
1. Recovery (捕获 Panic，保障进程存活)
   │
   ▼
2. CORS (处理预检 OPTIONS，放行复杂头)
   │
   ▼
3. Trace & Logger (注入 W3C TraceID，记录请求耗时)
   │
   ▼
4. Idempotency Guard (接口幂等性防刷：基于 X-Idempotency-Key + Redis SetNX)
   │
   ▼
5. JWT Guard (验证 Bearer Token，提取当前用户 user_id / role)
   │
   ▼
6. RBAC Guard (基于角色的权限守卫：校验是否达到路由要求的最低权限)
   │
   ▼
[Business Handler (业务控制器)]
```

### 3.1 核心防护中间件详述

#### 1. 接口幂等性中间件 (Idempotency Guard)
- **解决痛点**：前端网络抖动时用户多次猛击“提交/发布”，或前端断网重试导致后端在 MySQL 产生重复数据。
- **实现标准**：
  - 前端在创建类请求（`POST /api/v1/articles` 等）Header 中携带唯一键：`X-Idempotency-Key: <uuid-v4>`；
  - 中间件以 Redis 命令 `SET idempotency:<key> "PROCESSING" EX 10 NX` 抢占排他锁；
  - 若已存在且值为 `PROCESSING`，说明并发重复提交，直接拦截返回 `429 Too Many Requests`（业务码 30002）；
  - 若已存在且值为完整响应体 JSON，则直接返回原结果（实现网络超时重试时的安全幂等体验）。

#### 2. RBAC 角色权限中间件 (RBAC Guard)
- **角色等级定义**：
  - `admin` (管理员，最高权限，级别 100)
  - `editor` (内容创作者/普通业务员，具备写权限，级别 50)
  - `viewer` (访客，只读权限，级别 10)
- **拦截逻辑**：路由声明最低所需角色级别（如 `r.POST("/api/v1/config", RequireRole("admin"))`），未达标立即返回 `403 Forbidden`。
