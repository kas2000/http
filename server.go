package http

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"github.com/kas2000/logger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	GracefulTimeout time.Duration
	ApiVersion      string
	Timeout         time.Duration
	Logger          logger.Logger
}

type Server interface {
	ListenAndServe() error
	Handle(method string, path string, fn func(w http.ResponseWriter, r *http.Request))
}

type server struct {
	r   *mux.Router
	srv *http.Server
	log logger.Logger
	cfg Config
}

func NewServer(config Config) Server {
	r := mux.NewRouter().StrictSlash(true)
	s := &server{cfg: config, log: config.Logger, r: r}
	s.srv = &http.Server{
		Addr: config.Addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * config.Timeout,
		ReadTimeout:  time.Second * config.Timeout,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}
	return s
}

func (s *server) Handle(method string, path string, fn func(w http.ResponseWriter, r *http.Request)) {
	s.r.HandleFunc(path, Logging(fn, s.log)).Methods(method)
}

func (s *server) ListenAndServe() error {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second* s.cfg.GracefulTimeout, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	log := s.log

	log.Info("Starting server...")

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err.Error())
			}
		}
	}()

	log.Info("Server started on address: " + s.cfg.Addr)
	// Block until we receive our signal.
	<-c

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Warn(err.Error())
	}

	log.Info("The server is down")

	return nil
}
