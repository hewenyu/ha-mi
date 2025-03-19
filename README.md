# HA-MI: Home Assistant 与小米智能家居桥接服务

HA-MI 是一个基于 Go 语言开发的桥接服务，采用类似 VRF 中央控制器的模拟方案，实现小爱同学与 Home Assistant 的互通。项目的主要特点是：

- **无需小爱开发者平台**：不依赖小米开放平台的自定义技能
- **中央控制器模型**：模拟单个或少量虚拟中央设备，简化发现和控制
- **统一命令处理**：基于区域和设备类型的统一命令结构
- **场景和分组控制**：支持复杂的跨设备场景和分组操作
- **易于配置和扩展**：提供 Web 界面进行区域和设备映射管理

## 快速开始

### 构建

```bash
# 克隆仓库
git clone https://github.com/boringsoft/ha-mi.git
cd ha-mi

# 构建项目
go build -o ha-mi cmd/server/main.go
```

### 运行

```bash
# 使用默认配置
./ha-mi

# 使用指定配置文件
./ha-mi -config=/path/to/config.json
```

## 配置文件

配置文件支持 YAML 和 JSON 格式，默认使用 YAML 格式。默认配置如下：

### YAML 格式 (config.yaml)

```yaml
server:
  host: 0.0.0.0
  port: 8080

auth:
  user: admin
  password: admin
  secret_key: change-me-in-production-please
  access_token_expiry: 86400000000000  # 24 hours in nanoseconds
  refresh_token_expiry: 2592000000000000  # 30 days in nanoseconds
  nonce_expiry: 120000000000  # 2 minutes in nanoseconds

database:
  path: ha-mi.db

home_assistant:
  url: http://localhost:8123
  token: ""  # Put your Home Assistant long-lived access token here
```

### JSON 格式 (config.json)

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "auth": {
    "user": "admin",
    "password": "admin",
    "secret_key": "change-me-in-production-please",
    "access_token_expiry": 86400000000000,
    "refresh_token_expiry": 2592000000000000,
    "nonce_expiry": 120000000000
  },
  "database": {
    "path": "ha-mi.db"
  },
  "home_assistant": {
    "url": "http://localhost:8123",
    "token": ""
  }
}
```

配置项说明：

- **server**: 服务器配置
  - `host`: 服务器监听地址
  - `port`: 服务器监听端口

- **auth**: 认证配置
  - `user`: 用户名
  - `password`: 密码
  - `secret_key`: JWT 密钥
  - `access_token_expiry`: 访问令牌有效期（纳秒）
  - `refresh_token_expiry`: 刷新令牌有效期（纳秒）
  - `nonce_expiry`: 随机数有效期（纳秒）

- **database**: 数据库配置
  - `path`: SQLite 数据库文件路径

- **home_assistant**: Home Assistant 配置
  - `url`: Home Assistant URL
  - `token`: Home Assistant 长效访问令牌

## API 接口

### 认证

#### 获取随机数

```
GET /api/v1/auth/nonce?timestamp=1678942800000
```

#### 登录

```
POST /api/v1/auth/login
```

请求体：

```json
{
  "username": "admin",
  "password": "admin",
  "nonce": "8f7b3c1a2e5d4f6b8a9c7d5f3e1d2c4b",
  "timestamp": "1678942800000",
  "sign": "8f7b3c1a2e5d4f6b8a9c7d5f3e1d2c4b"
}
```

#### 刷新令牌

```
POST /api/v1/auth/refresh
```

请求体：

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "nonce": "8f7b3c1a2e5d4f6b8a9c7d5f3e1d2c4b",
  "timestamp": "1678942800000",
  "sign": "8f7b3c1a2e5d4f6b8a9c7d5f3e1d2c4b"
}
```

## 安全校验

所有 API 接口都需要包含以下参数：

1. **时间戳(timestamp)**: 可通过请求头 `X-Timestamp` 或请求参数 `timestamp` 提供
   - 格式为UNIX时间戳(毫秒)，例如：`1678942800000`
   - 时间戳与服务器时间差不能超过60秒

2. **随机数(nonce)**: 可通过请求头 `X-Nonce` 或请求参数 `nonce` 提供
   - 必须通过 `/api/v1/auth/nonce` 接口获取
   - 每个 nonce 仅能使用一次，有效期为2分钟

3. **签名(sign)**: 可通过请求头 `X-Sign` 或请求参数 `sign` 提供
   - 签名算法：HMAC-SHA256
   - 签名参数包括所有请求参数(包括timestamp和nonce，但不包括sign自身)
   - 参数按键名ASCII码从小到大排序并拼接成 `key1=value1&key2=value2` 的形式

## 请求流程

1. 首先通过 `/api/v1/auth/nonce` 获取服务端生成的 nonce（此接口仅需要 timestamp）
2. 使用获取到的 nonce 构建正式 API 请求参数
3. 生成签名并提交完整请求

## 设计方案

详细设计方案请参考 [design.md](docs/design.md)。

## 开发计划

查看当前开发计划和进度请参考 [todo.md](docs/todo.md)。 