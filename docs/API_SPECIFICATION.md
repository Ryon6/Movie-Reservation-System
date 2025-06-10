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

## 1. AuthService (认证服务)

*   **`POST /api/v1/auth/login`**
    *   **描述**: 用户登录。
    *   **请求体**: `登录请求` (例如: `{ "email": "user@example.com", "password": "password123" }`)
    *   **响应体**: `登录响应` (例如: `{ "accessToken": "...", "refreshToken": "...", "expiresIn": 3600, "user": { ...用户详情... } }`)
    *   **调用服务**: `AuthService.Login()`

*   **`POST /api/v1/auth/refresh-token`**
    *   **描述**: 使用刷新令牌获取新的访问令牌。
    *   **请求体**: `{ "refreshToken": "..." }`
    *   **响应体**: `{ "accessToken": "...", "expiresIn": 3600 }`
    *   **调用服务**: `AuthService.RefreshToken()`

*   **`POST /api/v1/auth/logout`** (可选, 取决于令牌失效策略)
    *   **描述**: 用户登出 (例如: 如果使用黑名单，则在服务器端使刷新令牌失效)。
    *   **需要认证**: 是
    *   **调用服务**: `AuthService.Logout()`

---

## 2. UserService (用户账户与角色服务)

*   **`POST /api/v1/users/register`**
    *   **描述**: 用户注册。
    *   **请求体**: `注册用户请求` (例如: `{ "name": "测试用户", "email": "test@example.com", "password": "securePassword" }`)
    *   **响应体**: `用户资料响应` (用户详情，不包括密码等敏感信息)
    *   **调用服务**: `UserService.Register(RegisterUserRequest)`

*   **`GET /api/v1/users/me`**
    *   **描述**: 获取当前认证用户的个人资料。
    *   **需要认证**: 是
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserService.GetUser(GetUserRequest)`

*   **`PUT /api/v1/users/me`**
    *   **描述**: 更新当前认证用户的个人资料。
    *   **需要认证**: 是
    *   **请求体**: `更新用户请求` (例如: `{ "name": "更新后的名称" }`)
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserService.UpdateUser(UpdateUserRequest)`

---

## 管理员端点 (前缀为 `/admin` 并需要管理员权限)

### 2.1. UserService (管理员 - 用户管理)

*   **`GET /api/v1/admin/users`**
    *   **描述**: 列出所有用户 (分页)。
    *   **响应体**: `分页响应包装器<用户资料响应>`
    *   **调用服务**: `UserService.ListUsers(ListUserRequest)`
