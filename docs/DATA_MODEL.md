# 数据模型详解 (MRS 数据库设计)

本文档详细描述了电影预订系统 (MRS) 的核心数据表结构、字段定义以及它们之间的关系。这些表构成了系统持久化存储的基础。

## 1. `User` 表 (用户表)

*   **含义**: 存储所有已注册用户的基本信息。
*   **对应领域实体**: `internal/domain/user/user.go` 中的 `User` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增 或 VARCHAR/UUID): 用户唯一标识符。
    *   `username` (VARCHAR, 唯一索引): 用户名，用于登录，具有唯一性。
    *   `password_hash` (VARCHAR): 存储用户密码的哈希值 (例如使用 bcrypt)。**严禁存储明文密码。**
    *   `email` (VARCHAR, 唯一索引): 用户电子邮箱，可用于登录、接收通知、密码找回，具有唯一性。
    *   `role_id` (BIGINT, 外键 -> Role.id): 关联到 `Role` 表，表示该用户的角色。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。

## 2. `Role` 表 (角色表)

*   **含义**: 定义系统中的用户角色，用于权限管理。
*   **对应领域实体**: `internal/domain/user/user.go` 中的 `Role` (或作为 `User` 的一部分)。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 角色唯一标识符。
    *   `name` (VARCHAR, 唯一索引): 角色名称 (例如: 'ADMIN', 'USER')。
    *   `description` (VARCHAR, 可空): 角色描述。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。

## 3. `Movie` 表 (电影表)

*   **含义**: 存储电影的基本信息。
*   **对应领域实体**: `internal/domain/movie/movie.go` 中的 `Movie` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 电影唯一标识符。
    *   `title` (VARCHAR): 电影标题。
    *   `description` (TEXT): 电影剧情简介或描述。
    *   `poster_url` (VARCHAR, 可空): 电影海报图片的 URL 地址。
    *   `duration_minutes` (INT): 电影时长，单位为分钟。
    *   `release_date` (DATE, 可空): 上映日期。
    *   `rating` (FLOAT, 可空): 评分
    *   `age_rating` (VARCHAR, 可空): 年龄分级
    *   `cast` (TEXT): 演员表
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。

## 4. `Genre` 表 (类型表)

*   **含义**: 存储电影的类型标签，如动作、喜剧、科幻等。
*   **对应领域实体**: `internal/domain/movie/genre.go` 中的 `Genre` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 类型唯一标识符。
    *   `name` (VARCHAR, 唯一索引): 类型名称 (例如: 'Action', 'Comedy')。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。

## 5. `MovieGenre` 表 (电影类型关联表)

*   **含义**: 中间表，用于表示电影 (`Movie`) 和类型 (`Genre`) 之间的多对多关系。
*   **字段**:
    *   `movie_id` (BIGINT, 外键 -> Movie.id, 联合主键的一部分): 关联的电影 ID。
    *   `genre_id` (BIGINT, 外键 -> Genre.id, 联合主键的一部分): 关联的类型 ID。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   **约束**: `(movie_id, genre_id)` 构成联合主键，确保唯一性。

## 6. `CinemaHall` 表 (影厅表)

*   **含义**: 存储电影院中各个影厅的基本信息。
*   **对应领域实体**: `internal/domain/movie/showtime.go` 中可能作为 `Showtime` 的一部分，或独立实体 `CinemaHall`。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 影厅唯一标识符。
    *   `name` (VARCHAR): 影厅名称 (例如: "1号厅", "IMAX厅")。
    *   `capacity` (INT): 影厅总容量 (可由 `Seat` 表聚合或直接存储)。
    *   `screen_type` (VARCHAR): 屏幕类型。
    *   `sound_system` (VARCHAR): 音响系统。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。

## 7. `Seat` 表 (座位表)

*   **含义**: 定义影厅内每个物理座位的具体信息，支持复杂座位布局和类型。
*   **对应领域实体**: `internal/domain/movie/seat.go` 中的 `Seat` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 座位唯一标识符。
    *   `hall_id` (BIGINT, 外键 -> CinemaHall.id): 该座位所属的影厅 ID。
    *   `row` (VARCHAR): 座位行标识 (例如: "A", "B", 或数字 "1", "2")。
    *   `row_number` (VARCHAR): 座位在本行中的编号 (例如: "1", "2", "3")。
    *   `type` (VARCHAR, 可空): 座位类型 (例如: 'REGULAR', 'VIP', 'ACCESSIBLE')。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   **索引**: 建议在 `(hall_id, row_identifier, seat_number_in_row)` 上创建联合唯一索引。

