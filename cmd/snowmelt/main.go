package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lacsar712/snowmelt/internal/app"
	"github.com/lacsar712/snowmelt/internal/web"
	"github.com/lacsar712/snowmelt/internal/web/api"
)

func main() {
	cfgPath := flag.String("config", "", "path to JSON config file")
	flag.Parse()

	application, err := app.Bootstrap(*cfgPath)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	if err := application.EnsureDataDir(); err != nil {
		log.Fatalf("data dir: %v", err)
	}

	handler, err := web.Handler(api.NewServer(application))
	if err != nil {
		log.Fatalf("web: %v", err)
	}

	cfg := application.Config()
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	ctx, stopTick := context.WithCancel(context.Background())
	defer stopTick()
	application.StartTickLoop(ctx)

	go func() {
		log.Printf("snowmelt listening on %s (unit %s)", cfg.ListenAddr, application.UnitID())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	application.StopTickLoop()
	stopTick()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
