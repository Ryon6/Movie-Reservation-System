// internal/api/router.go
package routers

import (
	"mrs/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并返回 Gin 引擎。
// healthHandler 应作为参数传入，或者在此函数内部创建。
func SetupRouter(
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	movieHandler *handlers.MovieHandler,
	cinemaHandler *handlers.CinemaHandler,
	showtimeHandler *handlers.ShowtimeHandler,
	bookingHandler *handlers.BookingHandler,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
	// ... 其他处理器 ...
) *gin.Engine {
	router := gin.Default() // 或者 gin.New() 并添加必要的中间件

	// 健康检查路由
	router.GET("/health", healthHandler.CheckHealth)

	apiV1 := router.Group("/api/v1")

	// 管理员路由，必须通过认证和权限验证
	adminRoutes := apiV1.Group("/admin")
	adminRoutes.Use(authMiddleware)
	adminRoutes.Use(adminMiddleware)

	// 认证路由
	authRoutes := apiV1.Group("/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
	}

	// 用户管理路由
	userRoutes := apiV1.Group("/users")
	{
		// 无需认证的用户路由
		userRoutes.POST("/register", userHandler.Register) // 用户注册

		// 需要认证的用户路由
		authUserRoutes := userRoutes.Group("")
		authUserRoutes.Use(authMiddleware)
		{
			authUserRoutes.GET("/me", userHandler.GetUserProfile) // 获取个人信息
		}
	}

	// 电影管理路由
	movieRoutes := apiV1.Group("/movies")
	movieRoutes.Use(authMiddleware)
	{
		movieRoutes.GET("", movieHandler.ListMovies)
		movieRoutes.GET("/:id", movieHandler.GetMovie) // 获取单个电影
	}
	movieAdminRoutes := adminRoutes.Group("/movies")
	{
		movieAdminRoutes.POST("", movieHandler.CreateMovie)
		movieAdminRoutes.PUT("/:id", movieHandler.UpdateMovie)
		movieAdminRoutes.DELETE("/:id", movieHandler.DeleteMovie)
	}

	genreRoutes := apiV1.Group("/genres")
	genreRoutes.Use(authMiddleware)
	{
		genreRoutes.GET("", movieHandler.ListAllGenres)
	}
	genreAdminRoutes := adminRoutes.Group("/genres")
	{
		genreAdminRoutes.POST("", movieHandler.CreateGenre)
		genreAdminRoutes.PUT("/:id", movieHandler.UpdateGenre)
		genreAdminRoutes.DELETE("/:id", movieHandler.DeleteGenre)
	}

	// 影厅管理路由
	cinemaHallRoutes := apiV1.Group("/cinema-halls")
	cinemaHallRoutes.Use(authMiddleware)
	{
		cinemaHallRoutes.GET("", cinemaHandler.ListAllCinemaHalls)
		cinemaHallRoutes.GET("/:id", cinemaHandler.GetCinemaHall)
	}
	cinemaHallAdminRoutes := adminRoutes.Group("/cinema-halls")
	{
		cinemaHallAdminRoutes.POST("", cinemaHandler.CreateCinemaHall)
		cinemaHallAdminRoutes.PUT("/:id", cinemaHandler.UpdateCinemaHall)
		cinemaHallAdminRoutes.DELETE("/:id", cinemaHandler.DeleteCinemaHall)
	}

	// 放映场次管理路由
	showtimeRoutes := apiV1.Group("/showtimes")
	showtimeRoutes.Use(authMiddleware)
	{
		showtimeRoutes.GET("", showtimeHandler.ListShowtimes)
		showtimeRoutes.GET("/:id", showtimeHandler.GetShowtime)
		showtimeRoutes.GET("/:id/seatmap", showtimeHandler.GetSeatMap)
	}
	showtimeAdminRoutes := adminRoutes.Group("/showtimes")
	{
		showtimeAdminRoutes.POST("", showtimeHandler.CreateShowtime)
		showtimeAdminRoutes.PUT("/:id", showtimeHandler.UpdateShowtime)
		showtimeAdminRoutes.DELETE("/:id", showtimeHandler.DeleteShowtime)
	}

	// 订单管理路由
	bookingRoutes := apiV1.Group("/bookings")
	bookingRoutes.Use(authMiddleware)
	{
		bookingRoutes.POST("", bookingHandler.CreateBooking)
		bookingRoutes.GET("", bookingHandler.ListBookings)
		bookingRoutes.GET("/:id", bookingHandler.GetBooking)
		bookingRoutes.POST("/:id/cancel", bookingHandler.CancelBooking)
		bookingRoutes.POST("/:id/confirm", bookingHandler.ConfirmBooking)
	}
	return router
}
