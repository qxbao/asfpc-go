package server

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/pkg/cron"
	"github.com/qxbao/asfpc/pkg/logger"
	"github.com/qxbao/asfpc/routes"
)

type Server struct {
	infras.Server
	Cron *cron.CronService
}

func (s *Server) Run() {
	if err := s.initDB(); err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	s.start()
}

func (s *Server) start() {
	if err := logger.InitLogger(false); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.FlushLogger()

	configs := s.loadConfigs()
	s.GlobalConfig = &configs

	s.Cron = &cron.CronService{
		Server: &s.Server,
	}
	s.Cron.Setup()
	s.Cron.Start()
	defer s.Cron.Scheduler.Shutdown()

	e := echo.New()
	e.Use(middleware.CORS())
  e.Use(middleware.Recover())

	s.Echo = e
	if err := s.initRoute(); err != nil {
		log.Fatal("Failed to initialize routes:", err)
	}
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")

	if s.Host == nil {
		s.Host = &HOST
	}

	if s.Port == nil {
		s.Port = &PORT
	}

	defer s.Database.Close()
	e.Logger.Fatal(e.Start(*s.Host + ":" + *s.Port))
}

func (s *Server) initDB() error {
	pgUser := os.Getenv("POSTGRE_USER")
	pgPassword := os.Getenv("POSTGRE_PASSWORD")
	pgDBName := os.Getenv("POSTGRE_DBNAME")
	pgHost := os.Getenv("POSTGRE_HOST")
	pgPort := os.Getenv("POSTGRE_PORT")
	dataSourceName := "postgres://" + pgUser + ":" + pgPassword + "@" + pgHost + ":" + pgPort + "/" + pgDBName + "?sslmode=disable"
	database, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		log.Fatal(err)
	}

	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		loggerName := "SERVER.GO"
		logger.GetLogger(&loggerName).Info("Connected to the database successfully!")
	}

	s.Database = database
	s.Queries = db.New(database)

	return nil
}

func (s Server) initRoute() error {
	routes.InitAccountRoutes(s.Server)
	return nil
}

func (s Server) loadConfigs() map[string]string {
	config := make(map[string]string)
	ctx := context.Background()
	defer ctx.Done()
	loadedConfigs, err := s.Queries.GetAllConfigs(ctx)
	if err != nil {
		loggerName := "loadConfigs"
		logger.GetLogger(&loggerName).Error("Error loading configs:", err)
		return config
	}
	for _, c := range loadedConfigs {
		config[c.Key] = c.Value
	}
	return config
}
