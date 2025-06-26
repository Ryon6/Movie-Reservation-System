package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mrs/internal/domain/cinema"
	"mrs/internal/domain/user"
	"mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/mysql/models"
	applog "mrs/pkg/log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserCredential 用于存储用户凭证信息
type UserCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

const (
	// 基础数据量
	numUsers       = 2000
	numMovies      = 100
	numCinemaHalls = 50
)

func main() {
	// 确保日志目录存在
	if err := os.MkdirAll("./var/log", 0755); err != nil {
		log.Fatalf("Failed to ensure log directory: %v", err)
	}

	components, cleanup, err := InitializeSeed(config.ConfigInput{
		Path: "config",
		Name: "app.dev",
		Type: "yaml",
	})
	if err != nil {
		log.Fatalf("Failed to initialize seed: %v", err)
	}
	defer cleanup()

	db := components.DB
	logger := components.Logger

	logger.Info("开始数据填充")

	// 设置随机种子
	gofakeit.Seed(time.Now().UnixNano())

	// 开始填充数据
	startTime := time.Now()
	if err := seedData(db, logger); err != nil {
		logger.Error("数据填充失败", applog.Error(err))
		log.Fatalf("Failed to seed data: %v", err)
	}

	duration := time.Since(startTime)
	logger.Info("数据填充完成",
		applog.Any("耗时", duration),
		applog.Int("用户数量", numUsers),
		applog.Int("电影数量", numMovies),
		applog.Int("影厅数量", numCinemaHalls),
		applog.Int("场次排期天数", 30),
		applog.String("用户凭证文件", filepath.Join("test", "performance", "data", "user_credentials.json")),
	)
}

func seedData(db *gorm.DB, logger applog.Logger) error {
	// 1. 获取预定义的角色
	logger.Info("开始获取角色数据")
	var roles []models.RoleGorm
	if err := db.Where("name IN ?", []string{user.AdminRoleName, user.UserRoleName}).Find(&roles).Error; err != nil {
		return fmt.Errorf("failed to fetch roles: %v", err)
	}
	logger.Info("角色数据获取成功", applog.Int("角色数量", len(roles)))
	logger.Info("角色数据", applog.Any("角色", roles))

	// 2. 创建用户
	logger.Info("开始创建用户数据", applog.Int("计划创建数量", numUsers))
	users := createUsers(roles)
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to create users: %v", err)
	}
	logger.Info("用户数据创建完成", applog.Int("实际创建数量", len(users)))

	// 3. 创建电影类型
	logger.Info("开始创建电影类型数据")
	genres := createGenres()
	if err := db.Create(&genres).Error; err != nil {
		return fmt.Errorf("failed to create genres: %v", err)
	}
	logger.Info("电影类型数据创建完成", applog.Int("数量", len(genres)))

	// 4. 创建电影
	logger.Info("开始创建电影数据", applog.Int("计划创建数量", numMovies))
	movies := createMovies(genres)
	if err := db.Create(&movies).Error; err != nil {
		return fmt.Errorf("failed to create movies: %v", err)
	}
	logger.Info("电影数据创建完成", applog.Int("实际创建数量", len(movies)))

	// 5. 创建影厅
	logger.Info("开始创建影厅数据", applog.Int("计划创建数量", numCinemaHalls))
	halls := createCinemaHalls()
	if err := db.Create(&halls).Error; err != nil {
		return fmt.Errorf("failed to create cinema halls: %v", err)
	}
	logger.Info("影厅数据创建完成", applog.Int("实际创建数量", len(halls)))

	// 6. 创建座位
	logger.Info("开始创建座位数据")
	seats := createSeats(halls)
	if err := db.Create(&seats).Error; err != nil {
		return fmt.Errorf("failed to create seats: %v", err)
	}
	logger.Info("座位数据创建完成", applog.Int("数量", len(seats)))

	// 7. 创建场次
	logger.Info("开始创建场次数据")
	showtimes := createShowtimes(movies, halls)
	if err := db.Create(&showtimes).Error; err != nil {
		return fmt.Errorf("failed to create showtimes: %v", err)
	}
	logger.Info("场次数据创建完成", applog.Int("实际创建数量", len(showtimes)))

	return nil
}

