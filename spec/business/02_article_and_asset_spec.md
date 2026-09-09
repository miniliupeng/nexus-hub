# 业务规格 02：知识资产状态机、深分页与互动规格 (Article & Asset Spec)

## 📌 1. 业务场景概述

知识资产是 Nexus-Hub 的核心价值承载。系统支持团队协作文档从草稿、提交审核、敏感词风控审查、正式发布到归档的**全流程状态机流转**，并提供**高性能游标分页**与**防刷点赞高频互动**支持。

---

## 🔄 2. 资产完整生命周期状态机

```text
               ┌───────────────────────┐
               │    draft (草稿箱)     │ ◀─── 仅作者本人可见
               └───────────┬───────────┘
                           │ 提交发布申请
                           ▼
               ┌───────────────────────┐
               │  pending (审核中)     │ ◀─── 触发 Asynq 异步敏感词合规风控
               └───────────┬───────────┘
                           │
             ┌─────────────┴─────────────┐
   审核通过  │                           │ 包含违规黑名单词
             ▼                           ▼
┌─────────────────────────┐ ┌─────────────────────────┐
│  published (已正式发布) │ │   rejected (驳回修改)   │ ──▶ 携带 reject_reason
└────────────┬────────────┘ └─────────────────────────┘     作者可修改后重新提交
             │ 归档下线
             ▼
┌─────────────────────────┐
│   archived (已归档)     │ ◀─── 仅管理员与原作者可查阅
└─────────────────────────┘
```

---

## 🗄️ 3. 数据库表结构模型设计 (`articles` 表)

```sql
CREATE TABLE IF NOT EXISTS `articles` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '资产主键ID',
    `title` VARCHAR(128) NOT NULL COMMENT '资产/文档标题',
    `summary` VARCHAR(255) DEFAULT '' COMMENT '简短摘要描述',
    `content` LONGTEXT NOT NULL COMMENT 'Markdown 正文内容',
    `category` VARCHAR(32) NOT NULL DEFAULT 'default' COMMENT '分类: 前端工程/后端架构/AI应用',
    `status` VARCHAR(16) NOT NULL DEFAULT 'draft' COMMENT '状态机: draft/pending/published/rejected/archived',
    `reject_reason` VARCHAR(255) DEFAULT '' COMMENT '风控审核不通过的原因说明',
    `view_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '阅读浏览次数',
    `like_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '获赞总数',
    `author_id` BIGINT UNSIGNED NOT NULL COMMENT '作者用户ID',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT 'GORM软删除时间戳',
    INDEX `idx_articles_author_id` (`author_id`),
    INDEX `idx_articles_category_status` (`category`, `status`),
    INDEX `idx_articles_status_id` (`status`, `id`),
    INDEX `idx_articles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识资产与文档主表';
```

---

## 📋 4. 关键接口设计与交互契约

### 4.1 高性能分页检索 `GET /api/v1/articles`
- **Query 入参**：
  | 参数名 | 类型 | 必填 | 默认值 | 说明 |
  | :--- | :--- | :--- | :--- | :--- |
  | `page` | int | 否 | 1 | 传统页码（浅分页） |
  | `page_size` | int | 否 | 10 | 每页条数 (1~100) |
  | `last_id` | int | 否 | 0 | **游标分页主键** (若传入则走 `id < last_id` 零回表极速深分页) |
  | `keyword` | string | 否 | 空 | 标题模糊匹配 |
  | `category` | string | 否 | 空 | 分类精确过滤 |
  | `status` | string | 否 | published | 状态 (普通用户仅可查 published) |

- **响应格式 (包含用户专属互动状态)**：
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "total": 56,
      "page": 1,
      "page_size": 10,
      "next_cursor": 195,
      "list": [
        {
          "id": 201,
          "title": "2026 大厂前端进阶 Go 深度指南",
          "summary": "详细剖析微前端与后端洋葱模型架构",
          "category": "后端架构",
          "status": "published",
          "view_count": 892,
          "like_count": 45,
          "is_liked": true,
          "author": {
            "id": 1001,
            "username": "developer_max"
          },
          "created_at": "2026-09-08T12:00:00Z"
        }
      ]
    }
  }
  ```

---

### 4.2 点赞/取消点赞 `POST /api/v1/articles/:id/like` (需登录 JWT)
- **前端交互痛点**：用户点赞后，界面小心心瞬间变亮；若用户反复连续点击，必须防刷防穿透。
- **后端高性能处理**：
  1. 使用 Redis 集合（Set）存储点赞用户列表：`article:likes:<article_id>`；
  2. 执行 `SISMEMBER` 检查是否已赞：
     - 若未赞：执行 `SADD` 记录用户 ID，并在 Redis 增量 `like_count + 1`；
     - 若已赞：执行 `SREM` 移除用户 ID，并在 Redis 增量 `like_count - 1`（支持取消点赞）；
  3. 定时由异步任务批量将最新的 `like_count` 刷回 MySQL，彻底免除数据库行锁高并发争用！
