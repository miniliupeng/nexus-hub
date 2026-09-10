# 实操剧本 00：工程身份确立与最小可运行骨架搭建 (Project Init Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 0 (Module 00) 阶段从零到一敲下的**每一行命令、执行时机、真实避坑排查与最终终端输出**。
> 本手册采用自包含设计，涵盖 Git 初始化、Standard Go Layout、Makefile 伪目标与链接器注入机制、以及进程生命周期排查。

---

## 🎯 一、 核心目标与前置条件

- **核心目标**：
  1. 确立 `nexus-hub` 独立 Go 模块身份并生成 `go.mod`；
  2. 初始化独立 Git 仓库，设置企业级 `.gitignore` 并关联远程 GitHub 仓库；
  3. 搭建业界标准的 `Standard Go Project Layout` 目录树；
  4. 编写最小可运行的 `cmd/server/main.go`（包含启动 Banner 与运维元数据）；
  5. 编写工程指令枢纽 `Makefile`，打通本地编译、伪目标（.PHONY）与清理流水线。
- **前置环境**：
  - 操作系统：macOS (darwin/arm64)
  - Go 版本：`go1.27.1`
  - 远程仓库：`git@github.com:miniliupeng/nexus-hub.git`

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：初始化 Go 模块
确立根包名并锁定 Go 工具链版本。
```bash
go mod init nexus-hub
go version
```
- **技术解析**：生成根目录 `go.mod` 文件，所有子目录的 import 路径将基于 `nexus-hub/...` 解析，探测到本地环境为 `go1.27.1 darwin/arm64`。

---

### 步骤 2：初始化 Git 仓库与企业级 `.gitignore`
在根目录下创建 `.gitignore` 并完成仓库初始化与远程关联：
```bash
# 1. 编写 .gitignore
cat << 'GITIGNORE' > .gitignore
bin/
/server
/main
*.test
*.out
coverage.html
.env
.env.*
configs/*.local.yaml
*.db
*.sqlite
.idea/
.vscode/
.DS_Store
data/
minio_data/
mysql_data/
redis_data/
etcd_data/
GITIGNORE

# 2. 初始化 Git 仓库并指定主分支为 main
git init -b main
git remote add origin git@github.com:miniliupeng/nexus-hub.git
```
- **技术解析**：
  - 严密过滤编译二进制产物（`bin/`）、测试临时覆盖率文件与数据库持久化目录；
  - 重点规则：二进制过滤使用 `/server` 与 `/main`，严格规避子目录同名冲突。

---

### 步骤 3：落地 Standard Go Project Layout 目录树
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
- **各核心目录职责**：
  - `cmd/server/`：主服务唯一编译入口，保持轻薄，不塞业务逻辑；
  - `internal/`：私有业务层（Go 编译器强制禁止外部模块导入，保障封装安全性）；
  - `pkg/`：纯技术基础设施包（零业务字段，可直接跨项目复用）；
  - `migrations/`：数据库版本化 DDL 迁移脚本专用目录；
  - `configs/`：环境参数配置文件与强类型结构体定义。

---

### 步骤 4：编写最小可运行入口 `cmd/server/main.go`
植入独创的【Why-Flow-Gotcha】三维透析注释，打印启动 ASCII Banner 与 Process PID、Go Runtime、平台架构：

```go
package main

import (
	"fmt"
	"os"
	"runtime"
)

// Version 与 BuildTime 变量由 Makefile 在编译时通过 -ldflags 动态覆盖注入
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

- **技术解析**：
  - `runtime.Version()` 与 `runtime.GOOS/GOARCH`：打印当前运行时环境，防止交叉编译平台错乱；
  - `os.Getpid()`：打印当前进程的唯一标识 PID，便于系统运维排障。

---

### 步骤 5：即时运行单步调试
```bash
go run ./cmd/server/main.go
```
- **技术解析**：验证源码无需 Makefile 即可通过 Go 原生工具链极速启动。

---

### 步骤 6：编写根目录工程管理文件 `Makefile`
将编译注入参数、单元测试、数据竞争检测与 Docker 编排封装为易记的短指令：

```makefile
.PHONY: help build run tidy test test-race clean docker-up docker-down docker-status wire