*   **`GET /api/v1/admin/users/{userId}`**
    *   **描述**: 获取特定用户的详细信息。
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserService.GetUser(GetUserRequest)`
*   **`PUT /api/v1/admin/users/{userId}`**
    *   **描述**: 管理员更新用户资料 (例如: 修改用户权限)。
    *   **请求体**: `管理员更新用户请求`
    *   **响应体**: `用户资料响应`
    *   **调用服务**: `UserService.UpdateUser(UpdateUserRequest)`
*   **`DELETE /api/v1/admin/users/{userId}`**
    *   **描述**: 管理员删除用户。
    *   **响应**: `204 No Content` 或 `成功响应`
    *   **调用服务**: `UserService.DeleteUser(DeleteUserRequest)`

### 2.2. UserService (管理员 - 角色管理)

*   **`POST /api/v1/admin/roles`**
    *   **描述**: 创建一个新角色。
    *   **请求体**: `{ "name": "editor", "description": "可以编辑内容" }`
    *   **响应体**: `角色响应`
    *   **调用服务**: `UserService.CreateRole(CreateRoleRequest)`
*   **`GET /api/v1/admin/roles`**
    *   **描述**: 列出所有角色。
    *   **响应体**: `角色响应列表`
    *   **调用服务**: `UserService.ListRoles()`
*   **`PUT /api/v1/admin/roles/{roleId}`**
    *   **描述**: 更新一个角色。
    *   **请求体**: `{ "name": "content_editor", "description": "可以编辑和发布内容" }`
    *   **响应体**: `角色响应`
    *   **调用服务**: `UserService.UpdateRole(UpdateRoleRequest)`
*   **`DELETE /api/v1/admin/roles/{roleId}`**
    *   **描述**: 删除一个角色。
    *   **响应**: `204 No Content`
    *   **调用服务**: `UserService.DeleteRole(DeleteRoleRequest)`
*   **`POST /api/v1/admin/users/roles`**
    *   **描述**: 为用户分配角色。
    *   **请求体**: `AssignRoleToUserRequest`
    *   **响应**: `成功响应`
    *   **调用服务**: `UserService.AssignRoleToUser(AssignRoleToUserRequest)`

---

## 3. MovieService (电影与类型服务)(Completed)

### 公开端点:

*   **`GET /api/v1/movies`**
    *   **描述**: 列出电影 (分页)。
    *   **查询参数**: `page`, `page_size`, `genre_name` (类型名称), `release_year`发行年份
    *   **响应体**: `分页响应包装器<电影响应>`
    *   **调用服务**: `MovieService.ListMovies(ListMovieRequest)`
*   **`GET /api/v1/movies/{movieId}`**
    *   **描述**: 获取电影详情。
    *   **响应体**: `电影响应` (包含类型信息)
    *   **调用服务**: `MovieService.GetMovie(GetMovieRequest)`
*   **`GET /api/v1/genres`**
    *   **描述**: 列出所有电影类型。
    *   **响应体**: `类型响应列表`
    *   **调用服务**: `MovieService.ListAllGenres()`

### 管理员端点:

*   **`POST /api/v1/admin/movies`**
    *   **描述**: 创建一部新电影。
    *   **请求体**: `创建电影请求` (标题, 描述, 上映日期, 时长, 类型名称列表等)
    *   **响应体**: `电影响应`
    *   **调用服务**: `MovieService.CreateMovie(CreateMovieRequest)`
*   **`PUT /api/v1/admin/movies/{movieId}`**
    *   **描述**: 更新一部电影。
    *   **请求体**: `更新电影请求`
    *   **响应体**: `成功消息`
    *   **调用服务**: `MovieService.UpdateMovie(UpdateMovieRequest)`
*   **`DELETE /api/v1/admin/movies/{movieId}`**
    *   **描述**: 删除一部电影。
    *   **响应**: `204 No Content`
    *   **调用服务**: `MovieService.DeleteMovie(DeleteMovieRequest)`
*   **`POST /api/v1/admin/genres`**
    *   **描述**: 创建一个新的电影类型。
    *   **请求体**: `创建类型请求` (名称, 描述)
    *   **响应体**: `类型响应`
    *   **调用服务**: `MovieService.CreateGenre(CreateGenreRequest)`
*   **`PUT /api/v1/admin/genres/{genreId}`**
    *   **描述**: 更新一个电影类型。
    *   **请求体**: `更新类型请求`
    *   **响应体**: `类型响应`
    *   **调用服务**: `MovieService.UpdateGenre(UpdateGenreRequest)`
*   **`DELETE /api/v1/admin/genres/{genreId}`**
    *   **描述**: 删除一个电影类型。
    *   **响应**: `204 No Content`
    *   **调用服务**: `MovieService.DeleteGenre(DeleteGenreRequest)`

---

## 4. CinemaService (影厅与座位布局服务)

### 公开/用户端点 (信息查询):

*   **`GET /api/v1/cinema-halls`**
    *   **描述**: 列出可用的影厅。
    *   **查询参数**: `page`, `pageSize`
    *   **响应体**: `分页响应包装器<影厅响应>` (基本信息)
    *   **调用服务**: `CinemaService.ListAllCinemaHalls()` 

*   **`GET /api/v1/cinema-halls/{hallId}`**
    *   **描述**: 获取特定影厅的详情。
    *   **响应体**: `影厅响应` (详细信息)
    *   **调用服务**: `CinemaService.GetCinemaHall(GetCinemaHallRequest)`

### 管理员端点:

*   **`POST /api/v1/admin/cinema-halls`**
    *   **描述**: 创建一个新的影厅。
    *   **请求体**: `创建影厅请求` (名称, 屏幕类型, 座位布局详情)
    *   **响应体**: `影厅响应` (包含座位布局的完整详情)
    *   **调用服务**: `CinemaService.CreateCinemaHall(CreateCinemaHallRequest)`
*   **`PUT /api/v1/admin/cinema-halls/{hallId}`**
    *   **描述**: 更新影厅详情 (名称, 屏幕类型)。
    *   **请求体**: `更新影厅请求`
    *   **响应体**: `影厅响应`
    *   **调用服务**: `CinemaService.UpdateCinemaHall(UpdateCinemaHallRequest)`
*   **`DELETE /api/v1/admin/cinema-halls/{hallId}`**
    *   **描述**: 删除一个影厅。
    *   **响应**: `204 No Content`
    *   **调用服务**: `CinemaService.DeleteCinemaHall(DeleteCinemaHallRequest)`

---

## 5. ShowtimeService (放映服务)

### 公开/用户端点:

*   **`GET /api/v1/showtimes`**
    *   **描述**: 列出放映场次 (分页)。
    *   **查询参数**: `page`, `pageSize`, `movieId` (电影ID), `hallId` (影厅ID), `date` (日期, 例如: YYYY-MM-DD), `startTimeAfter` (此时间之后开始)
    *   **响应体**: `分页响应包装器<场次响应>`
    *   **调用服务**: `ShowtimeService.ListShowtimes(params)`

*   **`GET /api/v1/showtimes/{showtimeId}`**
    *   **描述**: 获取特定放映场次的详情。
    *   **响应体**: `场次响应` (包含电影和影厅信息)
    *   **调用服务**: `ShowtimeService.GetShowtimeByID(showtimeId)`

*   **`GET /api/v1/showtimes/{showtimeId}/seat-map`**
    *   **描述**: 获取特定放映场次的座位图 (布局和可用状态)。
    *   **响应体**: `座位图响应` (例如: ` { "hallName": "1号厅", "seats": [ [ { "id": "A1", "status": "available" }, ... ] ] } `)
    *   **调用服务**: `ShowtimeService.GetShowtimeSeatMap(showtimeId)`

### 管理员端点:

*   **`POST /api/v1/admin/showtimes`**
    *   **描述**: 安排一个新的放映场次。
    *   **请求体**: `安排场次请求` (电影ID, 影厅ID, 开始时间, 票价)
    *   **响应体**: `场次响应`
    *   **调用服务**: `ShowtimeService.ScheduleShowtime(params)`
*   **`PUT /api/v1/admin/showtimes/{showtimeId}`**
    *   **描述**: 更新一个放映场次 (例如: 修改时间, 票价)。
    *   **请求体**: `更新场次请求`
    *   **响应体**: `场次响应`
    *   **调用服务**: `ShowtimeService.UpdateShowtime(showtimeId, params)`
*   **`DELETE /api/v1/admin/showtimes/{showtimeId}`**
    *   **描述**: 取消/删除一个放映场次。
    *   **响应**: `204 No Content`
    *   **调用服务**: `ShowtimeService.CancelShowtime(showtimeId)`

---

## 6. BookingService (预订服务)

### 需要认证 (用户):

*   **`POST /api/v1/bookings`**
    *   **描述**: 创建一个新的预订 (预留座位)。
    *   **请求体**: `创建预订请求` (例如: `{ "showtimeId": "...", "selectedSeatIds": ["A1", "A2"] }`)
    *   **响应体**: `预订确认响应` (预订ID, 总价, 状态: pending_payment/confirmed 待支付/已确认)
    *   **调用服务**: `BookingService.CreateBooking(userId, params)` (可能涉及临时锁座)

*   **`GET /api/v1/bookings`**
    *   **描述**: 列出当前用户的预订 (分页)。
    *   **查询参数**: `page`, `pageSize`, `status` (状态)
    *   **响应体**: `分页响应包装器<预订详情响应>`
    *   **调用服务**: `BookingService.ListUserBookings(userId, params)`

*   **`GET /api/v1/bookings/{bookingId}`**
    *   **描述**: 获取当前用户特定预订的详情。
    *   **响应体**: `预订详情响应`
    *   **调用服务**: `BookingService.GetUserBookingDetails(userId, bookingId)`

*   **`POST /api/v1/bookings/{bookingId}/cancel`**
    *   **描述**: 请求取消一个预订 (受取消政策约束)。
    *   **响应**: `预订详情响应` (包含更新后的状态) 或 `成功响应`
    *   **调用服务**: `BookingService.CancelBooking(userId, bookingId)`

*   **`POST /api/v1/bookings/{bookingId}/confirm-payment`** (如果支付是独立步骤)
    *   **描述**: 确认预订的支付。
    *   **请求体**: `{ "paymentReference": "..." }` (支付参考号)
    *   **响应体**: `预订详情响应` (状态: confirmed 已确认)
    *   **调用服务**: `BookingService.ConfirmPayment(bookingId, paymentDetails)`

### 管理员端点:

*   **`GET /api/v1/admin/bookings`**
    *   **描述**: 列出所有用户的全部预订 (分页)。
    *   **查询参数**: `page`, `pageSize`, `userId` (用户ID), `showtimeId` (场次ID), `status` (状态)
    *   **响应体**: `分页响应包装器<预订详情响应>`
    *   **调用服务**: `BookingService.ListAllBookings(params)`

*   **`GET /api/v1/admin/bookings/{bookingId}`**
    *   **描述**: 获取任意预订的详情。
    *   **响应体**: `预订详情响应`
    *   **调用服务**: `BookingService.GetBookingDetailsForAdmin(bookingId)`

---

## 7. ReportService (报告服务)

### 管理员端点:

*   **`GET /api/v1/admin/reports/sales`**
    *   **描述**: 获取销售报告。
    *   **查询参数**: `dateFrom` (起始日期), `dateTo` (结束日期), `movieId` (电影ID), `hallId` (影厅ID)
    *   **响应体**: `销售报告响应`
    *   **调用服务**: `ReportService.GenerateSalesReport(params)`

*   **`GET /api/v1/admin/reports/occupancy`**
    *   **描述**: 获取上座率报告 (针对场次/影厅)。
    *   **查询参数**: `dateFrom` (起始日期), `dateTo` (结束日期), `movieId` (电影ID), `hallId` (影厅ID)
    *   **响应体**: `上座率报告响应`
    *   **调用服务**: `ReportService.GenerateOccupancyReport(params)`
