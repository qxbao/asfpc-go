package server

import (
	"database/sql"
	"log"
	"os"
	"github.com/labstack/echo/v4"
	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
	"github.com/qxbao/asfpc/routes"
	_ "github.com/lib/pq"
)

type Server struct {
	infras.Server
}

func (s *Server) Run() {
	if err := s.initDB(); err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	s.start()
}

func (s *Server) start() {
	e := echo.New()
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
		log.Println("Connected to the database successfully!")
	}

	s.Database = database
	s.Queries = db.New(database)

	return nil
}

func (s Server) initRoute() error {
	routes.InitAccountRoutes(s.Server)
	routes.InitScanRoutes(s.Server)
	return nil
}
