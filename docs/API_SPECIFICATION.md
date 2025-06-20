# 电影预订系统 API 文档 (V1)

## 通用约定

*   **基础路径**: `/api/v1` (为未来的 API 版本保留)
*   **认证**:
    *   多数端点在登录后需要在请求头中包含 `Authorization: Bearer <JWT_TOKEN>`。
    *   管理员特定端点将有 `/admin` 前缀或通过基于角色的访问控制 (RBAC) 中间件进行保护。
*   **请求/响应格式**: JSON
*   **错误响应**: 使用标准的 HTTP 状态码和一致的 JSON 错误对象 (例如: `{"error": {"code": "UNIQUE_VIOLATION", "message": "资源已存在。"}}`)。
*   **分页**: 对于列表端点，使用查询参数如 `page` (例如: `1`) 和 `pageSize` (例如: `20`)。响应应包含分页信息 (总条目数、总页数)。

---

## 0. 健康检查

*   **`GET /health`**
    *   **描述**: 检查服务健康状态
    *   **响应体**: `健康状态响应`
    *   **调用服务**: `HealthHandler.CheckHealth()`

## 1. AuthService (认证服务)

*   **`POST /api/v1/auth/login`**
    *   **描述**: 用户登录
    *   **请求体**: `登录请求` (例如: `{ "email": "user@example.com", "password": "password123" }`)
    *   **响应体**: `登录响应` (例如: `{ "accessToken": "...", "user": { ...用户详情... } }`)
    *   **调用服务**: `AuthHandler.Login()`

## 2. UserService (用户账户与角色服务)

### 公开端点:

*   **`POST /api/v1/users/register`**
    *   **描述**: 用户注册
    *   **请求体**: `注册用户请求` (例如: `{ "name": "测试用户", "email": "test@example.com", "password": "securePassword" }`)
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserHandler.Register()`

### 需要认证的用户端点:

*   **`GET /api/v1/users/me`**
    *   **描述**: 获取当前认证用户的个人资料
    *   **需要认证**: 是
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserHandler.GetUserProfile()`

*   **`PUT /api/v1/users/me`**
    *   **描述**: 更新当前认证用户的个人资料
    *   **需要认证**: 是
    *   **请求体**: `更新用户请求`
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserHandler.UpdateUserProfile()`

### 管理员端点:

*   **`GET /api/v1/admin/users`**
    *   **描述**: 列出所有用户 (分页)
    *   **响应体**: `分页响应包装器<用户资料响应>`
    *   **调用服务**: `UserHandler.ListUsers()`

*   **`GET /api/v1/admin/users/{id}`**
    *   **描述**: 获取特定用户的详细信息
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserHandler.GetUser()`

