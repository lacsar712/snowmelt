package api

import "github.com/lacsar712/snowmelt/internal/app"

type Server struct {
	app *app.App
}

func NewServer(application *app.App) *Server {
	return &Server{app: application}
}

func (s *Server) App() *app.App { return s.app }
