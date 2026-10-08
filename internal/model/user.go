package model

import (
	"time"
)

// RBAC 角色常量定义
const (
	RoleAdmin  = "admin"  // 系统管理员 (具备全量权限)
	RoleEditor = "editor" // 创作者/团队成员 (默认角色，支持写操作)
	RoleViewer = "viewer" // 访客/只读成员
)

// 用户账号状态常量定义
const (
	UserStatusDisabled int8 = 0 // 账号冻结禁用
	UserStatusActive   int8 = 1 // 账号正常
)

// User 用户主体持久化数据模型
//
// 1. 【Why 为什么这么设计】:
//    严格映射 MySQL 8.0 users 表结构，提供完整的 GORM 结构体标签与 JSON 标签；
//    PasswordHash 字段打上 json:"-" 标签，杜绝因序列化漏洞将 Bcrypt 哈希明文泄漏给前端。
//
// 2. 【Flow 底层执行流程】:
//    第一步：通过 TableName 显式绑定数据库表名 users；
//    第二步：主键采用 uint64 自增 ID，与 MySQL BIGINT UNSIGNED 严格对齐；
//    第三步：基于 idx_users_role 与 idx_users_created_at 支撑后台按角色与注册时间的快速检索。
//
// 3. 【Gotcha 生产避坑】:
//    任何输出用户信息的 DTO 转换中，均严禁附带 PasswordHash 字段。
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username     string    `gorm:"size:64;not null;uniqueIndex;column:username" json:"username"`
	Email        string    `gorm:"size:128;not null;uniqueIndex;column:email" json:"email"`
	PasswordHash string    `gorm:"size:255;not null;column:password_hash" json:"-"`
	Avatar       string    `gorm:"size:255;default:'';column:avatar" json:"avatar"`
	Role         string    `gorm:"size:16;not null;default:'editor';index:idx_users_role;column:role" json:"role"`
	Status       int8      `gorm:"not null;default:1;column:status" json:"status"`
	CreatedAt    time.Time `gorm:"not null;index:idx_users_created_at;column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;column:updated_at" json:"updated_at"`
}

// TableName 显式指定表名
func (User) TableName() string {
	return "users"
}
