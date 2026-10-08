# Nexus 全栈 AI 智能体生态系统架构蓝图与模块演进规划 (Tri-Project Blueprint)

> **全局定位**：
> 本蓝图是面向现代企业级 **AI 全栈 / AI Agent 工程师** 岗位的终极实战交付规划。
> 针对纯前端转型痛点，打破单一语言或简单 Demo 的局限，采用业界成熟的**多语言分布式解耦架构（Polyglot Architecture）**。
> 整个生态由三个高内聚、低耦合的独立仓库协同驱动，各司其职，边做边学，产出可直接在线演示的高规格商业级作品集。

---

## 📑 架构全景拓扑与职责划分

```text
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                   浏览器 / 客户端用户端                                   │
└─────────────────────────────────────────────┬────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│  🎨 项目三：nexus-web (AI 原生交互与作品集门户)                                          │
│  • 技术栈：Next.js 15 (App Router, React 19) + TailwindCSS + Shadcn/ui + Vercel AI SDK   │
│  • 核心职责：消费 SSE 流式打字机、双栏 Artifacts 画板、Generative UI 动态渲染、RAG 溯源高亮 │
└──────────────────────────────┬───────────────────────────┬───────────────────────────────┘
                               │ (业务鉴权 / 静态资产)        │ (SSE 流式 / 任务调度)
                               ▼                           ▼
┌──────────────────────────────────────────────┐ ┌─────────────────────────────────────────┐
│  🏛️ 项目一：nexus-hub                         │ │  🧠 项目二：nexus-agent                 │
│  (企业级高并发业务中枢)                      │ │  (独立 AI Agent 与 RAG 智能引擎)        │
│  • 技术栈：Go 1.22+ + Gin + GORM + MySQL     │ │  • 技术栈：Node.js / TS + Hono          │
│    + Redis + Asynq + MinIO + Wire + ETCD     │ │    + LangGraph.js + Zod + pgvector      │
│  • 核心职责：用户与 RBAC 权限、数据库事务调优、│ │    + MCP SDK + Langfuse LLMOps          │
│    MinIO 预签名直传、分布式异步队列削峰      │ │  • 核心职责：大模型统一网关、SSE 流式推送、│
│                                              │ │    ReAct 状态机、工具调用、RAG 混合检索 │
└──────────────────────┬───────────────────────┘ └────────────────────┬────────────────────┘
                       │                                              │
                       ▼                                              ▼
┌──────────────────────────────────────────────┐ ┌─────────────────────────────────────────┐
│  基础设施集群：MySQL 8.0, Redis 7.2,          │ │  基础设施集群：PostgreSQL 16 + pgvector │
│  MinIO 对象存储, ETCD v3 (Docker 编排)        │ │  (关系型数据 + 高维向量存储统一)        │
└──────────────────────────────────────────────┘ └─────────────────────────────────────────┘
```

---

## 🏛️ 项目一：nexus-hub（企业级高并发业务中枢）

- **仓库路径**：`/Users/max/Desktop/lp/code/nexus-hub`
- **开发语言**：`Go 1.22+`
- **核心定位**：纯正大厂级高并发业务中台，建立扎实的后端体系，负责重度并发控制、权限鉴权与数据资产治理。

### 1. 核心技术栈与选型权衡
| 技术组件 | 选型标准 | 选型权衡与优势 (Trade-offs) |
| :--- | :--- | :--- |
| **Web 接入** | Gin / Echo | 极简、低内存占用、成熟的洋葱中间件生态 |
| **ORM & 迁移** | GORM + golang-migrate | 规范化关系映射，配合版本化 SQL 迁移避免生产表结构漂移 |
| **核心存储** | MySQL 8.0 (生产) + SQLite (双模) | 深入 InnoDB 联合索引、行级锁与事务隔离级别 |
| **缓存机制** | Redis 7.2 + go-redis v9 | Cache-Aside 模式、分布式防重锁、高频互动 Set 去重 |
| **对象存储** | MinIO SDK (S3 兼容) | 预签名 URL 直传（Pre-signed URL），解耦大文件上传带宽 |
| **异步削峰** | Asynq (基于 Redis) | 分布式延时任务、自动重试与死信队列，保障高并发削峰 |
| **配置中心** | ETCD 3.5 | Raft 一致性租约与 Watch 机制，支持运行期配置毫秒级热更 |
| **架构装配** | Google Wire | 编译期静态无反射依赖注入，根除全局变量，保障 Clean Architecture |

