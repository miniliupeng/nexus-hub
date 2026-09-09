# 通用规格 06：前端协同契约、领域错误码与交付验收标准 (Front-End Alignment & DoD)

## 📌 1. 前端协同原则与核心痛点解决

作为面向现代化微服务与前后端分离架构的商业级业务中台，Nexus-Hub 严格遵循契约优先（Contract-First）标准，解决客户端与服务端的四大经典协作痛点：
1. **统一返回体与 TypeScript 类型定义无缝映射**；
2. **生产级跨域 (CORS) 策略，完美兼容复杂 Header 与 OPTIONS 预检请求**；
3. **提供开箱即用的 Axios 双 Token 静默刷新拦截器**；
4. **大文件与资源对象存储预签名直传（Pre-signed URL），彻底免去传统多段转存的网络阻塞**；
5. **强类型领域错误码体系，支持前端做针对性精准 UI 提示（非简单弹窗模糊报错）**。

---

## 📐 2. 统一前后端交互响应体 (ApiResponse Specification)

后端所有控制器（Handler）禁止随意输出裸数据，统一遵循以下契约：

```typescript
/**
 * 标准全局 API 响应接口定义
 */
export interface ApiResponse<T = any> {
  code: number;        // 业务状态码 (参见下文领域错误码字典)
  message: string;     // 人类可读的提示信息 (用于前端 Toast / Message.error 弹窗)
  data: T;             // 具体业务泛型数据载荷
}

/**
 * 通用标准分页数据结构体 (传统分页)
 */
export interface PageResult<T> {
  total: number;       // 总记录数
  page: number;        // 当前页码
  page_size: number;   // 每页容量
  list: T[];           // 列表数据集
}

/**
 * 高性能游标分页结构体 (用于移动端瀑布流/无限滚动/深分页)
 */
export interface CursorPageResult<T> {
  list: T[];
  next_cursor?: number; // 下一页的 last_id，若为空说明到底了
  has_more: boolean;    // 是否还有更多
}
```

---

## 📖 3. 领域统一业务错误码字典 (Domain Error Codes)

```typescript
export enum BusinessErrorCode {
  SUCCESS = 200,

  // 10000 ~ 10999: 身份认证与用户域 (Auth & User Domain)
  USER_NOT_FOUND = 10001,          // 用户不存在
  USER_ALREADY_EXISTS = 10002,      // 用户名或邮箱已注册
  PASSWORD_INVALID = 10003,         // 密码错误
  TOKEN_EXPIRED = 10004,            // Access Token 过期 (触发 401 静默刷新)
  REFRESH_TOKEN_INVALID = 10005,    // Refresh Token 失效 (必须重新登录)
  PERMISSION_DENIED = 10006,        // 角色权限不足 (403 禁止越权)

  // 20000 ~ 20999: 资产与内容域 (Article & Asset Domain)
  ARTICLE_NOT_FOUND = 20001,        // 文章资产不存在或已软删除
  ARTICLE_STATUS_INVALID = 20002,   // 非法状态流转 (如已发布的文章不能二次提交草稿)
  ARTICLE_AUDIT_REJECTED = 20003,   // 文章包含违规敏感词被系统驳回
  ARTICLE_FORBIDDEN_DELETE = 20004, // 只能由作者本人或管理员删除
  ARTICLE_ALREADY_LIKED = 20005,    // 重复点赞防刷

  // 30000 ~ 30999: 通用基础设施与网络域 (Infra & Network Domain)
  VALIDATION_FAILED = 30001,        // 字段校验不通过 (请求体格式错误)
  IDEMPOTENT_REPEATED = 30002,      // 触发展单防重复提交 (请勿短时间连点)
  UPLOAD_TOO_LARGE = 30003,         // 上传文件超出最大允许体积
  SYSTEM_BUSY = 30004,              // 系统繁忙 / 服务降级保护
}
```

---

## 💻 4. 开箱即用的前端 Axios 拦截器完整实现 (TypeScript)

前端工程师可直接将以下代码复制进自身的前端项目中与 Nexus-Hub 后端对接：

```typescript
import axios, { AxiosInstance, AxiosResponse } from 'axios';

// 创建独立客户端实例
const client: AxiosInstance = axios.create({
  baseURL: 'http://127.0.0.1:8088/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

let isRefreshing = false;
let requestsQueue: Array<(token: string) => void> = [];

// 1. 请求拦截器: 自动注入当前存活的 AccessToken
client.interceptors.request.use((config) => {
  const accessToken = localStorage.getItem('access_token');
  if (accessToken && config.headers) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

// 2. 响应拦截器: 拦截 401 并触发无感静默刷新
client.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data;
    if (res.code !== 200) {
      return Promise.reject(new Error(res.message || 'Error'));
    }
    return res;
  },
  async (error) => {
    const originalRequest = error.config;

    // 捕获 401 且该请求未被重试过
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve) => {
          requestsQueue.push((newToken: string) => {
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            resolve(client(originalRequest));
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = localStorage.getItem('refresh_token');
      if (!refreshToken) {
        localStorage.clear();
        window.location.href = '/login';
        return Promise.reject(error);
      }

      try {
        const refreshResponse = await axios.post('http://127.0.0.1:8088/api/v1/auth/refresh', {
          refresh_token: refreshToken,
        });

        const newAccessToken = refreshResponse.data.data.access_token;
        localStorage.setItem('access_token', newAccessToken);

        requestsQueue.forEach((cb) => cb(newAccessToken));
        requestsQueue = [];

        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
        return client(originalRequest);
      } catch (refreshErr) {
        localStorage.clear();
        window.location.href = '/login';
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error.response?.data || error);
  }
);

export default client;
```

---

## 🎯 5. 交付验收标准 (Definition of Done - DoD)

| 验收维度 | 具体达标指标与测试手段 | 责任层 |
| :--- | :--- | :---: |
| **架构规范** | 1. 100% 消除任何包级全局变量 (`global.DB`)，全部采用 Google Wire 编译期注入装配。<br>2. 目录完全符合 Clean Architecture 分层 (`cmd/`, `internal/handler`, `internal/service`, `internal/repository`, `internal/task`)。 | 架构层 |
| **安全与 RBAC** | 具备完整的 RBAC 鉴权拦截器与接口幂等性防刷（`X-Idempotency-Key`）。 | 安全控制 |
| **数据库性能** | 1. 以 **MySQL 8.0** 为生产标准，拥有完整的 `.up.sql` 与 `.down.sql` 迁移脚本。<br>2. 支持一键切换 **SQLite 本地免安装零依赖双模式**。<br>3. 支持游标深分页（Keyset Pagination）优化防慢查询。 | 仓储层 |
| **对象存储直传** | 1. 提供 `/api/v1/storage/presigned-url` 接口签发有时效限制的 MinIO/S3 上传凭证。<br>2. 杜绝前端大文件穿透后端服务器，网络带宽节省 90% 以上。 | 存储层 |
| **异步任务解耦** | 1. 基于 Asynq 实现文章发布后的敏感词审核与阅读量合并刷盘。<br>2. 具备重试回退与死信队列机制，保证高可用。 | 异步队列 |
| **并发与质量** | 执行 `go test -race ./...` 保证 **100% 零数据竞争 (Zero Data Race)**。 | 质量门禁 |
