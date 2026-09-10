package main

import (
	"fmt"
	"os"
	"runtime"

	"nexus-hub/configs"
)

// Version 与 BuildTime 变量由 Makefile 在构建时通过 -ldflags 动态覆盖
var (
	Version   = "v1.0.0-dev"
	BuildTime = "2026-09-10"
)

// main 服务唯一主程序入口
//
// 1. 【Why 为什么这么设计】:
//    任何大厂生产级服务启动的第一准则：必须“配置先行 (Config-First)”；
//    如果配置文件丢失、格式损坏、或端口配置非法，服务应在 1 毫秒内 Fast-Fail 立即熔断退出，
//    严禁带着“半残”的默认配置盲目启动，导致后续产生静默的数据写入 Bug。
//
// 2. 【Flow 底层执行流程】:
//    第一步：打印服务启动 Banner 与基础运行时信息 (PID, Go Version)；
//    第二步：调用 configs.Load("") 读取并校验 configs/config.yaml；
//    第三步：若配置校验失败，向 stderr 输出原因并调用 os.Exit(1) 立即阻断；
//    第四步：格式化输出已就绪的核心中间件连接元数据（数据库驱动、缓存节点、对象存储桶）；
//    第五步：后续模块将在此处拉起真实 Web 监听与优雅停机。
//
// 3. 【Gotcha 生产避坑】:
//    切勿在控制台日志中明文打印完整的密码或密钥串（如数据库密码、JWT Secret），
//    在生产环境中，打印敏感凭证会违反数据安全红线并引发审计处罚。
func main() {
	banner := `
  _   _                     _   _       _     
 | \ | |                   | | | |     | |    
 |  \| | _____  ___   _ ___| |_| |_   _| |__  
 | . ` + "`" + ` |/ _ \ \/ / | | / __|  _  | | | | '_ \ 
 | |\  |  __/>  <| |_| \__ \ | | | |_| | |_) |
 |_| \_|\___/_/\_\\__,_|___/_| |_|\__,_|_.__/ 
                                              
 Nexus-Hub Enterprise Business Platform [Go 1.27]
`
	fmt.Println(banner)
	fmt.Printf(">> [Bootstrap] Version       : %s\n", Version)
	fmt.Printf(">> [Bootstrap] Build Time    : %s\n", BuildTime)
	fmt.Printf(">> [Bootstrap] Go Runtime    : %s\n", runtime.Version())
	fmt.Printf(">> [Bootstrap] Platform      : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf(">> [Bootstrap] Process PID   : %d\n", os.Getpid())

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
}