### 2. 职责边界
- ✅ **负责**：用户注册登录（JWT 双 Token 状态机）、RBAC 权限校验、文档全生命周期状态机、MinIO 预签名凭证签发、异步任务削峰。
- ❌ **不负责**：大模型接口转发、Prompt 组装、Agent 动态决策。

### 3. 模块化演进规划 (Module 00 ~ 10)
- **Module 00**：工程身份确立与最小可运行骨架 (`go.mod`, Standard Layout, Makefile) ✅
- **Module 01**：基础设施容器编排与强类型配置引擎 (Docker Compose, Fast-Fail 校验) ✅
- **Module 02**：通用响应契约与领域业务错误码 (Generic API Response, Domain Errors)
- **Module 03**：数据库连接池调优与版本化 SQL 迁移 (MySQL 8.0, golang-migrate)
- **Module 04**：洋葱中间件链路与安全防线 (Recovery, CORS, W3C TraceID, 接口幂等锁)
- **Module 05**：用户认证领域与双 Token 状态机 (Bcrypt, JWT 双 Token, Redis 白名单)
- **Module 06**：对象存储服务与预签名直传凭证 (MinIO / S3 SDK 直传，节约 90% 网卡带宽)
- **Module 07**：知识资产 CRUD、状态机与游标深分页 (GORM 联合索引, Keyset Pagination)
- **Module 08**：高频互动防刷与分布式异步任务队列 (Redis Set, Asynq 延时/重试/死信队列)
- **Module 09**：ETCD 动态配置中心与毫秒级热更新 (ETCD v3 Lease, Watch, atomic.Pointer)
- **Module 10**：Google Wire 依赖静态装配与平滑停机 (Wire 编译期注入, SIGTERM 优雅注销)

---

## 🧠 项目二：nexus-agent（独立 AI Agent 与 RAG 智能引擎）

- **仓库路径**：`/Users/max/Desktop/lp/code/nexus-agent`
- **开发语言**：`Node.js (LTS)` / `TypeScript 5+`
- **核心框架**：**Nest.js**（企业级 IoC/DI 模块化架构 + 纯手写自主 Agent 状态机引擎）
- **核心定位**：彻底告别第三方框架黑盒，基于 Nest.js 纯手撕打造的高可控、轻量级 Multi-Agent 多智能体协同与 RAG 知识检索系统。专攻底层 ReAct 状态机循环、Supervisor 多智能体调度、Zod 强类型工具注册表与 pgvector 向量混合检索。