# 默认展示帮助指南
help:
	@echo "=========================================================================="
	@echo "                   Nexus-Hub 工程管理效能指令中心 (Makefile)                "
	@echo "=========================================================================="
	@echo "  make run           - 运行服务主入口 (本地极速单步调试)"
	@echo "  make build         - 编译二进制文件至 bin/server"
	@echo "  make tidy          - 整理并修剪 go.mod 与 go.sum 依赖"
	@echo "  make test          - 运行全库所有单元测试"
	@echo "  make test-race     - 运行全库数据竞争并发扫描 (DoD 强制验收)"
	@echo "  make clean         - 清理本地编译产物与测试覆盖率文件"
	@echo "  make docker-up     - 一键在后台启动全套基础设施 (MySQL, Redis, ETCD, MinIO)"
	@echo "  make docker-status - 查看本地容器健康检查与端口映射状态"
	@echo "  make docker-down   - 一键安全停止并卸载所有基础设施容器"
	@echo "  make wire          - 触发 Google Wire 编译期自动依赖静态组装"
	@echo "=========================================================================="

build:
	@mkdir -p bin
	@echo ">> 正在编译 nexus-hub 服务..."
	go build -ldflags "-X main.Version=v1.0.0 -X main.BuildTime=$$(date +'%Y-%m-%d')" -o bin/server ./cmd/server
	@echo ">> 编译完成: bin/server"

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
	@echo ">> 清理完毕"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-status:
	docker compose ps

wire:
	wire ./cmd/server
```

- **Makefile 核心语法与机制详解**：
  1. **`.PHONY` 伪目标声明**：
     - Make 默认根据“磁盘上是否有同名文件”决定是否重新编译；
     - 若不声明 `.PHONY`，一旦根目录下存在名为 `build` 或 `clean` 的文件，Make 会误报 `is up to date` 并**拒绝执行任何命令**；
     - 声明 `.PHONY` 强制每次敲击均无条件执行动作。
  2. **Tab 制表符铁律**：
     - Makefile 目标下方的每一行命令缩进**必须是原生 Tab 键**，严禁使用 4 个空格，否则报 `missing separator` 语法错误。
  3. **`-ldflags "-X ..."` 链接器注入**：
     - 在编译期直接覆盖 Go 源码中的变量，`$$(date ...)` 使用双美元符号转义调用系统 Shell。

---

### 步骤 7：通过 Makefile 编译与运行二进制
```bash
make help
make build
./bin/server
```
- **技术解析**：验证独立二进制文件完全脱离源码即可在操作系统独立执行。

---

### 步骤 8：验证编译产物清理指令
```bash
make clean
```
- **技术解析**：清除本地 `bin/` 目录与测试临时文件，确保版本库整洁。

---

## 🛠️ 三、 真实避坑与修复过程 (Troubleshooting & Fixes)

### 避坑点 1：`.gitignore` 误过滤 `cmd/server/` 源码目录
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
  在 `.gitignore` 第 3 行写了 `server`，Git 默认会对所有名为 `server` 的文件或**子目录**生效，直接把 `cmd/server/` 目录整体验证性排除了！
- **修复方案**：
  将规则加上根路径斜杠，精准限定为根目录的编译产物，不影响源码子目录：
  ```diff
  - server
  - main
  + /server
  + /main
  ```
- **验证恢复**：重新执行 `git status`，`cmd/server/main.go` 成功恢复正常跟踪。

---

### 避坑点 2：为什么启动后在终端用 `ps -p <PID>` 查不到进程？
- **现象**：
  在执行 `make run` 后，控制台打印了 `Process PID : 34277`，但随后在终端执行 `ps -p 34277` 却显示进程不存在。
- **原因剖析**：
  - `func main()` 内部打印完毕后，所有指令执行完毕，**进程正常退出（Exit Code 0）**；
  - 进程退出的一瞬间，操作系统内核（Kernel）立刻销毁并回收了该 PID 及其内存；
  - 只有当服务内部存在**阻塞监听**（如随后的 HTTP Server `ListenAndServe` 或系统信号监听 `select {}`）时，服务才会常驻后台进程表中供随时抓取。

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
