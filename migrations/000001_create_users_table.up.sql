-- 000001_create_users_table.up.sql
-- 严格遵循 MySQL 8.0 规范，建立用户主体信息表
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户主键ID',
    `username` VARCHAR(64) NOT NULL UNIQUE COMMENT '登录账号名称',
    `email` VARCHAR(128) NOT NULL UNIQUE COMMENT '用户电子邮箱',
    `password_hash` VARCHAR(255) NOT NULL COMMENT 'Bcrypt 加密安全散列',
    `avatar` VARCHAR(255) DEFAULT '' COMMENT '头像地址 (MinIO 对象存储路径)',
    `role` VARCHAR(16) NOT NULL DEFAULT 'editor' COMMENT '角色: admin/editor/viewer',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账号状态: 1-正常, 0-冻结禁用',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间(毫秒)',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间(毫秒)',
    INDEX `idx_users_role` (`role`),
    INDEX `idx_users_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户主体信息表';