func createUsers(roles []models.RoleGorm) []models.UserGorm {
	users := make([]models.UserGorm, 0, numUsers)
	credentials := make([]UserCredential, 0, numUsers)

	for i := 0; i < numUsers; i++ {
		username := gofakeit.Username()
		password := gofakeit.Password(true, true, true, true, false, 10)
		email := gofakeit.Email()

		// 使用较低的哈希成本以降低生成开销
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

		// 99% 的用户是普通用户，1% 是管理员
		var roleID uint
		var roleName string
		if rand.Float32() < 0.99 {
			// 找到 USER 角色
			for _, role := range roles {
				if role.Name == user.UserRoleName {
					roleID = role.ID
					roleName = role.Name
					break
				}
			}
		} else {
			// 找到 ADMIN 角色
			for _, role := range roles {
				if role.Name == user.AdminRoleName {
					roleID = role.ID
					roleName = role.Name
					break
				}
			}
		}

		user := models.UserGorm{
			Username:     username,
			Email:        email,
			PasswordHash: string(hashedPassword),
			RoleID:       roleID,
		}
		users = append(users, user)

		// 保存用户凭证
		credential := UserCredential{
			Username: username,
			Password: password,
			Email:    email,
			Role:     roleName,
		}
		credentials = append(credentials, credential)
	}

	// 保存用户凭证到文件
	saveUserCredentials(credentials)

	return users
}

// saveUserCredentials 将用户凭证保存到文件
func saveUserCredentials(credentials []UserCredential) {
	// 确保目录存在
	dataDir := filepath.Join("test", "performance", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("创建目录失败: %v", err)
		return
	}

	// 将凭证写入文件
	filePath := filepath.Join(dataDir, "user_credentials.json")
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		log.Printf("序列化用户凭证失败: %v", err)
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("保存用户凭证失败: %v", err)
		return
	}
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
	// 使用map来跟踪已使用的标题
	usedTitles := make(map[string]bool)

	for i := 0; i < numMovies; {
		title := gofakeit.MovieName()
		// 如果标题已存在，继续生成新的标题
		if usedTitles[title] {
			continue
		}
		usedTitles[title] = true

		selectedGenre := &genres[rand.Intn(len(genres))]
		movie := models.MovieGorm{
			Title:       title,
			Description: gofakeit.Sentence(10),
			// DurationMinutes: gofakeit.Number(60, 180),
			ReleaseDate: time.Date(gofakeit.Number(2000, 2025), time.Month(gofakeit.Number(1, 12)), gofakeit.Number(1, 28), 0, 0, 0, 0, time.UTC),
			Rating:      float32(gofakeit.Float32Range(1, 5)),
			AgeRating:   "PG-13",
			Cast:        gofakeit.Name() + ", " + gofakeit.Name(),
			Genres:      []*models.GenreGorm{selectedGenre},
		}
		movies = append(movies, movie)
		i++ // 只有在成功添加电影后才增加计数器
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
					RowIdentifier: string(rune('A' - 1 + row)),
					SeatNumber:    fmt.Sprintf("%d", col),
					Type:          cinema.SeatTypeStandard,
				}
				seats = append(seats, seat)
			}
		}
	}
	return seats
}

func createShowtimes(movies []models.MovieGorm, halls []models.CinemaHallGorm) []models.ShowtimeGorm {
	var showtimes []models.ShowtimeGorm

	// 设置基本参数
	const (
		workStartHour   = 10  // 影院营业开始时间（10:00）
		workEndHour     = 22  // 影院营业结束时间（22:00）
		daysToSchedule  = 30  // 排期天数
		movieDuration   = 120 // 电影固定时长120分钟
		cleaningTime    = 30  // 清场时间30分钟
		advertisingTime = 15  // 广告和预告片时间15分钟
		totalSlotTime   = movieDuration + cleaningTime + advertisingTime
	)

	// 计算每天可以放映的场次数
	workMinutesPerDay := (workEndHour - workStartHour) * 60
	slotsPerDay := workMinutesPerDay / totalSlotTime

	// 获取当前日期，并设置到当天的营业开始时间
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), workStartHour, 0, 0, 0, now.Location())

	// 为每个影厅创建一个月的排期
	for _, hall := range halls {
		// 遍历30天
		for day := 0; day < daysToSchedule; day++ {
			currentDate := startDate.AddDate(0, 0, day)

			// 创建当天的所有场次
			for slot := 0; slot < slotsPerDay; slot++ {
				// 计算本场次的开始时间
				slotStartTime := currentDate.Add(time.Duration(slot*totalSlotTime) * time.Minute)

				// 确保不超过营业结束时间
				if slotStartTime.Hour() >= workEndHour {
					continue
				}

				// 随机选择一部电影
				movie := movies[rand.Intn(len(movies))]

				showtime := models.ShowtimeGorm{
					MovieID:      movie.ID,
					CinemaHallID: hall.ID,
					StartTime:    slotStartTime,
					EndTime:      slotStartTime.Add(time.Duration(movieDuration) * time.Minute),
					Price:        float64(gofakeit.Number(3000, 8000)) / 100, // 30-80元
				}

				showtimes = append(showtimes, showtime)
			}
		}
	}

	// 按开始时间排序
	sort.Slice(showtimes, func(i, j int) bool {
		return showtimes[i].StartTime.Before(showtimes[j].StartTime)
	})

	return showtimes
}
