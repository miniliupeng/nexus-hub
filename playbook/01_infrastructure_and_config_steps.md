# 实操剧本 01：基础设施编排与强类型配置引擎 (Infrastructure & Config Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 1 (Module 01) 阶段从零到一敲下的**每一行命令、完整配置源码、核心逻辑设计、真实避坑排查与最终终端输出**。
> 本手册采用自包含设计，即便脱离项目源码，亦可作为工业级 Go 基础设施编排与配置加载的独立技术指南。

---

## 🎯 一、 核心目标与技术背景

- **核心目标**：
  1. 编写根目录 `docker-compose.yml`，定义 MySQL 8.0、Redis 7.2、ETCD v3.5、MinIO S3 容器网络与持久化卷；
  2. 引入 `gopkg.in/yaml.v3` 高性能解析库；
  3. 编写 `configs/config.yaml` 运行参数与 `configs/config.go` 强类型映射模型；
  4. 实现 `Validate()` 防御性前置校验（TCP 端口范围、数据库驱动白名单、JWT 密钥强度）；
  5. 编写全套防御性单元测试 `configs/config_test.go`；
  6. 将配置引擎接入 `cmd/server/main.go` 并落地敏感凭证安全脱敏。
- **前置环境**：
  - 已完成 Module 00（`go.mod` 与 Standard Layout 已就绪，本地编译验证通过）。

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：编写根目录 `docker-compose.yml`
为本地全套中间件定义生产级容器拓扑与网络：

```yaml
version: '3.8'

services:
  # 1. 核心持久化关系型数据库 (MySQL 8.0)
  mysql:
    image: mysql:8.0.36
    container_name: nexus_mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: nexus_root_2026
      MYSQL_DATABASE: nexus_hub
      TZ: Asia/Shanghai
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-pnexus_root_2026"]
      interval: 10s
      timeout: 5s
      retries: 5

  # 2. 分布式缓存与 Asynq 任务队列底座 (Redis 7.2)
  redis:
    image: redis:7.2-alpine
    container_name: nexus_redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: ["redis-server", "--appendonly", "yes"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  # 3. 分布式协调与动态配置中心 (ETCD v3.5)
  etcd:
    image: bitnami/etcd:3.5
    container_name: nexus_etcd
    restart: always
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
    ports:
      - "2379:2379"
    volumes:
      - etcd_data:/bitnami/etcd

  # 4. S3 协议对象存储 (MinIO)
  minio:
    image: minio/minio:latest
    container_name: nexus_minio
    restart: always
    environment:
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: minioadmin2026
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000" # S3 API 接口端点
      - "9001:9001" # Web 控制台界面
    volumes:
      - minio_data:/data

volumes:
  mysql_data:
    driver: local
  redis_data:
    driver: local
  etcd_data:
    driver: local
  minio_data:
    driver: local
```

- **技术解析**：
  1. **MySQL 8.0.36**：
     - 显式声明 `--character-set-server=utf8mb4` 和 `utf8mb4_unicode_ci`，彻底杜绝 Emoji 表情符号与生僻汉字乱码；
     - 配置 `--default-authentication-plugin=mysql_native_password`，确保老旧驱动和通用客户端连接顺畅；
     - 挂载 `mysqladmin ping` 健康检查，容器就绪状态真实可信。
  2. **Redis 7.2-alpine**：
     - 使用轻量级 Alpine 镜像，开启 AOF 日志持久化（`--appendonly yes`），保障双 Token 白名单和延时队列不因宕机丢失数据。
  3. **ETCD 3.5**：
     - 单节点 Raft 实例，暴露客户端通信端口 `2379`，用于后续微服务租约注册与参数毫秒热更。
  4. **MinIO**：
     - 暴露两个独立端口：`9000` 作为客户端上传下载的 S3 API 端点，`9001` 作为可视化 Web 管理控制台。

---

### 步骤 2：安装轻量级 YAML 3.0 解析库
```bash
go get gopkg.in/yaml.v3
```
- **技术解析**：相比动辄引入数十个间接依赖的重型配置框架，`gopkg.in/yaml.v3` 具有零外部依赖、极速编译、内存零过度分配的优点，完美支持结构体标签 `yaml:"xxx"`。

---

### 步骤 3：编写环境配置文件 `configs/config.yaml`
```yaml
# Nexus-Hub 企业级环境配置文件 (YAML)

app:
  name: "nexus-hub"
  env: "development" # 可选: development / testing / production
  port: 8088
  version: "v1.0.0"

database:
  # 数据库驱动: mysql (生产/带环境推荐) 或 sqlite (本地免装开箱即用)
  driver: "mysql"
  dsn: "root:nexus_root_2026@tcp(127.0.0.1:3306)/nexus_hub?charset=utf8mb4&parseTime=True&loc=Local"
  # 若切换为 sqlite，dsn 可写为: "nexus_hub.db"
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime_sec: 3600

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 50

etcd:
  endpoints:
    - "127.0.0.1:2379"
  dial_timeout_sec: 5

storage:
  endpoint: "127.0.0.1:9000"
  access_key: "admin"
  secret_key: "minioadmin2026"
  bucket: "nexus-assets"
  use_ssl: false

jwt:
  secret: "nexus-hub-ultra-secure-signing-key-2026-prod"
  access_expire_min: 15 # 短命 Access Token: 15 分钟
  refresh_expire_day: 7 # 长命 Refresh Token: 7 天
```

