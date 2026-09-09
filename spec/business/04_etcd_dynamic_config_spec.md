# Nexus-Hub 规格说明书 03：ETCD 毫秒级动态配置热更规格 (ETCD Dynamic Config Spec)

## 📌 1. 业务痛点与技术诉求

传统的微服务配置大多固化在 `config.yaml` 文件中，一旦业务方需要调整：
- “系统临时维护公告 (Announcement)”；
- “紧急关闭新用户注册入口 (AllowRegistration)”；
- “修改全局文件上传体积上限 (MaxUploadSizeMB)”。

若每次修改都需要经过重新打 Docker 镜像、发布 Pod、重启容器，会导致长达数分钟的发布周期与长连接中断。
**基于 ETCD v3 的分布式配置中心支持以毫秒级时延将变更推送到所有运行中的 Go 进程内存中，且无需重启服务**。

---

## 🏗️ 2. 核心机制与原子内存指针流转拓扑

```text
       [ 管理员在后台提交动态配置修改 ]
                     │
                     ▼
          [ PUT /api/v1/configs/dynamic ]
                     │
                     ▼
        [ 写入 ETCD 集群 Key: /nexus-hub/config/dynamic ]
                     │
                     ├──────────────(ETCD Raft 复制与分发)──────────────┐
                     ▼                                                ▼
     [ Nexus-Hub 实例 A (Goroutine) ]                 [ Nexus-Hub 实例 B (Goroutine) ]
     • clientv3.Watch 监听事件触发                    • clientv3.Watch 监听事件触发
     • 解析最新 JSON 数据                             • 解析最新 JSON 数据
     • atomic.Pointer[DynamicConfig].Store(newCfg)   • atomic.Pointer[DynamicConfig].Store(newCfg)
                     │                                                │
       (完全免锁、0 性能损耗原子替换)                   (完全免锁、0 性能损耗原子替换)
                     ▼                                                ▼
     [ 业务请求直接读取内存指针，即刻生效 ]            [ 业务请求直接读取内存指针，即刻生效 ]
```

---

## 📋 3. 动态配置数据结构契约

### 3.1 ETCD 存储 Key 规范
- **Key 路径**：`/nexus-hub/config/dynamic`
- **数据格式**：标准 JSON 字符串
- **JSON Payload 结构契约**：
  ```json
  {
    "allow_registration": true,           // 布尔值，是否允许外部注册新账号
    "max_upload_size_mb": 25,              // 整型，单文件上传最大限制 (MB)
    "announcement": "系统将于今晚 24:00 进行常规网络演练",  // 字符串，全局运营公告
    "maintenance_mode": false             // 布尔值，是否开启只读维护模式
  }
  ```

---

## 🛠️ 4. 详细接口设计与数据契约

### 4.1 获取当前生效的动态配置 `GET /api/v1/configs/dynamic`
- **权限要求**：公开接口，无需 JWT
- **响应体 (200 OK)**：
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "allow_registration": true,
      "max_upload_size_mb": 25,
      "announcement": "系统将于今晚 24:00 进行常规网络演练",
      "maintenance_mode": false,
      "updated_at": "2026-09-09T16:20:00Z"
    }
  }
  ```

---

### 4.2 修改动态配置 `PUT /api/v1/configs/dynamic` (Admin 专享)
- **请求 Header**：`Authorization: Bearer <admin_access_token>`
- **权限校验**：拦截器校验当前用户角色，非 `admin` 返回 `403 Forbidden`。
- **请求体**：
  ```json
  {
    "allow_registration": false,
    "max_upload_size_mb": 50,
    "announcement": "【紧急通知】目前新用户注册入口已临时关闭维护！"
  }
  ```
- **后端执行步骤**：
  1. 校验请求参数有效性（如 `max_upload_size_mb` 必须在 1~100 之间）；
  2. 调用 `etcdClient.Put(ctx, "/nexus-hub/config/dynamic", jsonString)`；
  3. 记录运维审计日志：操作者 ID、操作时间、变更前后 Diff；
  4. 返回修改成功响应。
