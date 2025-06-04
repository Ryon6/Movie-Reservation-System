package shared

import (
	"context"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/showtime"
	"mrs/internal/domain/user"
)

// RepositoryProvider 提供所有领域对象的仓库接口，用于在事务上下文中获取仓库实例。
// 应用服务通过这个接口获取在当前事务中操作的仓库。
type RepositoryProvider interface {
	GetUserRepository() user.UserRepository
	GetRoleRepository() user.RoleRepository
	GetMovieRepository() movie.MovieRepository
	GetGenreRepository() movie.GenreRepository
	GetShowtimeRepository() showtime.ShowtimeRepository
	GetCinemaHallRepository() cinema.CinemaHallRepository
	GetSeatRepository() cinema.SeatRepository
}

// UnitOfWork 定义了单元工作的接口。
// 应用服务将使用这个接口来执行事务性操作。
type UnitOfWork interface {
	// Execute 在一个事务中执行给定的函数 fn。
	// fn 会接收一个 RepositoryProvider，用于获取在该事务中操作的仓库。
	// 如果函数执行过程中发生错误，事务会自动回滚。
	// 否则，事务会自动提交，并返回nil。
	Execute(ctx context.Context, fn func(ctx context.Context, provider RepositoryProvider) error) error
}