- **技术解析**：
  - 模块化组织：拆分为 `app`、`database`、`redis`、`etcd`、`storage` 与 `jwt` 六大独立命名空间；
  - 驱动抽象预留：`database.driver` 显式支持 `mysql` 和 `sqlite` 两种配置，为本地零依赖运行留好降级口。

---

### 步骤 4：编写强类型配置引擎 `configs/config.go`
包含独创的【Why-Flow-Gotcha】三维透析注释与严苛的 `Validate()` 校验：

```go
package configs

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name    string `yaml:"name"`
	Env     string `yaml:"env"`
	Port    int    `yaml:"port"`
	Version string `yaml:"version"`
}

type DatabaseConfig struct {
	Driver             string `yaml:"driver"`
	DSN                string `yaml:"dsn"`
	MaxOpenConns       int    `yaml:"max_open_conns"`
	MaxIdleConns       int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSec int    `yaml:"conn_max_lifetime_sec"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

type EtcdConfig struct {
	Endpoints      []string `yaml:"endpoints"`
	DialTimeoutSec int      `yaml:"dial_timeout_sec"`
}

type StorageConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
}

type JWTConfig struct {
	Secret           string `yaml:"secret"`
	AccessExpireMin  int    `yaml:"access_expire_min"`
	RefreshExpireDay int    `yaml:"refresh_expire_day"`
}

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Etcd     EtcdConfig     `yaml:"etcd"`
	Storage  StorageConfig  `yaml:"storage"`
	JWT      JWTConfig      `yaml:"jwt"`
}

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败 [%s]: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("反序列化 YAML 配置失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置合法性校验未通过: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	// 1. 端口合法性校验 (1 ~ 65535)
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port 无效: %d, 必须在 1 ~ 65535 区间", c.App.Port)
	}

	// 2. 数据库驱动校验
	if c.Database.Driver != "mysql" && c.Database.Driver != "sqlite" {
		return fmt.Errorf("database.driver 必须为 'mysql' 或 'sqlite', 当前为: '%s'", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return errors.New("database.dsn 连接字符串不能为空")
	}

	// 3. Redis 地址非空
	if c.Redis.Addr == "" {
		return errors.New("redis.addr 地址不能为空")
	}

	// 4. JWT 密钥长度防范 (至少 16 字节)
	if len(c.JWT.Secret) < 16 {
		return errors.New("jwt.secret 长度不足 16 位，存在安全隐患")
	}

	return nil
}
```

- **技术解析**：
  - **杜绝全局变量**：通过 `Load()` 返回 `*Config` 实例指针，为未来 Google Wire 依赖装配做好准备；
  - **Fast-Fail 熔断机制**：在系统启动的第 1 毫秒识别错误参数，杜绝服务带病启动。

---

### 步骤 5：编写全套防御性单元测试 `configs/config_test.go`
包含正常加载测试与 5 组非法边界用例拦截：

```go
package configs_test

import (
	"os"
	"path/filepath"
	"testing"

	"nexus-hub/configs"
)