### 1. 核心技术栈与选型权衡
| 技术组件 | 选型标准 | 选型权衡与优势 (Trade-offs) |
| :--- | :--- | :--- |
| **核心工程底座** | **Nest.js** (TypeScript) | 企业级 IoC 控制反转与依赖注入架构，每个智能体与工具作为独立 Service 解耦装配 |
| **底层流式通信** | **Vercel AI SDK Core (`ai`)** | 仅使用其最纯粹、低侵入的流式通信层与统一模型适配，零多余黑盒，性能与灵活性兼备 |
| **自研状态机引擎** | **Nest.js 原生 Provider (纯手撕)** | 拒绝 LangGraph 等第三方黑盒，亲手实现基于 StateGraph 的状态转移、循环熔断与 Handoff 调度 |
| **契约与工具中心**| **Zod** + 自研 ToolRegistry | 纯 TS 强类型契约，自动提取 JSON Schema 供大模型调用，提供入参自愈式修正与重试 |
| **向量与 RAG** | **PostgreSQL + pgvector** + Drizzle | 统一关系数据与高维向量存储，自研 HNSW 向量检索与 BM25 关键词检索的 RRF 融合算法 |
| **安全与人工审批**| Nest.js Guards & Checkpointer | 破坏性高危工具（写/删/支付）前置拦截，状态机自动挂起并生成快照，支持前端审批唤醒 |
| **工具标准扩展** | **MCP SDK** (`@modelcontextprotocol/sdk`) | 适配标准化 Model Context Protocol，实现标准 MCP Client，支持动态挂载外部沙箱工具 |
| **LLMOps 监控** | **Langfuse** + Nest.js Interceptors | 利用 AOP 切面无侵入统一捕获全链路 Trace、Agent 决策分支、Token 消耗统计与延迟归因 |

### 2. 职责边界
- ✅ **负责**：大模型统一接入（DeepSeek/OpenAI）、基于 Nest.js `@Sse()` 的响应式流式推流、自研 ReAct 循环状态机、自研 Supervisor 多智能体协同（Researcher + Coder + Reviewer）、工具调度与幂等防重、文档清洗切块与自研混合检索 RAG、人机协同（Human-in-the-loop）挂起审批。
- ❌ **不负责**：前端 UI 渲染、高并发通用权限与核心业务资产管理（由 `nexus-hub` 负责）。

### 3. 模块化演进规划 (Module 00 ~ 10，深度沉淀手撕路线)
- **Module 00**：Nest.js 骨架与 pgvector 容器编排 (Nest.js CLI, Docker Compose `pgvector`, TypeScript 严格工程规范)
- **Module 01**：大模型统一网关与 RxJS `@Sse()` 流式通道 (适配各大模型, Observable 响应式管道与心跳保活)
- **Module 02**：手写 ToolRegistry 工具中心与 Zod 强类型契约 (Schema 自动生成, 参数验证失败自愈修正, 幂等执行锁)
- **Module 03**：手撕 ReAct 状态机与核心循环引擎 (纯手写“思考-调用-观察”状态流转, 条件分支与死循环安全熔断)
- **Module 04**：手撕 Supervisor 多智能体协同系统 (意图分类路由, Researcher + Coder + Reviewer 角色分工与 A2A 协议)
- **Module 05**：检查点机制与 Human-in-the-loop 人工审批 (Nest.js Guard 高危拦截, 状态快照挂起持久化与断点唤醒)
- **Module 06**：分层记忆系统 (Memory Architecture) (短期 Scratchpad 暂存草稿 + 跨会话长期用户画像持久化与语义注入)
- **Module 07**：高级 RAG：知识治理与查询重写引擎 (手写多源文档递归切分算法, 多轮对话指代消解 Query Rewrite)
- **Module 08**：高级 RAG：向量混合检索与 RRF 融合重排 (手写 Dense 向量 + BM25 稀疏检索, 实现 RRF 倒数排名融合算法)
- **Module 09**：MCP 标准协议工具扩展适配器 (实现标准 MCP Client 模块, 动态挂载本地沙箱与外部标准化 API)
- **Module 10**：AOP 切面追踪与质量评估 Evals (Nest.js Interceptor 全链路 Trace 监控, RAG 三元组自动化质量评估)

---

## 🎨 项目三：nexus-web（AI 原生交互与作品集门户）

- **仓库路径**：`/Users/max/Desktop/lp/code/nexus-web`
- **开发语言**：`TypeScript` + `React 19`
- **核心定位**：充分展现 7 年资深前端人机交互底蕴，集流式打字机、双栏 Artifacts 画板、Generative UI 于一体的商业级作品集产品。

