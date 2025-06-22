package main

import (
	"fmt"
	"log"
	"math/rand"
	"mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/mysql/models"
	"mrs/internal/infrastructure/persistence/mysql/repository"
	applog "mrs/pkg/log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// 基础数据量
	numUsers       = 1000
	numMovies      = 100
	numCinemaHalls = 50
	numShowtimes   = 200
	numBookings    = 2000

	// 批量插入大小
	batchSize = 100
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config", "app.dev", "yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, err := applog.NewZapLogger(cfg.LogConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 初始化数据库连接
	dbFactory := repository.NewMysqlDBFactory(logger)
	db, err := dbFactory.CreateDBConnection(cfg.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 设置随机种子
	gofakeit.Seed(time.Now().UnixNano())

	// 开始填充数据
	if err := seedData(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	log.Println("Data seeding completed successfully!")
}

func seedData(db *gorm.DB) error {
	// 1. 创建用户角色
	roles := createRoles()
	if err := db.Create(&roles).Error; err != nil {
		return fmt.Errorf("failed to create roles: %v", err)
	}

	// 2. 创建用户
	users := createUsers(roles)
	if err := batchInsert(db, users, batchSize); err != nil {
		return fmt.Errorf("failed to create users: %v", err)
	}

	// 3. 创建电影类型
	genres := createGenres()
	if err := db.Create(&genres).Error; err != nil {
		return fmt.Errorf("failed to create genres: %v", err)
	}

	// 4. 创建电影
	movies := createMovies(genres)
	if err := batchInsert(db, movies, batchSize); err != nil {
		return fmt.Errorf("failed to create movies: %v", err)
	}

	// 5. 创建影厅
	halls := createCinemaHalls()
	if err := batchInsert(db, halls, batchSize); err != nil {
		return fmt.Errorf("failed to create cinema halls: %v", err)
	}

	// 6. 创建座位
	seats := createSeats(halls)
	if err := batchInsert(db, seats, batchSize); err != nil {
		return fmt.Errorf("failed to create seats: %v", err)
	}

	// 7. 创建场次
	showtimes := createShowtimes(movies, halls)
	if err := batchInsert(db, showtimes, batchSize); err != nil {
		return fmt.Errorf("failed to create showtimes: %v", err)
	}

	// 8. 创建订单
	bookings := createBookings(users, showtimes, seats)
	if err := batchInsert(db, bookings, batchSize); err != nil {
		return fmt.Errorf("failed to create bookings: %v", err)
	}

	return nil
}

func createRoles() []models.RoleGorm {
	return []models.RoleGorm{
		{Name: "admin"},
		{Name: "user"},
	}
}

func createUsers(roles []models.RoleGorm) []models.UserGorm {
	users := make([]models.UserGorm, 0, numUsers)
	for i := 0; i < numUsers; i++ {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(gofakeit.Password(true, true, true, true, false, 10)), bcrypt.DefaultCost)
		user := models.UserGorm{
			Username:     gofakeit.Username(),
			Email:        gofakeit.Email(),
			PasswordHash: string(hashedPassword),
			RoleID:       roles[rand.Intn(len(roles))].ID,
		}
		users = append(users, user)
	}
	return users
}

func createGenres() []models.GenreGorm {
	genres := []string{"动作", "冒险", "喜剧", "剧情", "恐怖", "科幻", "动画", "纪录片"}
	result := make([]models.GenreGorm, len(genres))
	for i, name := range genres {
		result[i] = models.GenreGorm{Name: name}
	}
	return result
}

func createMovies(genres []models.GenreGorm) []models.MovieGorm {
	movies := make([]models.MovieGorm, 0, numMovies)
	for i := 0; i < numMovies; i++ {
		selectedGenre := &genres[rand.Intn(len(genres))]
		movie := models.MovieGorm{
			Title:           gofakeit.MovieName(),
			Description:     gofakeit.Sentence(10),
			DurationMinutes: gofakeit.Number(60, 180),
			ReleaseDate:     gofakeit.Date(),
			Rating:          float32(gofakeit.Float32Range(1, 5)),
			AgeRating:       "PG-13",
			Cast:            gofakeit.Name() + ", " + gofakeit.Name(),
			Genres:          []*models.GenreGorm{selectedGenre},
		}
		movies = append(movies, movie)
	}
	return movies
}

func createCinemaHalls() []models.CinemaHallGorm {
	halls := make([]models.CinemaHallGorm, 0, numCinemaHalls)
	for i := 0; i < numCinemaHalls; i++ {
		rowCount := gofakeit.Number(5, 15)
		colCount := gofakeit.Number(8, 20)
		hall := models.CinemaHallGorm{
			Name:        fmt.Sprintf("放映厅-%d", i+1),
			ScreenType:  []string{"2D", "3D", "IMAX"}[rand.Intn(3)],
			SoundSystem: []string{"Dolby", "DTS", "SDDS"}[rand.Intn(3)],
			RowCount:    rowCount,
			ColCount:    colCount,
		}
		halls = append(halls, hall)
	}
	return halls
}

func createSeats(halls []models.CinemaHallGorm) []models.SeatGorm {
	var seats []models.SeatGorm
	for _, hall := range halls {
		for row := 1; row <= hall.RowCount; row++ {
			for col := 1; col <= hall.ColCount; col++ {
				seat := models.SeatGorm{
					CinemaHallID:  hall.ID,
					RowIdentifier: fmt.Sprintf("%d", row),
					SeatNumber:    fmt.Sprintf("%d", col),
					Type:          "STANDARD",
				}
				seats = append(seats, seat)
			}
		}
	}
	return seats
}

func createShowtimes(movies []models.MovieGorm, halls []models.CinemaHallGorm) []models.ShowtimeGorm {
	showtimes := make([]models.ShowtimeGorm, 0, numShowtimes)
	startDate := time.Now()

	for i := 0; i < numShowtimes; i++ {
		movie := movies[rand.Intn(len(movies))]
		hall := halls[rand.Intn(len(halls))]

		showtime := models.ShowtimeGorm{
			MovieID:      movie.ID,
			CinemaHallID: hall.ID,
			StartTime:    startDate.Add(time.Duration(i) * time.Hour * 3),
			Price:        float64(gofakeit.Number(3000, 8000)) / 100, // 30-80元
		}
		showtimes = append(showtimes, showtime)
	}
	return showtimes
}

func createBookings(users []models.UserGorm, showtimes []models.ShowtimeGorm, seats []models.SeatGorm) []models.BookingGorm {
	bookings := make([]models.BookingGorm, 0, numBookings)
	now := time.Now()

	for i := 0; i < numBookings; i++ {
		user := users[rand.Intn(len(users))]
		showtime := showtimes[rand.Intn(len(showtimes))]
		seat := seats[rand.Intn(len(seats))]
		bookingTime := now.Add(-time.Duration(rand.Intn(72)) * time.Hour)

		booking := models.BookingGorm{
			UserID:      user.ID,
			ShowtimeID:  showtime.ID,
			BookingTime: bookingTime,
			TotalAmount: showtime.Price,
			Status:      "CONFIRMED",
		}

		// 添加订座信息
		bookedSeat := models.BookedSeatGorm{
			BookingID: booking.ID,
			SeatID:    seat.ID,
		}
		booking.BookedSeats = []models.BookedSeatGorm{bookedSeat}

		bookings = append(bookings, booking)
	}
	return bookings
}

// batchInsert 批量插入数据
func batchInsert(db *gorm.DB, records interface{}, batchSize int) error {
	return db.CreateInBatches(records, batchSize).Error
}
