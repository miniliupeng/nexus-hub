# 通用规格 05：通用分布式异步与延时任务引擎 (Async Task Engine)

## 📌 1. 架构定位
基于 Redis + Asynq 构建通用的分布式任务队列基础设施。它对具体的业务任务完全无知，仅负责：
- 任务入队与持久化调度（生产者 Producer）；
- 多工作协程并发拉取消费（消费者 Worker）；
- 自动指数退避重试（Exponential Backoff）；
- 超出最大重试次数转入死信队列（Dead Letter Queue - DLQ）。

---

## ⚙️ 2. 通用配置项
- 并发工作协程数（Concurrency）：默认 10；
- 队列优先级划分：`critical` (权重 6), `default` (权重 3), `low` (权重 1)；
- 任务超时控制（Task Timeout）与任务防丢持久化。