### 1. 核心技术栈与选型权衡
| 技术组件 | 选型标准 | 选型权衡与优势 (Trade-offs) |
| :--- | :--- | :--- |
| **核心框架** | **Next.js 15 (App Router)** | SSR 服务端渲染、流式 Suspense、企业级路由体系 |
| **UI 设计系统** | TailwindCSS + **Shadcn/ui** | 现代高质感设计规范，零冗余样式，纯组件代码完全可控 |
| **AI 交互驱动** | **Vercel AI SDK (`ai/react`)** | 工业级前端流式状态管理，深度封装 `useChat` 与工具回调 |
| **流式富文本** | `react-markdown` + `shiki` + KaTeX | 流式增量 Markdown 代码高亮，防标签未闭合白屏与抖动 |
| **全局状态** | Zustand | 轻量、无样板代码管理双栏联动与工作台状态 |

### 2. 职责边界
- ✅ **负责**：消费 `nexus-agent` 的 SSE 流、平滑打字机渲染、双栏联动（左侧对话 + 右侧实时代码/画板 Artifacts）、动态组件映射（Generative UI）、知识库文档切块可视化与引用溯源高亮、接入 `nexus-hub` 登录鉴权。
- ❌ **不负责**：大模型底层逻辑推理、底层长耗时任务执行。

### 3. 模块化演进规划 (Module 00 ~ 06)
- **Module 00**：Next.js 15 骨架与 Design Tokens (App Router, Tailwind, Shadcn/ui 工业级配置)
- **Module 01**：流式对话核心窗口与状态控制 (`useChat` 深度定制, 平滑触底滚动与流中断)
- **Module 02**：流式富文本解析与防闪烁高亮 (Shiki 增量语法高亮, Markdown 容错流式解析)
- **Module 03**：双栏联动 Artifacts / Canvas 预览 (Zustand 驱动, 可拖拽 Split-Pane, 沙箱渲染)
- **Module 04**：Generative UI 与工具审批弹窗 (结构化卡片渲染, Tool Call 执行前人工交互卡片)
- **Module 05**：知识库管理与 RAG 溯源高亮 (拖拽上传, 向量分块可视化, 问答证据链悬浮高亮)
- **Module 06**：容器化交付与全站性能优化 (Dockerfile, Standalone 构建, 移动端全响应式)

---

## 🎯 整体开发推进路线建议 (Sprint 计划)

```text
【阶段一：夯实 Go 业务底座】
  nexus-hub 推进 Module 02 ~ 06（完成鉴权、MySQL/Redis、MinIO 直传与中间件）
    │
    ▼
【阶段二：构建 Agent 核心大脑】
  nexus-agent 初始化并落地 Module 00 ~ 04（打通 SSE 流式、Zod 工具中心与自研 ReAct 状态机）
    │
    ▼
【阶段三：前端流式工作台连通】
  nexus-web 初始化并落地 Module 00 ~ 03（打通流式打字机与双栏 Artifacts 画板，形成第一版可交互 Demo）
    │
    ▼
【阶段四：RAG 知识库与全链路闭环】
  nexus-agent 接入 pgvector 与混合检索 -> nexus-web 接入知识库管理与溯源 -> 接入 Langfuse 监控
```

---

## 💡 求职面试答辩优势归纳

1. **破除“纯前端”标签**：你不仅具备 7 年前端交互统治力，更能拿出遵循大厂标准的 Go 高并发服务（`nexus-hub`）和 Node.js 独立 Agent 引擎（`nexus-agent`）。
2. **多语言异构解耦架构**：清晰讲出“高并发与存储中枢在 Go，智能决策与工具编排在 Node/TS，极致用户体验在 Next.js”的设计权衡。
3. **可直接在线把玩的真实作品**：每一个项目都有清晰的代码规范、纯命令行验证方式与完整的 Docker 交付，彻底击穿“需要提供上线作品集”的硬性要求。
