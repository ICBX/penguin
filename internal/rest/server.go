package rest

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Server struct {
	db  *gorm.DB
	app *fiber.App
}

func New(db *gorm.DB) (s *Server) {
	app := fiber.New(fiber.Config{})
	s = &Server{
		db:  db,
		app: app,
	}

	// TODO: Add routes below 👇
	app.Get("/", s.routeIndex)
	app.Post("/video/add", s.routeVideoAdd)
	// TODO: Add routes above 👆

	return
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}