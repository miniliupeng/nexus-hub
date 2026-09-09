# 业务规格 03：业务异步任务与最终一致性 (Business Async Tasks)

## 📌 1. 业务场景定义

在知识资产业务流中，以下两项重型操作必须脱离 HTTP 同步事务，交由异步任务执行：

### 任务 1：文章内容敏感词合规风控审查 (`task:article_audit`)
- **触发时机**：用户成功发布文章后，立即以 1ms 投递任务到队列；
- **业务处理**：
  1. Worker 提取文章正文；
  2. 比对系统敏感词库（涉政、黑产、侵权）；
  3. 若违规，开启 DB 事务将文章状态由 `published` 改为 `rejected`，并向作者用户发送系统告警站内信。

### 任务 2：高并发阅读量合并刷盘 (`task:view_count_sync`)
- **痛点**：若每被访问一次就执行 `UPDATE articles SET view_count = view_count + 1 WHERE id = ?`，高并发下行锁剧烈争用，MySQL 会被瞬间打垮。
- **业务处理**：
  1. 用户阅读文章时，仅在 Redis 中执行 `HINCRBY article:views <article_id> 1` 极速返回；
  2. 定时延时任务每 30 秒聚合读取 Redis 中的增量，合并批量一次性 `UPDATE` 回写 MySQL。
