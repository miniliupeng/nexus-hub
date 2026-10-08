-- 000001_create_users_table.down.sql
-- 幂等回滚删除 users 表
DROP TABLE IF EXISTS `users`;
