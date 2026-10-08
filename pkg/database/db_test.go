package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"

	"nexus-hub/configs"
)

// TestNew_UnsupportedDriver 验证非 MySQL 驱动被严格拦截
func TestNew_UnsupportedDriver(t *testing.T) {
	cfg := configs.DatabaseConfig{
		Driver: "postgres",
		DSN:    "postgres://localhost",
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("期望返回非法驱动错误，但实际返回 nil")
	}
}

// TestDatabaseLifecycle_WithMock 基于 sqlmock 测试连接池配置与生命周期
func TestDatabaseLifecycle_WithMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("创建 sqlmock 失败: %v", err)
	}
	defer sqlDB.Close()

	cfg := configs.DatabaseConfig{
		Driver:             "mysql",
		MaxOpenConns:       80,
		MaxIdleConns:       15,
		ConnMaxLifetimeSec: 2700,
	}

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	// GORM 初始化时会触发一次 ping 检查
	mock.ExpectPing()

	db, err := NewWithDialector(dialector, cfg)
	if err != nil {
		t.Fatalf("NewWithDialector 失败: %v", err)
	}

	// 1. 验证连接池配置生效 (通过 stats 校验)
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 80 {
		t.Errorf("MaxOpenConnections = %d, want 80", stats.MaxOpenConnections)
	}

	// 2. 验证 Ping 健康探测
	mock.ExpectPing()
	if err := Ping(db); err != nil {
		t.Errorf("Ping 失败: %v", err)
	}

	// 3. 验证 Close 正常关闭
	mock.ExpectClose()
	if err := Close(db); err != nil {
		t.Errorf("Close 失败: %v", err)
	}

	// 4. 验证 mock 预期全部满足
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock 预期未达成: %v", err)
	}
}

// TestDatabase_NilGuards 测试对 nil 实例的防御性拦截
func TestDatabase_NilGuards(t *testing.T) {
	if err := Ping(nil); err != ErrNilDatabase {
		t.Errorf("Ping(nil) = %v, want %v", err, ErrNilDatabase)
	}

	if err := Close(nil); err != ErrNilDatabase {
		t.Errorf("Close(nil) = %v, want %v", err, ErrNilDatabase)
	}
}