*   **`PUT /api/v1/admin/users/{id}`**
    *   **描述**: 管理员更新用户资料
    *   **请求体**: `管理员更新用户请求`
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserHandler.UpdateUser()`

*   **`DELETE /api/v1/admin/users/{id}`**
    *   **描述**: 管理员删除用户
    *   **响应**: `204 No Content`
    *   **调用服务**: `UserHandler.DeleteUser()`

*   **`POST /api/v1/admin/users/roles`**
    *   **描述**: 为用户分配角色
    *   **请求体**: `AssignRoleToUserRequest`
    *   **响应**: `成功响应`
    *   **调用服务**: `UserHandler.AssignRoleToUser()`

### 角色管理端点:

*   **`GET /api/v1/admin/roles`**
    *   **描述**: 列出所有角色
    *   **响应体**: `角色响应列表`
    *   **调用服务**: `UserHandler.ListRoles()`

*   **`POST /api/v1/admin/roles`**
    *   **描述**: 创建一个新角色
    *   **请求体**: `{ "name": "editor", "description": "可以编辑内容" }`
    *   **响应体**: `角色响应`
    *   **调用服务**: `UserHandler.CreateRole()`

*   **`PUT /api/v1/admin/roles/{id}`**
    *   **描述**: 更新一个角色
    *   **请求体**: `{ "name": "content_editor", "description": "可以编辑和发布内容" }`
    *   **响应体**: `角色响应`
    *   **调用服务**: `UserHandler.UpdateRole()`

*   **`DELETE /api/v1/admin/roles/{id}`**
    *   **描述**: 删除一个角色
    *   **响应**: `204 No Content`
    *   **调用服务**: `UserHandler.DeleteRole()`

## 3. MovieService (电影与类型服务)

### 需要认证的用户端点:

*   **`GET /api/v1/movies`**
    *   **描述**: 列出电影 (分页)
    *   **查询参数**: `page`, `pageSize`, `genre_name` (类型名称), `release_year` (发行年份)
    *   **响应体**: `分页响应包装器<电影响应>`
    *   **调用服务**: `MovieHandler.ListMovies()`

*   **`GET /api/v1/movies/{id}`**
    *   **描述**: 获取电影详情
    *   **响应体**: `电影响应`
    *   **调用服务**: `MovieHandler.GetMovie()`

*   **`GET /api/v1/genres`**
    *   **描述**: 列出所有电影类型
    *   **响应体**: `类型响应列表`
    *   **调用服务**: `MovieHandler.ListAllGenres()`

### 管理员端点:

*   **`POST /api/v1/admin/movies`**
    *   **描述**: 创建一部新电影
    *   **请求体**: `创建电影请求`
    *   **响应体**: `电影响应`
    *   **调用服务**: `MovieHandler.CreateMovie()`

*   **`PUT /api/v1/admin/movies/{id}`**
    *   **描述**: 更新一部电影
    *   **请求体**: `更新电影请求`
    *   **响应体**: `电影响应`
    *   **调用服务**: `MovieHandler.UpdateMovie()`

*   **`DELETE /api/v1/admin/movies/{id}`**
    *   **描述**: 删除一部电影
    *   **响应**: `204 No Content`
    *   **调用服务**: `MovieHandler.DeleteMovie()`

*   **`POST /api/v1/admin/genres`**
    *   **描述**: 创建一个新的电影类型
    *   **请求体**: `创建类型请求`
    *   **响应体**: `类型响应`
    *   **调用服务**: `MovieHandler.CreateGenre()`

*   **`PUT /api/v1/admin/genres/{id}`**
    *   **描述**: 更新一个电影类型
    *   **请求体**: `更新类型请求`
    *   **响应体**: `类型响应`
    *   **调用服务**: `MovieHandler.UpdateGenre()`

*   **`DELETE /api/v1/admin/genres/{id}`**
    *   **描述**: 删除一个电影类型
    *   **响应**: `204 No Content`
    *   **调用服务**: `MovieHandler.DeleteGenre()`

## 4. CinemaService (影厅与座位布局服务)

### 需要认证的用户端点:

*   **`GET /api/v1/cinema-halls`**
    *   **描述**: 列出可用的影厅
    *   **查询参数**: `page`, `pageSize`
    *   **响应体**: `分页响应包装器<影厅响应>`
    *   **调用服务**: `CinemaHandler.ListAllCinemaHalls()`

*   **`GET /api/v1/cinema-halls/{id}`**
    *   **描述**: 获取特定影厅的详情
    *   **响应体**: `影厅响应`
    *   **调用服务**: `CinemaHandler.GetCinemaHall()`

### 管理员端点:

*   **`POST /api/v1/admin/cinema-halls`**
    *   **描述**: 创建一个新的影厅
    *   **请求体**: `创建影厅请求`
    *   **响应体**: `影厅响应`
    *   **调用服务**: `CinemaHandler.CreateCinemaHall()`

*   **`PUT /api/v1/admin/cinema-halls/{id}`**
    *   **描述**: 更新影厅详情
    *   **请求体**: `更新影厅请求`
    *   **响应体**: `影厅响应`
    *   **调用服务**: `CinemaHandler.UpdateCinemaHall()`

*   **`DELETE /api/v1/admin/cinema-halls/{id}`**
    *   **描述**: 删除一个影厅
    *   **响应**: `204 No Content`
    *   **调用服务**: `CinemaHandler.DeleteCinemaHall()`

## 5. ShowtimeService (放映服务)

### 需要认证的用户端点:

*   **`GET /api/v1/showtimes`**
    *   **描述**: 列出放映场次 (分页)
    *   **查询参数**: `page`, `pageSize`, `movieId`, `hallId`, `date`, `startTimeAfter`
    *   **响应体**: `分页响应包装器<场次响应>`
    *   **调用服务**: `ShowtimeHandler.ListShowtimes()`

*   **`GET /api/v1/showtimes/{id}`**
    *   **描述**: 获取特定放映场次的详情
    *   **响应体**: `场次响应`
    *   **调用服务**: `ShowtimeHandler.GetShowtime()`

*   **`GET /api/v1/showtimes/{id}/seatmap`**
    *   **描述**: 获取特定放映场次的座位图
    *   **响应体**: `座位图响应`
    *   **调用服务**: `ShowtimeHandler.GetSeatMap()`

### 管理员端点:

*   **`POST /api/v1/admin/showtimes`**
    *   **描述**: 安排一个新的放映场次
    *   **请求体**: `创建场次请求`
    *   **响应体**: `场次响应`
    *   **调用服务**: `ShowtimeHandler.CreateShowtime()`

*   **`PUT /api/v1/admin/showtimes/{id}`**
    *   **描述**: 更新一个放映场次
    *   **请求体**: `更新场次请求`
    *   **响应体**: `场次响应`
    *   **调用服务**: `ShowtimeHandler.UpdateShowtime()`

*   **`DELETE /api/v1/admin/showtimes/{id}`**
    *   **描述**: 删除一个放映场次
    *   **响应**: `204 No Content`
    *   **调用服务**: `ShowtimeHandler.DeleteShowtime()`

## 6. BookingService (预订服务)

### 需要认证的用户端点:

*   **`POST /api/v1/bookings`**
    *   **描述**: 创建一个新的预订
    *   **请求体**: `创建预订请求`
    *   **响应体**: `预订确认响应`
    *   **调用服务**: `BookingHandler.CreateBooking()`

*   **`GET /api/v1/bookings`**
    *   **描述**: 列出当前用户的预订 (分页)
    *   **查询参数**: `page`, `pageSize`, `status`
    *   **响应体**: `分页响应包装器<预订详情响应>`
    *   **调用服务**: `BookingHandler.ListBookings()`

*   **`GET /api/v1/bookings/{id}`**
    *   **描述**: 获取当前用户特定预订的详情
    *   **响应体**: `预订详情响应`
    *   **调用服务**: `BookingHandler.GetBooking()`

*   **`POST /api/v1/bookings/{id}/cancel`**
    *   **描述**: 取消一个预订
    *   **响应体**: `预订详情响应`
    *   **调用服务**: `BookingHandler.CancelBooking()`

*   **`POST /api/v1/bookings/{id}/confirm`**
    *   **描述**: 确认预订
    *   **响应体**: `预订详情响应`
    *   **调用服务**: `BookingHandler.ConfirmBooking()`

## 7. ReportService (报告服务)

### 管理员端点:

*   **`GET /api/v1/admin/reports/sales`**
    *   **描述**: 获取销售报告
    *   **查询参数**: `dateFrom`, `dateTo`, `movieId`, `hallId`
    *   **响应体**: `销售报告响应`
    *   **调用服务**: `ReportHandler.GenerateSalesReport()`
