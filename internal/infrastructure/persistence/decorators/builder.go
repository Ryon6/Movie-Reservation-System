package decorators

import (
	"mrs/internal/domain/movie"
	"mrs/internal/domain/showtime"
	"mrs/internal/infrastructure/persistence/decorators/repository_circuitbreaker"
	"mrs/internal/infrastructure/persistence/mysql/repository"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

func NewMovieRepository(db *gorm.DB, logger applog.Logger) movie.MovieRepository {
	repo := repository.NewGormMovieRepository(db, logger)
	return repository_circuitbreaker.NewMovieRepositoryWithCircuitBreaker(repo, logger)
}

func NewShowtimeRepository(db *gorm.DB, logger applog.Logger) showtime.ShowtimeRepository {
	repo := repository.NewGormShowtimeRepository(db, logger)
	return repository_circuitbreaker.NewShowtimeRepositoryWithCircuitBreaker(repo, logger)
}