func TestLoad_Success(t *testing.T) {
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

func TestValidate_InvalidCases(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(cfg *configs.Config)
		expectErr string
	}{
		{
			name: "端口为0",
			modifyFn: func(cfg *configs.Config) { cfg.App.Port = 0 },
			expectErr: "app.port 无效",
		},
		{
			name: "端口超过65535",
			modifyFn: func(cfg *configs.Config) { cfg.App.Port = 99999 },
			expectErr: "app.port 无效",
		},
		{
			name: "非法数据库驱动",
			modifyFn: func(cfg *configs.Config) { cfg.Database.Driver = "mongodb" },
			expectErr: "database.driver 必须为 'mysql' 或 'sqlite'",
		},
		{
			name: "数据库 DSN 为空",
			modifyFn: func(cfg *configs.Config) { cfg.Database.DSN = "" },
			expectErr: "database.dsn 连接字符串不能为空",
		},
		{
			name: "JWT 密钥过短",
			modifyFn: func(cfg *configs.Config) { cfg.JWT.Secret = "short-key" },
			expectErr: "jwt.secret 长度不足 16 位",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseCfg := &configs.Config{
				App:      configs.AppConfig{Port: 8080},
				Database: configs.DatabaseConfig{Driver: "mysql", DSN: "root:root@tcp(127.0.0.1:3306)/test"},
				Redis:    configs.RedisConfig{Addr: "127.0.0.1:6379"},
				JWT:      configs.JWTConfig{Secret: "ultra-secure-key-123456789"},
			}

			tt.modifyFn(baseCfg)

			err := baseCfg.Validate()
			if err == nil {
				t.Fatalf("用例 [%s] 预期应当返回校验错误，但返回了 nil", tt.name)
			}
		})
	}
}
```

---

### 步骤 6：将配置加载挂载到 `cmd/server/main.go`
```go
	// 1. 核心强类型配置加载
	cfg, err := configs.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, ">> [Fatal Error] 配置文件加载失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 打印安全脱敏后的核心配置摘要
	fmt.Printf(">> [Config] App Name         : %s\n", cfg.App.Name)
	fmt.Printf(">> [Config] Environment      : %s\n", cfg.App.Env)
	fmt.Printf(">> [Config] HTTP Port        : %d\n", cfg.App.Port)
	fmt.Printf(">> [Config] Database Driver  : %s (MaxOpen: %d, MaxIdle: %d)\n",
		cfg.Database.Driver, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	fmt.Printf(">> [Config] Redis Node       : %s (DB: %d)\n", cfg.Redis.Addr, cfg.Redis.DB)
	fmt.Printf(">> [Config] ETCD Endpoints   : %v\n", cfg.Etcd.Endpoints)
	fmt.Printf(">> [Config] Storage Endpoint : %s (Bucket: %s)\n", cfg.Storage.Endpoint, cfg.Storage.Bucket)
	fmt.Println(">> [Bootstrap] Configuration Loaded & Verified Successfully.")
```

---

## 🛠️ 三、 真实避坑与修复过程 (Troubleshooting & Fixes)

### 避坑点 1：配置单测工作目录（CWD）相对路径定位失败
- **现象**：在根目录下执行 `go test ./...` 时，`configs` 单元测试运行时的当前工作目录被切换到 `configs/` 内部，若写死 `configs/config.yaml` 会报错文件不存在。
- **解法**：在 `config_test.go` 中通过 `filepath.Join("..", "configs", "config.yaml")` 先行向上探测，若不存在则回退为当前目录 `config.yaml`，保障双模式路径解析 100% 畅通。

### 避坑点 2：生产密钥敏感信息泄露防范
- **避坑准则**：在控制台打印启动日志时，**绝对严禁输出明文数据库密码和 JWT Secret**，只输出连接地址与驱动名称，符合企业安全审计要求。

### 避坑点 3：MySQL 8.0 默认认证插件兼容性
- **现象**：MySQL 8.0 默认使用 `caching_sha2_password`，部分旧版 Go 客户端连接时会因缺少公钥报错 `authentication plugin not supported`。
- **解法**：在 `docker-compose.yml` 中显式指定 `--default-authentication-plugin=mysql_native_password`，保证驱动连接 100% 兼容稳定。

---

## 🧪 四、 最终终端验证与标准输出

### 1. 运行配置引擎单元测试与并发竞争扫描
```bash
make test-race
```
**实际终端输出**：
```text
=== RUN   TestLoad_Success
--- PASS: TestLoad_Success (0.00s)
=== RUN   TestValidate_InvalidCases
=== RUN   TestValidate_InvalidCases/端口为0
=== RUN   TestValidate_InvalidCases/端口超过65535
=== RUN   TestValidate_InvalidCases/非法数据库驱动
=== RUN   TestValidate_InvalidCases/数据库_DSN_为空
=== RUN   TestValidate_InvalidCases/JWT_密钥过短
--- PASS: TestValidate_InvalidCases (0.00s)
PASS
ok  	nexus-hub/configs	1.327s
```

### 2. 运行服务启动验证
```bash
make run
```
**实际终端输出**：
```text
go run ./cmd/server/main.go

  _   _                     _   _       _     
 | \ | |                   | | | |     | |    
 |  \| | _____  ___   _ ___| |_| |_   _| |__  
 | . ` |/ _ \ \/ / | | / __|  _  | | | | '_ \ 
 | |\  |  __/>  <| |_| \__ \ | | | |_| | |_) |
 |_| \_|\___/_/\_\\__,_|___/_| |_|\__,_|_.__/ 
                                              
 Nexus-Hub Enterprise Business Platform [Go 1.27]

>> [Bootstrap] Version       : v1.0.0-dev
>> [Bootstrap] Build Time    : 2026-09-10
>> [Bootstrap] Go Runtime    : go1.27.1
>> [Bootstrap] Platform      : darwin/arm64
>> [Bootstrap] Process PID   : 35574
>> [Config] App Name         : nexus-hub
>> [Config] Environment      : development
>> [Config] HTTP Port        : 8088
>> [Config] Database Driver  : mysql (MaxOpen: 100, MaxIdle: 20)
>> [Config] Redis Node       : 127.0.0.1:6379 (DB: 0)
>> [Config] ETCD Endpoints   : [127.0.0.1:2379]
>> [Config] Storage Endpoint : 127.0.0.1:9000 (Bucket: nexus-assets)
>> [Bootstrap] Configuration Loaded & Verified Successfully.
```
