# Movie Reservation System (MRS)

## 目录
1.  [项目概述](#项目概述)
2.  [核心功能](#核心功能)
    *   [用户认证与授权](#用户认证与授权)
    *   [电影管理 (管理员)](#电影管理-管理员)
    *   [预订管理](#预订管理)
    *   [报告功能 (管理员)](#报告功能-管理员)
3.  [技术栈 (初步设想)](#技术栈-初步设想)
4.  [数据模型 (初步设想)](#数据模型-初步设想)
5.  [API 端点设计 (初步设想)](#api-端点设计-初步设想)
6.  [关键设计考量](#关键设计考量)
7.  [项目搭建与运行 (占位符)](#项目搭建与运行-占位符)
    *   [环境要求](#环境要求)
    *   [安装步骤](#安装步骤)
    *   [运行应用](#运行应用)
    *   [运行测试](#运行测试)
8.  [后续步骤与未来增强](#后续步骤与未来增强)
9.  [贡献指南 (占位符)](#贡献指南-占位符)

## 1. 项目概述

Movie Reservation System (MRS) 是一个后端服务，旨在提供一个完整的电影票在线预订解决方案。用户可以注册账户、登录、浏览电影信息、查询特定日期的放映场次、选择座位并完成预订。系统还包含管理员功能，用于管理电影、放映时间以及查看系统运营数据。

本项目的主要目标是：
*   实现复杂的业务逻辑，特别是座位预订与调度。
*   深入理解数据模型设计及其关系。
*   构建健壮的用户认证和授权机制。
*   实践复杂查询和数据报告的生成。

## 2. 核心功能

### 用户认证与授权
*   **用户注册**：新用户可以创建账户。
*   **用户登录**：已注册用户可以登录系统。
*   **角色管理**：
    *   **普通用户 (User)**：可以浏览电影、预订场次、查看和取消自己的预订。
    *   **管理员 (Admin)**：拥有普通用户的所有权限，并能管理电影、放映时间、查看所有预订和系统报告。
*   **初始管理员**：系统将通过种子数据（Seed Data）创建初始管理员账户。
*   **权限提升**：只有管理员可以将其他普通用户提升为管理员。

### 电影管理 (管理员)
*   **添加电影**：管理员可以添加新的电影信息。
    *   属性：标题 (Title)、描述 (Description)、海报图片 (Poster Image URL)。
*   **更新电影**：管理员可以修改已存在的电影信息。
*   **删除电影**：管理员可以删除电影。
*   **电影分类**：电影应按类型 (Genre) 进行分类（例如：动作、喜剧、科幻）。
*   **放映时间管理**：
    *   管理员可以为电影安排放映时间 (Showtimes)。
    *   每个放映时间关联到特定的电影、影厅和时间。

### 预订管理
*   **浏览电影与场次**：用户可以获取特定日期的电影列表及其可用的放映时间。
*   **座位预订**：
    *   用户可以查看特定放映时间的座位图 (Seat Map) 及可用座位。
    *   用户可以选择一个或多个座位进行预订。
    *   系统需防止超额预订。
*   **查看我的预订**：用户可以查看自己所有已预订的电影票。
*   **取消预订**：用户可以取消尚未开始的预订。

### 报告功能 (管理员)
*   **查看所有预订**：管理员可以查看系统中所有的预订记录。
*   **上座率/容量报告**：管理员可以查看各场次的上座率和影院容量使用情况。
*   **收入报告**：管理员可以查看基于预订产生的收入情况。

## 3. 技术栈 (初步设想)

*   **编程语言**: [待定 - 例如：Python, Java, Node.js, Go]
*   **框架**: [待定 - 例如：Django/Flask (Python), Spring Boot (Java), Express.js (Node.js)]
*   **数据库**: 关系型数据库 (推荐 PostgreSQL 或 MySQL)
*   **认证机制**: [待定 - 例如：JWT (JSON Web Tokens), OAuth 2.0]
*   **API 类型**: RESTful API
*   **(可选) ORM**: [待定 - 例如：SQLAlchemy (Python), Hibernate (Java), TypeORM (Node.js)]
*   **(可选) 容器化**: Docker

选择具体技术栈时，将优先考虑开发效率、社区支持、性能以及团队熟悉度。

## 4. 数据模型 (初步设想)

以下是核心实体及其关系的基本构想：

*   `User (id, username, password_hash, email, role_id)`
*   `Role (id, name)` (e.g., 'ADMIN', 'USER')
*   `Movie (id, title, description, poster_image_url, duration_minutes)`
*   `Genre (id, name)`
*   `MovieGenre (movie_id, genre_id)` (多对多关系)
*   `CinemaHall (id, name, total_seats)` (影厅)
*   `Showtime (id, movie_id, hall_id, start_time, end_time, price_per_seat)`
*   `Seat (id, hall_id, row_identifier, seat_number, type)` (物理座位，可用于复杂座位图)
*   `Booking (id, user_id, showtime_id, booking_time, total_amount, status)` (e.g., 'CONFIRMED', 'CANCELLED', 'PENDING')
*   `BookedSeat (id, booking_id, showtime_id, seat_id)` (或者 `row_identifier`, `seat_number` 如果座位不是全局唯一ID)

此数据模型会随着开发的深入进行迭代和优化。

## 5. API 端点设计 (初步设想)

以下是一些关键的 API 端点（路径和方法可能会调整）：

**认证 (Auth)**
*   `POST /api/auth/register` - 用户注册
*   `POST /api/auth/login` - 用户登录
*   `POST /api/auth/logout` - 用户登出 (如果使用服务端 session 或 token 黑名单)

**用户 (Users)**
*   `GET /api/users/me` - 获取当前用户信息
*   `PUT /api/admin/users/{userId}/promote` - (Admin) 提升用户为管理员

**电影 (Movies)**
*   `GET /api/movies` - 获取所有电影列表 (可带分页、筛选、排序参数)
*   `GET /api/movies/{movieId}` - 获取特定电影详情
*   `POST /api/admin/movies` - (Admin) 添加新电影
*   `PUT /api/admin/movies/{movieId}` - (Admin) 更新电影信息
*   `DELETE /api/admin/movies/{movieId}` - (Admin) 删除电影
*   `GET /api/genres` - 获取所有电影类型

**放映时间 (Showtimes)**
*   `GET /api/showtimes?date={YYYY-MM-DD}[&movieId={id}]` - 获取特定日期（可选电影）的放映时间
*   `GET /api/showtimes/{showtimeId}/seats` - 获取特定场次的座位图及可用状态
*   `POST /api/admin/movies/{movieId}/showtimes` - (Admin) 为电影添加放映时间
*   `PUT /api/admin/showtimes/{showtimeId}` - (Admin) 更新放映时间
*   `DELETE /api/admin/showtimes/{showtimeId}` - (Admin) 删除放映时间

**预订 (Bookings)**
*   `POST /api/bookings` - 用户创建新预订 (请求体包含 `showtimeId`, `seatIds`)
*   `GET /api/bookings/me` - 用户查看自己的预订列表
*   `GET /api/bookings/{bookingId}` - 用户查看特定预订详情
*   `PATCH /api/bookings/{bookingId}/cancel` - 用户取消预订 (仅限未开始的)
*   `GET /api/admin/bookings` - (Admin) 查看所有用户预订 (可带分页、筛选参数)

**报告 (Reports) (Admin)**
*   `GET /api/admin/reports/revenue?startDate={date}&endDate={date}` - 获取指定时间范围内的收入报告
*   `GET /api/admin/reports/occupancy?showtimeId={id}` - 获取特定场次的上座率
*   `GET /api/admin/reports/movie-performance` - 获取电影表现报告

## 6. 关键设计考量

*   **数据一致性与并发控制**：座位预订是核心功能，必须确保在高并发场景下不会发生超额预订或数据冲突。可能需要使用数据库事务、行级锁或乐观锁机制。
*   **座位管理**：如何高效地表示和查询座位状态（可用、已预订、锁定）。是为每个 `Showtime` 动态生成座位状态，还是持久化每个 `Seat` 在每个 `Showtime` 的状态？
*   **放映时间调度**：如何设计调度逻辑，避免同一影厅在同一时间段安排多个场次。
*   **认证与授权**：选择合适的认证方案 (如 JWT)，并实现精细的基于角色的授权控制。
*   **可扩展性**：系统设计应考虑到未来可能的扩展，如引入支付网关、消息通知（邮件/短信）、更复杂的座位类型和价格策略等。
*   **错误处理与日志记录**：健壮的错误处理机制和全面的日志记录对于问题排查和系统监控至关重要。
*   **性能优化**：对于频繁查询的接口（如查询电影、场次），需要考虑数据库索引和缓存策略。

## 7. 项目搭建与运行 (占位符)

本部分将在项目初始化后填充具体的指令。

### 环境要求
*   [编程语言版本]
*   [数据库类型及版本]
*   [其他依赖，如 Node.js, pip, Maven 等]

### 安装步骤
```bash
# 1. 克隆仓库
git clone [repository-url]
cd movie-reservation-system

# 2. 安装依赖
# (示例，根据技术栈调整)
# pip install -r requirements.txt
# npm install
# mvn install

# 3. 配置环境变量
# cp .env.example .env
# (编辑 .env 文件配置数据库连接、密钥等)

# 4. 数据库迁移 (如果使用 ORM)
# python manage.py migrate
# npx sequelize-cli db:migrate
