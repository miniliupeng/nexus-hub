# 业务规格 01：用户认证、双 Token 状态机与 RBAC 权限 (Auth & User Spec)

## 📌 1. 业务场景概述

用户是系统的操作主体。为满足企业级组织架构管理，系统内置**基于角色的权限访问控制（RBAC - Role-Based Access Control）**，并通过 **JWT 双 Token（短命 Access + 长命 Refresh）** 实现高安全与零感知的极致用户体验。

---

## 👥 2. 角色定义与权限分级矩阵

| 角色标识 (`role`) | 权限等级 | 业务含义 | 权限范围 |
| :--- | :---: | :--- | :--- |
| **`admin`** | 100 | 系统管理员 | 拥有全量资产的审阅、强制删除、用户启停、ETCD 动态配置修改权限 |
| **`editor`** | 50 | 研发团队成员 / 创作者 | 可创建、编辑、软删除属于本人的知识资产，可参与互动点赞 |
| **`viewer`** | 10 | 访客 / 审计员 | 仅具备文章、文档资产的只读检索权限，禁止写操作 |

---

## 🗄️ 3. 数据库表模型设计 (`users` 表)

```sql
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户主键ID',
    `username` VARCHAR(64) NOT NULL UNIQUE COMMENT '登录账号名称',
    `email` VARCHAR(128) NOT NULL UNIQUE COMMENT '用户电子邮箱',
    `password_hash` VARCHAR(255) NOT NULL COMMENT 'Bcrypt 加密安全散列',
    `avatar` VARCHAR(255) DEFAULT '' COMMENT '头像地址 (MinIO 对象存储路径)',
    `role` VARCHAR(16) NOT NULL DEFAULT 'editor' COMMENT '角色: admin/editor/viewer',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账号状态: 1-正常, 0-冻结禁用',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    INDEX `idx_users_role` (`role`),
    INDEX `idx_users_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户主体信息表';
```

---

## 🔄 4. 双 Token 状态机与 Redis 生命周期流转

```text
[ 用户输入账号密码登录 ]
          │
          ▼
[ 验证 Bcrypt 散列正确 ]
          │
          ├──▶ 签发 Access Token (有效期 15 分钟，前端放 localStorage / Pinia)
          │
          └──▶ 签发 Refresh Token (有效期 7 天，加密存入 Redis 白名单)
                   键名: auth:refresh:<refresh_token> ➔ 值: user_id:1001,role:editor
```

### 4.1 静默刷新状态机流转规则
1. **正常请求**：前端请求业务 API，携带 `Authorization: Bearer <access_token>`；
2. **Access Token 过期**：后端返回 HTTP 401（业务码 10004）；
3. **静默刷新触发**：前端 Axios 拦截器捕获 401，挂起并发队列，调用 `/api/v1/auth/refresh` 传入 `refresh_token`；
4. **Redis 状态机校验**：
   - 后端查询 Redis：若 `auth:refresh:<refresh_token>` 存在且对应用户未被冻结，**执行旧 Token 轮转废弃并生成全新 Access Token**；
   - 若 Redis 中已被删除（如用户在其他端修改密码或主动注销），则返回 `401 Refresh Token Invalid`（业务码 10005），前端强制注销回登录页。
