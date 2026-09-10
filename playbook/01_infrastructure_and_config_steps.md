# 实操剧本 01：基础设施编排与强类型配置引擎 (Infrastructure & Config Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 1 (Module 01) 阶段从零到一敲下的**每一行命令、执行时机、真实避坑排查与最终终端输出**。
> 涵盖 Docker Compose 全套生产级中间件配置、YAML 强类型反序列化与启动期防御性前置校验。

---

## 🎯 一、 核心目标与前置条件

- **核心目标**：
  1. 编写根目录 `docker-compose.yml`，定义 MySQL 8.0、Redis 7.2、ETCD v3.5、MinIO S3 容器网络与持久化卷；
  2. 引入 `gopkg.in/yaml.v3` 高性能解析库；
  3. 编写 `configs/config.yaml` 运行参数与 `configs/config.go` 强类型映射模型；
  4. 实现 `Validate()` 防御性前置校验（端口范围、驱动合法性、JWT 密钥强度）；
  5. 编写全套单元测试 `configs/config_test.go` 并挂载至 `cmd/server/main.go`。
- **前置环境**：
  - 已完成 Module 00（`go.mod` 与 Standard Layout 已就绪）。

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：编写根目录 `docker-compose.yml`
为本地全套中间件定义容器拓扑：
```yaml
# 包含 mysql:8.0.36 (3306), redis:7.2-alpine (6379), bitnami/etcd:3.5 (2379), minio/minio (9000, 9001)
```
- **技术解析**：
  - MySQL 开启 `utf8mb4_unicode_ci` 字符集与原生密码插件兼容性；
  - Redis 开启 AOF 持久化（`--appendonly yes`）；
  - MinIO 暴露 9000 (S3 API) 与 9001 (Web 控制台)，配置数据卷持久化。

---

### 步骤 2：引入轻量级 YAML 3.0 解析依赖
```bash
go get gopkg.in/yaml.v3
```
- **技术解析**：相比重型的第三方配置库，`gopkg.in/yaml.v3` 具有零外部依赖、极低内存分配与标准 struct tag 映射特性。

---

### 步骤 3：编写环境配置文件 `configs/config.yaml`
声明服务运行参数：端口 8088、数据库驱动 `mysql` 与 DSN、Redis 连接参数、ETCD 节点与 MinIO 密钥。

---

### 步骤 4：编写强类型配置引擎 `configs/config.go`
包含【Why-Flow-Gotcha】三维透析注释，并设计防御性前置校验 `Validate()`：
```go
func Load(configPath string) (*Config, error) { ... }
func (c *Config) Validate() error { ... }
```
- **校验边界**：
  - 端口必须在 `1 ~ 65535` 区间；
  - 数据库驱动仅放行 `mysql` 或 `sqlite`；
  - 数据库 DSN 与 Redis 地址不能为空；
  - JWT Secret 长度必须 `>= 16` 字节。

---

### 步骤 5：编写独立单元测试 `configs/config_test.go`
覆盖正常加载用例与 5 组非法配置边界拦截用例（端口为0、超限端口、非法驱动、空 DSN、短密钥）。

---

### 步骤 6：将配置引擎挂载进主入口 `cmd/server/main.go`
服务启动第一步调用 `configs.Load("")`，若发生错误向 stderr 输出并触发 `os.Exit(1)` 立即熔断；若成功则安全脱敏输出配置摘要。

---

## 🛠️ 三、 真实避坑与修复过程 (Troubleshooting & Fixes)

### 避坑点 1：配置单测相对路径定位失败问题
- **现象**：在根目录下执行 `go test ./...` 时，`configs` 包的测试文件如果写死当前工作目录为 `configs/config.yaml`，会报错找不到文件（因为单测运行时的当前工作目录即为 `configs/` 内部）。
- **解法**：
  在 `config_test.go` 中通过路径探测逻辑动态回退：
  ```go
  cfgPath := filepath.Join("..", "configs", "config.yaml")
  if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
      cfgPath = "config.yaml"
  }
  ```
  保证无论在根目录执行 `go test ./...` 还是在子目录执行 `go test .`，均能 100% 准确定位配置文件。

### 避坑点 2：生产密钥敏感信息脱敏
- **避坑准则**：在 `main.go` 打印启动信息时，严禁输出明文 `jwt.secret` 和数据库密码，仅输出端口、驱动名与脱敏后的节点地址，规避企业安全审计红线。

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
