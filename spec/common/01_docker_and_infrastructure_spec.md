# Nexus-Hub 规格说明书 06：基础设施与 Docker 编排 (Infrastructure & Docker Spec)

## 📌 1. 基础设施拓扑

为保证无论开发者本机是否安装了 MySQL、Redis、ETCD 或 MinIO，均可在 30 秒内拉起全套标准化生产级拓扑，`nexus-hub` 在根目录提供开箱即用的 `docker-compose.yml`。

| 组件服务 | 镜像版本 | 容器端口 | 宿主机端口 | 默认账号 / 密码 | 业务用途 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **MySQL** | `mysql:8.0.36` | `3306` | `3306` | `root / nexus_root_2026` | 核心业务持久化（ACID 事务、联合索引） |
| **Redis** | `redis:7.2-alpine` | `6379` | `6379` | 无密码 (开发环境) | 双 Token 黑名单、Cache-Aside 缓存、Asynq 队列底座 |
| **ETCD** | `bitnami/etcd:3.5` | `2379` | `2379` | 免密 (单节点 Raft) | 服务租约注册心跳、Watch 动态参数热更新 |
| **MinIO** | `minio/minio:latest` | `9000, 9001` | `9000, 9001` | `admin / minioadmin2026` | 兼容 S3 的对象存储（API: 9000，控制台: 9001） |

---

## 🚀 2. 一键启动与效能脚本

```bash
# 一键在后台启动全套基础设施
make docker-up

# 查看运行状态与健康检查
make docker-status

# 一键清理并停止容器
make docker-down
```
