package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Conf struct {
	TimeoutRead  time.Duration
	TimeoutWrite time.Duration
	TimeoutIdle  time.Duration
}

func ServerConfigs() *Conf {
	return &Conf{
		TimeoutRead:  time.Second * 30,
		TimeoutWrite: time.Second * 30,
		TimeoutIdle:  time.Second * 30,
	}
}

// Server represents the top-level http server. This encapsulates
// the MCP Server object within its handlers using a call to
// setupRoutes().
type Server struct {
	StartTime time.Time
	Svr       *http.Server
}

func NewServer(handlers http.Handler) *Server {
	svrCfgs := ServerConfigs()
	return &Server{
		StartTime: time.Now().UTC(),
		Svr: &http.Server{
			Handler:      handlers,
			Addr:         "localhost:9090",
			ReadTimeout:  svrCfgs.TimeoutRead,
			WriteTimeout: svrCfgs.TimeoutWrite,
			IdleTimeout:  svrCfgs.TimeoutIdle,
		},
	}
}

func secondsToTimeStr(seconds float64) string {
	duration := time.Duration(int64(seconds)) * time.Second
	timeValue := time.Time{}.Add(duration)
	return timeValue.Format("15:04:05")
}

// returns the current run time of the server
// as a HH:MM:SS formatted string.
func (s *Server) RunTime() string {
	return secondsToTimeStr(time.Since(s.StartTime).Seconds())
}

// forcibly shuts down server and returns total run time in seconds.
func (s *Server) Shutdown() (string, error) {
	if err := s.Svr.Close(); err != nil && err != http.ErrServerClosed {
		return "0", fmt.Errorf("server shutdown failed: %v", err)
	}
	return s.RunTime(), nil
}

// starts a server that can be shut down via ctrl-c
func (s *Server) Run() {
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// shutdown signal with grace period of 10 seconds
		// nolint:golint
		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Println("shutdown timed out. forcing exit.")
				if _, err := s.Shutdown(); err != nil {
					log.Fatal(err)
				}
				log.Printf("server run time: %s", s.RunTime())
			}
		}()

		log.Println("shutting down server...")
		if err := s.Svr.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		log.Printf("server run time: %v", s.RunTime())
		serverStopCtx()
	}()

	log.Printf("🛠️ MCP-TLS server is running on at %s...", s.Svr.Addr)
	if err := s.Svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
