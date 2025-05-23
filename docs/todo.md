# HA-MI 开发计划

本文档跟踪 HA-MI 项目的开发进度，记录已完成和待完成的任务。

## 已完成 ✅

### 基础框架
- ✅ 项目结构搭建
- ✅ Go 模块初始化
- ✅ 配置加载（支持 YAML/JSON）
- ✅ SQLite 数据库集成
- ✅ JWT 认证实现
- ✅ API 安全机制（nonce, timestamp, sign）
- ✅ API 路由和中间件设置

## 进行中 🔄

### 核心功能开发

## 待完成 📝

### 第一阶段：基础功能

#### Home Assistant 集成
- [ ] Home Assistant 客户端实现
  - [ ] API 调用封装
  - [ ] 状态查询功能
  - [ ] 服务调用功能
  - [ ] WebSocket 状态订阅

#### 中央控制器模拟
- [ ] 控制器设备类型定义
  - [ ] 智能家居控制中心
  - [ ] 媒体控制中心
  - [ ] 环境控制中心
  - [ ] 场景控制中心
- [ ] MIoT Spec V2 设备描述实现
- [ ] 本地设备发现服务 (mDNS/SSDP)
- [ ] 设备认证和在线状态管理

#### 命令路由系统
- [ ] 命令解析模块
- [ ] 区域-设备-操作映射表实现
- [ ] 命令参数转换机制
- [ ] 命令路由执行引擎

#### 数据管理
- [ ] 区域数据模型和CRUD操作
- [ ] 设备类型数据模型和CRUD操作
- [ ] 操作数据模型和CRUD操作
- [ ] 映射关系数据模型和CRUD操作

#### API 实现
- [ ] 区域管理 API
- [ ] 设备类型管理 API
- [ ] 操作管理 API
- [ ] 映射关系管理 API
- [ ] 设备状态查询 API
- [ ] 设备控制 API

### 第二阶段：功能完善

#### 场景管理
- [ ] 场景数据模型和CRUD操作
- [ ] 场景定义语法实现
- [ ] 场景执行引擎
- [ ] 场景管理 API

#### Web 管理界面
- [ ] 前端框架设置 (Vue 3 + Element Plus)
- [ ] 认证与授权界面
- [ ] 区域和设备管理界面
- [ ] 映射关系配置界面
- [ ] 场景编辑器界面
- [ ] 系统设置界面

#### 状态反馈机制
- [ ] 实时状态更新
- [ ] 设备状态变化通知
- [ ] 命令执行结果反馈

### 第三阶段：优化和扩展

#### 性能优化
- [ ] 命令执行性能优化
- [ ] 数据库查询优化
- [ ] 并发处理优化

#### 扩展功能
- [ ] 多语言支持
- [ ] 更多设备类型支持
- [ ] 高级场景条件支持
- [ ] 语音命令模式识别改进

#### 部署方案
- [ ] Docker 容器化支持
- [ ] systemd 服务配置
- [ ] Home Assistant 插件形式打包

## 测试计划

### 单元测试
- [ ] 配置模块测试
- [ ] 认证模块测试
- [ ] 数据库模块测试
- [ ] API 模块测试
- [ ] 命令路由模块测试

### 集成测试
- [ ] API 端到端测试
- [ ] 与 Home Assistant 交互测试
- [ ] 与小爱同学交互测试

### 性能测试
- [ ] 命令处理性能测试
- [ ] 并发请求处理测试

## 文档计划

- [ ] API 接口文档
- [ ] 部署指南
- [ ] 用户配置指南
- [ ] 开发者文档 