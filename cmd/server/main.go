package main

import (
	"fmt"
	"os"
	"runtime"
)

// Version 与 BuildTime 变量可由 Go 编译器在构建时通过 -ldflags 动态注入
var (
	Version   = "v1.0.0-dev"
	BuildTime = "2026-09-10"
)

// main 服务唯一主程序入口
//
// 1. 【Why 为什么这么设计】:
//    按照 Go 官方标准项目布局 (Standard Go Project Layout)，cmd/server/ 存放主程序的唯一起点；
//    生产级服务启动时，第一步必须以格式化 Banner 输出核心元数据（版本、环境、系统架构、进程 PID），
//    便于运维人员在容器控制台 (K8s Pod logs) 与排障现场第一时间确认版本与实例身份。
//
// 2. 【Flow 底层执行流程】:
//    第一步：从 runtime 包动态获取当前宿主架构与 Go 运行时版本；
//    第二步：获取操作系统分配给当前服务的唯一 Process ID (PID)；
//    第三步：向 stdout 打印标准的启动状态与就绪信号；
//    第四步：后续阶段将在此处挂载配置加载、依赖装配与优雅停机监听。
//
// 3. 【Gotcha 生产避坑】:
//    主入口 main 函数应保持极致精炼，严禁在 main.go 中堆砌具体的数据库 SQL 或 HTTP 路由业务代码，
//    main 的核心职责仅为：启动参数解析 -> 依赖装配 (Wire) -> 启动监听 -> 阻塞捕获系统停机信号。
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
	fmt.Println(">> [Bootstrap] Status        : Engine Core Initialized Successfully.")
}
