package repository

import (
	"context"
	"mrs/internal/domain/booking"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/movie"
	"mrs/internal/domain/shared"
	"mrs/internal/domain/showtime"
	"mrs/internal/domain/user"
	applog "mrs/pkg/log"

	"gorm.io/gorm"
)

type gormRepositoryProvider struct {
	tx     *gorm.DB
	logger applog.Logger
}

func newGormRepositoryProvider(tx *gorm.DB, logger applog.Logger) shared.RepositoryProvider {
	return &gormRepositoryProvider{tx: tx, logger: logger}
}

func (p *gormRepositoryProvider) GetUserRepository() user.UserRepository {
	return NewGormUserRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetRoleRepository() user.RoleRepository {
	return NewGormRoleRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetMovieRepository() movie.MovieRepository {
	return NewGormMovieRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetGenreRepository() movie.GenreRepository {
	return NewGormGenreRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetShowtimeRepository() showtime.ShowtimeRepository {
	return NewGormShowtimeRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetCinemaHallRepository() cinema.CinemaHallRepository {
	return NewGormCinemaHallRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetSeatRepository() cinema.SeatRepository {
	return NewGormSeatRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetBookingRepository() booking.BookingRepository {
	return NewGormBookingRepository(p.tx, p.logger)
}

func (p *gormRepositoryProvider) GetBookedSeatRepository() booking.BookedSeatRepository {
	return NewGormBookedSeatRepository(p.tx, p.logger)
}

// gormUnitOfWork 实现了 shared.UnitOfWork 接口。
type gormUnitOfWork struct {
	tx     *gorm.DB // 全局的gorm.DB实例，用于开启事务
	logger applog.Logger
}

func NewGormUnitOfWork(tx *gorm.DB, logger applog.Logger) shared.UnitOfWork {
	return &gormUnitOfWork{
		tx:     tx,
		logger: logger.With(applog.String("Component", "gormUnitOfWork")),
	}
}

func (uow *gormUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context, provider shared.RepositoryProvider) error) error {
	logger := uow.logger.With(applog.String("Operation", "Execute"))
	tx := uow.tx.WithContext(ctx).Begin()
	if tx.Error != nil {
		logger.Error("failed to begin transaction", applog.Error(tx.Error))
		return shared.NewTransactionError("Begin", shared.ErrTransactionBeginFailed)
	}

	logger.Debug("transaction started")
	provider := newGormRepositoryProvider(tx, logger)

	// defer 函数确保事务最终会被回滚或提交
	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback().Error; err != nil {
				logger.Error("transaction failed to rollback after panic",
					applog.Any("panic", r),
					applog.Error(err))
				panic(shared.NewTransactionError("Rollback", shared.ErrTransactionRollbackFailed).WithRollbackError(err))
			} else {
				logger.Error("transaction rolled back due to panic", applog.Any("panic", r))
			}
			panic(r)
		}
	}()

	// 执行传入的函数，如果函数返回错误，则回滚事务
	if err := fn(ctx, provider); err != nil {
		// 业务逻辑错误，回滚事务
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			logger.Error("failed to rollback transaction",
				applog.Error(rollbackErr),
				applog.Error(err))
			return shared.NewTransactionError("BusinessLogic", err).WithRollbackError(rollbackErr)
		}
		logger.Error("transaction rolled back due to business logic error", applog.Error(err))
		return err
	}

	// 业务逻辑成功，提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("failed to commit transaction", applog.Error(err))
		// 事务提交失败，尝试回滚
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			logger.Error("failed to rollback transaction after commit error",
				applog.Error(rollbackErr),
				applog.Error(err))
			return shared.NewTransactionError("Commit", shared.ErrTransactionCommitFailed).WithRollbackError(rollbackErr)
		}
		logger.Error("transaction rolled back due to commit error", applog.Error(err))
		return shared.NewTransactionError("Commit", shared.ErrTransactionCommitFailed)
	}

	logger.Debug("transaction committed successfully")
	return nil
}
