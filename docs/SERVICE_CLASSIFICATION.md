
# 电影预定系统服务划分

## 1. AuthService (认证服务)
- **职责**:
  - 用户登录认证
  - JWT令牌生成与校验
  - 密码安全处理（哈希/验证）
- **划分理由**:
  - 核心安全功能需要独立封装
- **对应功能**:
  - 用户登录

## 2. UserService (用户服务)
- **职责**:
  - 用户账户管理（注册/查询/更新）
  - 角色分配与管理
  - 用户-角色关联维护
- **划分理由**:
  - 用户信息管理独立于认证业务
  - 权限控制的基础支撑
- **对应功能**:
  - 用户注册
  - 角色管理
  - 权限控制基础

## 3. MovieService (电影服务)
- **职责**:
  - 电影实体CRUD管理
  - 电影类型实体CRUD管理
  - 电影/类型查询接口（前后台）
- **划分理由**:
  - 电影核心元素高内聚
  - 独立于场馆/放映等动态业务
- **对应功能**:
  - 电影信息管理（管理员）
  - 电影浏览（用户）

## 4. CinemaService (场馆服务)
- **职责**:
  - 影厅实体CRUD管理
  - 影厅静态座位布局配置
  - 影厅/座位布局查询
- **划分理由**:
  - 基础设施信息相对静态
  - 变更频率低于放映计划
- **对应功能**:
  - 影厅与座位管理（管理员）

## 5. ShowtimeService (放映服务)
- **职责**:
  - 放映计划CRUD管理（关联电影/影厅）
  - 排期冲突检测
  - 场次查询（电影/日期/影厅维度）
  - 动态座位图及实时状态提供
- **划分理由**:
  - 连接电影和影厅的动态业务
  - 高实时性要求（座位状态）
- **对应功能**:
  - 放映计划管理（管理员）
  - 场次浏览（用户）
  - 座位图展示（用户）

## 6. BookingService (预订服务)
- **职责**:
  - 座位锁定（含并发控制）
  - 订单创建与管理
  - 订单查询/取消
- **划分理由**:
  - 核心交易流程独立封装
  - 复杂业务逻辑（状态机/分布式锁）
  - 未来支付集成扩展点
- **对应功能**:
  - 下单处理（用户）
  - 订单管理（用户）

## 7. ReportService (报告服务)
- **职责**:
  - 跨域数据聚合（预订/场次/电影）
  - 运营报表生成
  - 数据分析接口
- **划分理由**:
  - 分析功能与核心业务解耦
  - 避免数据聚合污染服务边界
- **对应功能**:
  - 报告与分析（管理员）

## 服务依赖关系
```mermaid
graph TD
    A[AuthService] -->|提供认证| B[UserService]
    B -->|用户数据| F[BookingService]
    C[MovieService] -->|电影数据| E[ShowtimeService]
    D[CinemaService] -->|影厅数据| E
    E -->|场次数据| F[BookingService]
    F -->|订单数据| G[ReportService]
    E -->|座位状态| G