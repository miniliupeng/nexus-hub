package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"nexus-hub/configs"
)

var (
	ErrUnsupportedDriver = errors.New("不支持的数据库驱动类型，生产基准仅允许 mysql")
	ErrNilDatabase       = errors.New("数据库实例为 nil")
)

// New 使用强类型配置初始化生产级 MySQL GORM 实例
//
// 1. 【Why 为什么这么设计】:
//    严格遵守 Clean Architecture 准则，杜绝 global.DB 全局变量反模式；
//    通过显式构造函数传递依赖，并在工厂内部集中调优 database/sql 底层连接池。
//
// 2. 【Flow 底层执行流程】:
//    第一步：驱动合法性守卫（锁定生产基准 mysql）；
//    第二步：调用 mysql.Open(cfg.DSN) 创建驱动 Dialector；
//    第三步：调用 NewWithDialector 装配 GORM 实例并完成连接池各项阈值配置。
//
// 3. 【Gotcha 生产避坑】:
//    切勿在服务启动时不配置连接池直接裸跑，默认配置下的最大空闲连接数仅为 2，
//    高并发下会引发频繁建立/关闭 TCP 连接，极易导致客户端端口耗尽与 TIME_WAIT 堆积。
func New(cfg configs.DatabaseConfig) (*gorm.DB, error) {
	if cfg.Driver != "mysql" {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDriver, cfg.Driver)
	}

	dialector := mysql.Open(cfg.DSN)
	return NewWithDialector(dialector, cfg)
}

// NewWithDialector 支持传入自定义 Dialector 初始化并配置连接池（便于单元测试与驱动扩展）
//
// 1. 【Why 为什么这么设计】:
//    提高代码的可测试性（Testability），单元测试可通过此接口注入 mock 连接；
//    统一收口连接池的 MaxOpenConns、MaxIdleConns 与 ConnMaxLifetime 参数。
//
// 2. 【Flow 底层执行流程】:
//    第一步：调用 gorm.Open 初始化 GORM 会话，设置静音日志；
//    第二步：通过 db.DB() 获取底层标准库 *sql.DB 句柄；
//    第三步：设置连接池三大核心生命周期参数。
//
// 3. 【Gotcha 生产避坑】:
//    ConnMaxLifetime 必须设置（建议 30~60 分钟），避免云服务商或自建机房网关在后台
//    静默关闭空闲 TCP 连接，导致 Go 运行时复用僵尸连接引发 broken pipe 报错。
func NewWithDialector(dialector gorm.Dialector, cfg configs.DatabaseConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("初始化 GORM 数据库连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 sql.DB 连接池句柄失败: %w", err)
	}

	ConfigurePool(sqlDB, cfg)
	return db, nil
}

// ConfigurePool 集中调优底层 database/sql 连接池参数
func ConfigurePool(sqlDB *sql.DB, cfg configs.DatabaseConfig) {
	if sqlDB == nil {
		return
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeSec > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSec) * time.Second)
	}
}

// Ping 对当前数据库连接执行主动健康检查探测
func Ping(db *gorm.DB) error {
	if db == nil {
		return ErrNilDatabase
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

// Close 优雅安全关闭底层数据库连接池
func Close(db *gorm.DB) error {
	if db == nil {
		return ErrNilDatabase
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