## 8. `Showtime` 表 (放映时间表)

*   **含义**: 核心表之一，存储电影的具体放映安排。
*   **对应领域实体**: `internal/domain/movie/showtime.go` 中的 `Showtime` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 放映场次唯一标识符。
    *   `movie_id` (BIGINT, 外键 -> Movie.id): 放映的电影 ID。
    *   `hall_id` (BIGINT, 外键 -> CinemaHall.id): 放映所在的影厅 ID。
    *   `start_time` (TIMESTAMP): 放映开始时间 (日期和时间)。
    *   `end_time` (TIMESTAMP): 放映结束时间 (日期和时间)。可由 `start_time` 和电影时长计算，也可显式存储。
    *   `price_per_seat` (DECIMAL): 该场次每个座位的基准票价。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   **索引**: 建议在 `(movie_id, start_time)` 和 `(hall_id, start_time)` 上创建索引。

## 9. `Booking` 表 (预订订单表)

*   **含义**: 存储用户的预订订单信息。
*   **对应领域实体**: `internal/domain/booking/booking.go` 中的 `Booking` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增 或 VARCHAR/UUID): 预订订单唯一标识符。
    *   `user_id` (BIGINT, 外键 -> User.id): 下单用户的 ID。
    *   `showtime_id` (BIGINT, 外键 -> Showtime.id): 预订的场次 ID。
    *   `booking_time` (TIMESTAMP, 默认 CURRENT_TIMESTAMP): 订单创建时间。
    *   `total_amount` (DECIMAL): 订单总金额。
    *   `status` (VARCHAR): 订单状态 (例如: 'PENDING_PAYMENT', 'CONFIRMED', 'CANCELLED', 'COMPLETED')。
    *   `payment_id` (VARCHAR, 可空): 关联的支付系统订单号。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   **索引**: 建议在 `(user_id, booking_time)` 和 `(showtime_id)` 上创建索引。

## 10. `BookedSeat` 表 (已预订座位表)

*   **含义**: 记录一个订单具体预订了某个场次的哪些座位。这是实现座位锁定和防止超额预订的核心。
*   **对应领域实体**: `internal/domain/booking/booked_seat.go` 中的 `BookedSeat` 实体 (通常作为 `Booking` 聚合的一部分)。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 记录唯一标识符。
    *   `booking_id` (BIGINT, 外键 -> Booking.id): 所属预订订单的 ID。
    *   `showtime_id` (BIGINT, 外键 -> Showtime.id): 关联的场次 ID (冗余字段，方便查询，也可通过 `booking_id` 关联)。
    *   `seat_id` (BIGINT, 外键 -> Seat.id): 预订的具体物理座位 ID。
    *   `price` (DECIMAL): 该座位在此订单中的实际价格 (可能因促销或座位类型而异)。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   **约束**: 对于 `(showtime_id, seat_id)` 组合应有唯一约束，确保一个座位在一个场次中只被有效预订一次 (需考虑订单状态，例如已取消的预订应释放座位)。

## 表关系总结 (ER 图概览)

*   `User (1) -- (0..N) Booking`
*   `Role (1) -- (0..N) User`
*   `Movie (1) -- (0..N) MovieGenre (N) -- (1) Genre` (Movie 和 Genre 是多对多)
*   `Movie (1) -- (0..N) Showtime`
*   `CinemaHall (1) -- (0..N) Showtime`
*   `CinemaHall (1) -- (1..N) Seat`
*   `Showtime (1) -- (0..N) Booking`
*   `Booking (1) -- (1..N) BookedSeat`
*   `Seat (1) -- (0..N) BookedSeat` (一个物理座位可被多次预订，但针对不同场次)

**注意**:

*   所有表都应包含 `created_at` 和 `updated_at` 时间戳字段，用于审计和追踪。
*   主键类型可以根据需求选择自增整数 (BIGINT) 或全局唯一标识符 (如 UUID，VARCHAR(36))，后者在分布式环境中更有优势。
*   外键约束应明确定义，以保证数据完整性。
*   索引的创建需根据实际查询模式进行优化。

此数据模型为 MRS 系统的核心业务提供了结构化存储方案，并考虑了未来的扩展性。