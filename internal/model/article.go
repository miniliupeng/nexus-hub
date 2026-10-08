package model

import (
	"time"

	"gorm.io/gorm"
)

// 知识资产状态机流转枚举
const (
	ArticleStatusDraft     = "draft"     // 草稿箱 (仅作者可见)
	ArticleStatusPending   = "pending"   // 审核中 (触发异步风控审查)
	ArticleStatusPublished = "published" // 已正式发布 (全网可查阅)
	ArticleStatusRejected  = "rejected"  // 审核驳回 (作者可修改重提)
	ArticleStatusArchived  = "archived"  // 已归档下线
)

// Article 知识资产持久化数据模型
//
// 1. 【Why 为什么这么设计】:
//    严格映射 MySQL 8.0 articles 表结构；
//    设计复合联合索引 (status, id) 支撑游标深分页（Keyset Pagination），避免高并发下的慢查询 I/O；
//    内置 gorm.DeletedAt 软删除机制，保障核心数据资产可审计追溯。
//
// 2. 【Flow 底层执行流程】:
//    第一步：通过 TableName 绑定 articles 表；
//    第二步：提供完整的状态机常量，供 Service 层流转校验；
//    第三步：基于 GORM 软删除机制，默认查询将自动追加 deleted_at IS NULL 过滤条件。
//
// 3. 【Gotcha 生产避坑】:
//    在游标深分页场景下，WHERE id < ? AND status = 'published' ORDER BY id DESC 会精准命中
//    (status, id) 联合索引的最左前缀与主键聚簇扫描，切勿将索引列顺序倒置！
type Article struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Title        string         `gorm:"size:128;not null;column:title" json:"title"`
	Summary      string         `gorm:"size:255;default:'';column:summary" json:"summary"`
	Content      string         `gorm:"type:longtext;not null;column:content" json:"content"`
	Category     string         `gorm:"size:32;not null;default:'default';index:idx_articles_category_status,priority:1;column:category" json:"category"`
	Status       string         `gorm:"size:16;not null;default:'draft';index:idx_articles_category_status,priority:2;index:idx_articles_status_id,priority:1;column:status" json:"status"`
	RejectReason string         `gorm:"size:255;default:'';column:reject_reason" json:"reject_reason,omitempty"`
	ViewCount    uint32         `gorm:"not null;default:0;column:view_count" json:"view_count"`
	LikeCount    uint32         `gorm:"not null;default:0;column:like_count" json:"like_count"`
	AuthorID     uint64         `gorm:"not null;index:idx_articles_author_id;column:author_id" json:"author_id"`
	CreatedAt    time.Time      `gorm:"not null;column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index:idx_articles_deleted_at;column:deleted_at" json:"-"`
}

// TableName 显式指定表名
func (Article) TableName() string {
	return "articles"
}
