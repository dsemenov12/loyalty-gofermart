package main

import (
	"fmt"
	"net/http"
	"database/sql"
	"errors"

	"github.com/dsemenov12/loyalty-gofermart/internal/config"
	"github.com/dsemenov12/loyalty-gofermart/internal/storage"
	"github.com/dsemenov12/loyalty-gofermart/internal/storage/pg"
	"github.com/dsemenov12/loyalty-gofermart/internal/handlers"
	"github.com/dsemenov12/loyalty-gofermart/internal/logger"
	"github.com/dsemenov12/loyalty-gofermart/internal/middlewares/loggerhandler"
	"github.com/dsemenov12/loyalty-gofermart/internal/middlewares/gziphandler"
	"github.com/dsemenov12/loyalty-gofermart/internal/middlewares/authhandler"
	"go.uber.org/zap"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if error := run(); error != nil {
        fmt.Println(error)
    }
}

func run() error {
	var storage storage.Storage

	if config.FlagDatabaseURI == "" {
		return errors.New("empty database DSN")
	}

	// Подключение к БД
	conn, err := sql.Open("pgx", config.FlagDatabaseURI)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Запуск миграций
	if err = upMigrations(conn); err != nil {
		return err
	}

	storage = pg.NewStorage(conn)
	app := handlers.NewApp(storage)

	if err = logger.Initialize(config.FlagLogLevel); err != nil {
        return err
    }
	logger.Log.Info("Running server", zap.String("address", config.FlagRunAddr))

	router := chi.NewRouter()

	router.Post("/api/user/register", loggerhandler.RequestLogger(app.UserRegister))
	router.Post("/api/user/login", loggerhandler.RequestLogger(app.UserLogin))
	router.Post("/api/user/orders", loggerhandler.RequestLogger(authhandler.AuthHandle(app.UserUploadOrder)))
	router.Get("/api/user/orders", loggerhandler.RequestLogger(authhandler.AuthHandle(app.UserGetOrders)))
	router.Get("/api/user/balance", loggerhandler.RequestLogger(authhandler.AuthHandle(app.GetUserBalance)))
	router.Post("/api/user/balance/withdraw", loggerhandler.RequestLogger(authhandler.AuthHandle(app.WithdrawUserBalance)))
	router.Get("/api/user/withdrawals", loggerhandler.RequestLogger(authhandler.AuthHandle(app.GetUserWithdrawals)))

	err = http.ListenAndServe(
		config.FlagRunAddr,
		gziphandler.GzipHandle(router),
	)
    if err != nil {
        return err
    }

	return nil
}

// Запуск миграций
func upMigrations(conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}
	m.Up()

	return nil
}

func init() {
	config.ParseFlags()
}