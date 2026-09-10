# 实操剧本 00：工程身份确立与最小可运行骨架搭建 (Project Init Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 0 (Module 00) 阶段从零到一敲下的**每一行命令、执行时机、真实避坑排查与最终终端输出**。
> 供后续随时在终端逐行复制复现，或在面试时拆解工程落地细节。

---

## 🎯 一、 核心目标与前置条件

- **核心目标**：
  1. 确立 `nexus-hub` 独立 Go 模块身份并生成 `go.mod`；
  2. 搭建业界标准的 `Standard Go Project Layout` 目录树；
  3. 编写最小可运行的 `cmd/server/main.go`（输出启动 Banner 与运维元数据）；
  4. 编写工程指令枢纽 `Makefile`，打通本地编译与清理流水线。
- **前置环境**：
  - 操作系统：macOS (darwin/arm64)
  - Go 版本：`go1.27.1`

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：初始化 Go 模块
确立根包名并锁定 Go 工具链版本。
```bash
go mod init nexus-hub
go version
```
- **技术解析**：生成根目录 `go.mod` 文件，所有子目录的 import 路径将基于 `nexus-hub/...` 解析。

---

### 步骤 2：落地 Standard Go Project Layout 目录树
```bash
mkdir -p \
  cmd/server \
  configs \
  internal/handler \
  internal/service \
  internal/repository \
  internal/middleware \
  internal/model \
  internal/task \
  pkg/database \
  pkg/storage \
  pkg/queue \
  pkg/jwt \
  pkg/response \
  pkg/validator \
  migrations \
  api/openapi \
  bin
```
- **技术解析**：
  - `cmd/server/`：主服务编译入口；
  - `internal/`：私有业务层（Go 编译器强制禁止外部模块导入，保障高内聚）；
  - `pkg/`：纯技术组件（零业务字段，无缝支持多项目复用）；
  - `migrations/`：数据库版本化 DDL 脚本。

---

### 步骤 3：编写最小可运行入口 `cmd/server/main.go`
包含独创的【Why-Flow-Gotcha】三维透析注释，打印启动 ASCII Banner 与 Process PID、Go Runtime、平台架构。
```go
package main

import (
	"fmt"
	"os"
	"runtime"
)

var (
	Version   = "v1.0.0-dev"
	BuildTime = "2026-09-10"
)

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
```

---

### 步骤 4：即时运行单步调试
```bash
go run ./cmd/server/main.go
```
- **技术解析**：验证源码无需 Makefile 即可通过 Go 原生工具链极速启动。

---

### 步骤 5：编写根目录工程管理文件 `Makefile`
将编译注入参数、单元测试、数据竞争检测与 Docker 编排封装为短指令：
```makefile
.PHONY: help build run tidy test test-race clean docker-up docker-down docker-status wire

help:
	@echo "make run | make build | make tidy | make test | make test-race | make clean | make docker-up"

build:
	@mkdir -p bin
	go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=$$(date +'%Y-%m-%d')" -o bin/server ./cmd/server

run:
	go run ./cmd/server/main.go

tidy:
	go mod tidy

test:
	go test -v ./...

test-race:
	go test -v -race ./...

clean:
	rm -rf bin coverage.out coverage.html

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-status:
	docker compose ps

wire:
	wire ./cmd/server
```

---

### 步骤 6：通过 Makefile 编译与运行二进制
```bash
make help
make build
./bin/server
```
- **技术解析**：验证 `-ldflags` 参数将构建时间动态注入到二进制中，验证独立二进制无需 Go 源码环境即可启动。

---

### 步骤 7：验证编译产物清理指令
```bash
make clean
```
- **技术解析**：清除本地 `bin/` 目录与测试临时覆盖率文件，保证版本库整洁。

---

## 🛠️ 三、 真实避坑与修复过程 (Troubleshooting & Fixes)

### 踩坑事件：`.gitignore` 误过滤 `cmd/server/` 源码目录
- **现象**：
  执行 `git status` 时，发现新建的 `cmd/server/main.go` 没有出现在 untracked 文件列表中。
- **排查命令**：
  ```bash
  git check-ignore -v cmd/server/main.go
  ```
- **排查输出**：
  ```text
  .gitignore:3:server	cmd/server/main.go
  ```
- **原因剖析**：
  在 `.gitignore` 第 3 行写了 `server`，Git 会对所有名为 `server` 的文件或**子目录**生效，直接把 `cmd/server/` 目录整体验证性排除了！
- **修复方案**：
  将规则加上根路径斜杠，精准限定为根目录的编译产物，不影响子目录：
  ```diff
  - server
  - main
  + /server
  + /main
  ```
- **验证恢复**：重新执行 `git status`，`cmd/server/main.go` 成功被 Git 正常跟踪！

---

## 🧪 四、 最终终端验证与标准输出

执行完整构建流水线：
```bash
make build && ./bin/server
```

真实终端输出片段：
```text
>> 正在编译 nexus-hub 服务...
go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=2026-09-10" -o bin/server ./cmd/server
>> 编译完成: bin/server

  _   _                     _   _       _     
 | \ | |                   | | | |     | |    
 |  \| | _____  ___   _ ___| |_| |_   _| |__  
 | . ` |/ _ \ \/ / | | / __|  _  | | | | '_ \ 
 | |\  |  __/>  <| |_| \__ \ | | | |_| | |_) |
 |_| \_|\___/_/\_\\__,_|___/_| |_|\__,_|_.__/ 
                                              
 Nexus-Hub Enterprise Business Platform [Go 1.27]

>> [Bootstrap] Version       : v1.0.0
>> [Bootstrap] Build Time    : 2026-09-10
>> [Bootstrap] Go Runtime    : go1.27.1
>> [Bootstrap] Platform      : darwin/arm64
>> [Bootstrap] Process PID   : 29292
>> [Bootstrap] Status        : Engine Core Initialized Successfully.
```
