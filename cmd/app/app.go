package app

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type App struct {
	Logger zap.Logger
	// config string
	DB *pgxpool.Pool
}

var pgOnce sync.Once
var dbInstance *pgxpool.Pool

func InitializeApp() *App {
	fmt.Println("initiating application")
	initializeDB()
	logger := initializeLogger()

	app := App{Logger: *logger, DB: dbInstance}
	app.Logger.Info("app initiated")
	return &app
}

func initializeDB() {
	fmt.Println("initiating db")

	postgresConnectionUrl := "postgresql://localhost/z_cricket"

	// sync once does not return value
	pgOnce.Do(func() {
		pgDB, err := pgxpool.New(context.Background(), postgresConnectionUrl)
		if err != nil {
			log.Fatal("error in connecting to postgres")
		}
		dbInstance = pgDB
		_ = dbInstance.Ping(context.Background())
		//defer dbInstance.Close()

		var greeting string

		err = dbInstance.QueryRow(context.Background(), "select 'Hello world'").Scan(&greeting)
		if err != nil {
			log.Fatal("error in querying", err.Error())
		}
		fmt.Println(greeting, "greeting here")
	})
}

func initializeLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	return logger
}
