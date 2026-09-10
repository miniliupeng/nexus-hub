package configs_test

import (
	"os"
	"path/filepath"
	"testing"

	"nexus-hub/configs"
)

// TestLoad_Success 验证成功加载默认的 config.yaml 文件
func TestLoad_Success(t *testing.T) {
	// 在单测中动态定位到项目的 configs/config.yaml
	cfgPath := filepath.Join("..", "configs", "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = "config.yaml"
	}

	cfg, err := configs.Load(cfgPath)
	if err != nil {
		t.Fatalf("预期加载配置成功，实际报错: %v", err)
	}

	if cfg.App.Port != 8088 {
		t.Errorf("预期端口为 8088，实际为: %d", cfg.App.Port)
	}

	if cfg.Database.Driver != "mysql" {
		t.Errorf("预期默认数据库驱动为 mysql，实际为: %s", cfg.Database.Driver)
	}

	if len(cfg.JWT.Secret) < 16 {
		t.Errorf("预期 JWT 密钥长度至少 16 字符")
	}
}

// TestValidate_InvalidCases 验证防御性前置校验的拦截边界
func TestValidate_InvalidCases(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(cfg *configs.Config)
		expectErr string
	}{
		{
			name: "端口为0",
			modifyFn: func(cfg *configs.Config) {
				cfg.App.Port = 0
			},
			expectErr: "app.port 无效",
		},
		{
			name: "端口超过65535",
			modifyFn: func(cfg *configs.Config) {
				cfg.App.Port = 99999
			},
			expectErr: "app.port 无效",
		},
		{
			name: "非法数据库驱动",
			modifyFn: func(cfg *configs.Config) {
				cfg.Database.Driver = "mongodb"
			},
			expectErr: "database.driver 必须为 'mysql' 或 'sqlite'",
		},
		{
			name: "数据库 DSN 为空",
			modifyFn: func(cfg *configs.Config) {
				cfg.Database.DSN = ""
			},
			expectErr: "database.dsn 连接字符串不能为空",
		},
		{
			name: "JWT 密钥过短",
			modifyFn: func(cfg *configs.Config) {
				cfg.JWT.Secret = "short-key"
			},
			expectErr: "jwt.secret 长度不足 16 位",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构造一个标准合规的基础配置对象
			baseCfg := &configs.Config{
				App:      configs.AppConfig{Port: 8080},
				Database: configs.DatabaseConfig{Driver: "mysql", DSN: "root:root@tcp(127.0.0.1:3306)/test"},
				Redis:    configs.RedisConfig{Addr: "127.0.0.1:6379"},
				JWT:      configs.JWTConfig{Secret: "ultra-secure-key-123456789"},
			}

			// 应用异常修改
			tt.modifyFn(baseCfg)

			err := baseCfg.Validate()
			if err == nil {
				t.Fatalf("用例 [%s] 预期应当返回校验错误，但返回了 nil", tt.name)
			}
		})
	}
}
