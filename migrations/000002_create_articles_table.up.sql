-- 000002_create_articles_table.up.sql
-- 严格遵循 MySQL 8.0 规范，建立知识资产与文档主表 (支持状态机流转与游标深分页)
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
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间(毫秒)',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间(毫秒)',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT 'GORM软删除时间戳',
    INDEX `idx_articles_author_id` (`author_id`),
    INDEX `idx_articles_category_status` (`category`, `status`),
    INDEX `idx_articles_status_id` (`status`, `id`),
    INDEX `idx_articles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识资产与文档主表';
