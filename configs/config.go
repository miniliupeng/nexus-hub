package configs

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig 服务基础配置
type AppConfig struct {
	Name    string `yaml:"name"`
	Env     string `yaml:"env"`
	Port    int    `yaml:"port"`
	Version string `yaml:"version"`
}

// DatabaseConfig 持久化数据库配置 (支持 MySQL / SQLite 双驱动)
type DatabaseConfig struct {
	Driver             string `yaml:"driver"`
	DSN                string `yaml:"dsn"`
	MaxOpenConns       int    `yaml:"max_open_conns"`
	MaxIdleConns       int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSec int    `yaml:"conn_max_lifetime_sec"`
}

// RedisConfig 缓存与任务队列底座配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// EtcdConfig 分布式协调与动态配置中心
type EtcdConfig struct {
	Endpoints      []string `yaml:"endpoints"`
	DialTimeoutSec int      `yaml:"dial_timeout_sec"`
}

// StorageConfig S3/MinIO 对象存储配置
type StorageConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
}

// JWTConfig 鉴权令牌安全参数
type JWTConfig struct {
	Secret           string `yaml:"secret"`
	AccessExpireMin  int    `yaml:"access_expire_min"`
	RefreshExpireDay int    `yaml:"refresh_expire_day"`
}

// Config 全局根配置实体
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Etcd     EtcdConfig     `yaml:"etcd"`
	Storage  StorageConfig  `yaml:"storage"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// Load 解析并初始化系统全局强类型配置
//
// 1. 【Why 为什么这么设计】:
//    生产环境中，严禁代码硬编码端口与密码，严禁使用弱类型的全局 map[string]interface{}；
//    通过将 YAML 文件反序列化为编译期强类型 Struct，并在启动第一步执行严格的 Validate 校验，
//    能够杜绝“漏配参数导致服务在运行数小时后因空指针 Panic 崩溃”的恶性生产事故。
//
// 2. 【Flow 底层执行流程】:
//    第一步：若未指定路径，则回退使用默认路径 "configs/config.yaml"；
//    第二步：通过 os.ReadFile 一次性将配置文件读取到内存字节切片；
//    第三步：调用 yaml.Unmarshal 进行强类型反序列化；
//    第四步：调用 cfg.Validate() 进行网络端口范围、数据库驱动合法性与安全密钥长度校验；
//    第五步：返回初始化就绪的 *Config 指针对象。
//
// 3. 【Gotcha 生产避坑】:
//    严禁将 *Config 赋值给包级全局变量（如 global.Config），必须通过构造函数返回其实例，
//    并在随后的 Google Wire 依赖装配中显式注入给下游组件，实现依赖关系的透明与可测试性。
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

	// 启动期防御性前置校验
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置合法性校验未通过: %w", err)
	}

	return &cfg, nil
}

// Validate 校验配置参数是否处于合法取值区间
func (c *Config) Validate() error {
	// 1. 端口合法性校验 (TCP 端口范围 1 ~ 65535)
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port 无效: %d, 必须在 1 ~ 65535 区间", c.App.Port)
	}

	// 2. 数据库驱动校验 (仅允许生产 MySQL 或本地免装 SQLite)
	if c.Database.Driver != "mysql" && c.Database.Driver != "sqlite" {
		return fmt.Errorf("database.driver 必须为 'mysql' 或 'sqlite', 当前为: '%s'", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return errors.New("database.dsn 连接字符串不能为空")
	}

	// 3. Redis 必须声明地址
	if c.Redis.Addr == "" {
		return errors.New("redis.addr 地址不能为空")
	}

	// 4. JWT 安全密钥强度防范 (HMAC-SHA256 签名密钥长度至少 16 字节，防暴力破解)
	if len(c.JWT.Secret) < 16 {
		return errors.New("jwt.secret 长度不足 16 位，存在安全隐患")
	}

	return nil
}
