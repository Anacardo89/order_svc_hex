package orderorchestrator

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
)

type Server struct {
	httpSrv         *http.Server
	router          http.Handler
	addr            string
	ShutdownTimeout time.Duration
}

func NewServer(cfg *config.Server, h *OrderHandler) *Server {
	s := &Server{
		router: NewRouter(h),
		addr:   fmt.Sprintf(":%s", cfg.Port),
	}
	s.httpSrv = &http.Server{
		Addr:              s.addr,
		Handler:           s.router,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	return s
}

func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}
