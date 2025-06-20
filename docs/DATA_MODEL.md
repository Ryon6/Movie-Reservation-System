# 数据模型详解 (MRS 数据库设计)

本文档详细描述了电影预订系统 (MRS) 的核心数据表结构、字段定义以及它们之间的关系。这些表构成了系统持久化存储的基础。

## 1. `User` 表 (用户表)

*   **含义**: 存储所有已注册用户的基本信息。
*   **对应领域实体**: `internal/domain/user/user.go` 中的 `User` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 用户唯一标识符。
    *   `username` (VARCHAR(100), 唯一索引, 非空): 用户名，用于登录。
    *   `password_hash` (VARCHAR(255), 非空): 存储用户密码的哈希值。**严禁存储明文密码。**
    *   `email` (VARCHAR(255), 唯一索引, 非空): 用户电子邮箱，可用于登录、接收通知、密码找回。
    *   `role_id` (BIGINT, 外键 -> Role.id, 非空): 关联到 `Role` 表，表示该用户的角色。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。

## 2. `Role` 表 (角色表)

*   **含义**: 定义系统中的用户角色，用于权限管理。
*   **对应领域实体**: `internal/domain/user/role.go` 中的 `Role` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 角色唯一标识符。
    *   `name` (VARCHAR(50), 唯一索引, 非空): 角色名称 (例如: 'ADMIN', 'USER')。
    *   `description` (VARCHAR(255), 可空): 角色描述。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。

## 3. `Movie` 表 (电影表)

*   **含义**: 存储电影的基本信息。
*   **对应领域实体**: `internal/domain/movie/movie.go` 中的 `Movie` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 电影唯一标识符。
    *   `title` (VARCHAR(255), 唯一索引, 非空): 电影标题。
    *   `release_date` (DATE, 非空): 上映日期。
    *   `description` (TEXT, 可空): 电影剧情简介或描述。
    *   `poster_url` (VARCHAR(500), 可空): 电影海报图片的 URL 地址。
    *   `duration_minutes` (INT): 电影时长，单位为分钟。
    *   `rating` (FLOAT): 评分。
    *   `age_rating` (VARCHAR(50), 可空): 年龄分级。
    *   `cast` (TEXT, 可空): 演员表。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。

## 4. `Genre` 表 (类型表)

*   **含义**: 存储电影的类型标签，如动作、喜剧、科幻等。
*   **对应领域实体**: `internal/domain/movie/genre.go` 中的 `Genre` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 类型唯一标识符。
    *   `name` (VARCHAR(100), 唯一索引, 非空): 类型名称。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。

## 5. `MovieGenre` 表 (电影类型关联表)

*   **含义**: 中间表，用于表示电影 (`Movie`) 和类型 (`Genre`) 之间的多对多关系。
*   **字段**:
    *   `movie_id` (BIGINT, 外键 -> Movie.id): 关联的电影 ID。
    *   `genre_id` (BIGINT, 外键 -> Genre.id): 关联的类型 ID。
    *   **约束**: 
        - `(movie_id, genre_id)` 构成联合主键，确保唯一性。
        - 删除电影时级联删除关联记录 (ON DELETE CASCADE)。
        - 删除类型时限制删除 (ON DELETE RESTRICT)。

## 6. `CinemaHall` 表 (影厅表)

*   **含义**: 存储电影院中各个影厅的基本信息。
*   **对应领域实体**: `internal/domain/cinema/cinema_hall.go` 中的 `CinemaHall` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 影厅唯一标识符。
    *   `name` (VARCHAR(50), 唯一索引, 非空): 影厅名称 (例如: "1号厅", "IMAX厅")。
    *   `screen_type` (VARCHAR(50), 可空): 屏幕类型 (例如: "2D", "3D", "IMAX")。
    *   `sound_system` (VARCHAR(100), 可空): 音响系统。
    *   `row_count` (INT, 非空): 座位行数。
    *   `col_count` (INT, 非空): 座位列数。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。

## 7. `Seat` 表 (座位表)

