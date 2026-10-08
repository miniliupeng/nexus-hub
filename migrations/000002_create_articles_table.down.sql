-- 000002_create_articles_table.down.sql
-- 幂等回滚删除 articles 表
DROP TABLE IF EXISTS `articles`;