*   **含义**: 定义影厅内每个物理座位的具体信息。
*   **对应领域实体**: `internal/domain/cinema/seat.go` 中的 `Seat` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 座位唯一标识符。
    *   `cinema_hall_id` (BIGINT, 外键 -> CinemaHall.id, 非空): 该座位所属的影厅 ID。
    *   `row_identifier` (VARCHAR(10), 非空): 座位行标识 (例如: "A", "B", 或数字 "1", "2")。
    *   `seat_number` (VARCHAR(10), 非空): 座位在本行中的编号。
    *   `type` (VARCHAR(50), 默认值 'STANDARD'): 座位类型。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。
    *   **索引**: 
        - 在 `cinema_hall_id` 上创建索引以优化查询。
        - `(cinema_hall_id, row_identifier, seat_number)` 构成联合唯一索引。
    *   **约束**: 删除影厅时级联删除座位 (ON DELETE CASCADE)。

## 8. `Showtime` 表 (放映时间表)

*   **含义**: 核心表之一，存储电影的具体放映安排。
*   **对应领域实体**: `internal/domain/showtime/showtime.go` 中的 `Showtime` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 放映场次唯一标识符。
    *   `movie_id` (BIGINT, 外键 -> Movie.id, 非空): 放映的电影 ID。
    *   `cinema_hall_id` (BIGINT, 外键 -> CinemaHall.id, 非空): 放映所在的影厅 ID。
    *   `start_time` (TIMESTAMP, 非空): 放映开始时间。
    *   `end_time` (TIMESTAMP, 非空): 放映结束时间。
    *   `price` (DECIMAL, 非空): 该场次的基准票价。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。
    *   **索引**: 
        - `(movie_id, start_time)` 联合索引。
        - `(cinema_hall_id, start_time)` 联合索引。

## 9. `Booking` 表 (预订订单表)

*   **含义**: 存储用户的预订订单信息。
*   **对应领域实体**: `internal/domain/booking/booking.go` 中的 `Booking` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 预订订单唯一标识符。
    *   `user_id` (BIGINT, 外键 -> User.id, 非空): 下单用户的 ID。
    *   `showtime_id` (BIGINT, 外键 -> Showtime.id, 非空): 预订的场次 ID。
    *   `booking_time` (TIMESTAMP, 非空): 订单创建时间。
    *   `total_amount` (DECIMAL, 非空): 订单总金额。
    *   `status` (VARCHAR, 非空): 订单状态。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。
    *   **索引**: 
        - 在 `user_id` 上创建索引。
        - 在 `showtime_id` 上创建索引。

## 10. `BookedSeat` 表 (已预订座位表)

*   **含义**: 记录一个订单具体预订了某个场次的哪些座位。
*   **对应领域实体**: `internal/domain/booking/booked_seat.go` 中的 `BookedSeat` 实体。
*   **字段**:
    *   `id` (BIGINT, 主键, 自增): 记录唯一标识符。
    *   `booking_id` (BIGINT, 外键 -> Booking.id, 非空): 所属预订订单的 ID。
    *   `showtime_id` (BIGINT, 外键 -> Showtime.id, 非空): 关联的场次 ID。
    *   `seat_id` (BIGINT, 外键 -> Seat.id, 非空): 预订的具体物理座位 ID。
    *   `price` (DECIMAL, 非空): 该座位在此订单中的实际价格。
    *   `created_at` (TIMESTAMP): 记录创建时间。
    *   `updated_at` (TIMESTAMP): 记录最后更新时间。
    *   `deleted_at` (TIMESTAMP, 可空): 软删除时间戳。
    *   **索引**: 
        - 在 `booking_id` 上创建索引。
        - `(showtime_id, seat_id)` 构成联合唯一索引，确保一个座位在一个场次中只能被预订一次。

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

*   所有表都使用 GORM 的 `Model` 嵌入结构，包含 `id`、`created_at`、`updated_at` 和 `deleted_at` 字段。
*   所有表都支持软删除功能（通过 `deleted_at` 字段）。
*   外键约束和索引的设计旨在保证数据完整性的同时优化查询性能。
*   字段长度和类型的选择基于实际业务需求，并考虑了存储效率。
*   所有时间相关字段使用 `TIMESTAMP` 类型，支持时区处理。

此数据模型为 MRS 系统的核心业务提供了结构化存储方案，并考虑了未来的扩展性。