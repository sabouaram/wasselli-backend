package handlers

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"wasselli-backend/internal/http/middlewares"
)

func (h *Handler) Serve() {

	if h.Storage == nil || h.Minio == nil || h.Logger == nil || h.Emailing == nil ||
		h.Config == nil {
		panic("api handler instances are nil")
	}

	h.Mux.Post(
		"/api/v1/login",
		middlewares.JwtMiddleware(func(writer http.ResponseWriter, request *http.Request) {

		} /*Example: h.HandleLogin*/))

	listenAddress := h.Config.GetString("server.listen")

	h.Logger.Info("api server listening on:", zap.Any("address =>", listenAddress))

	if err := http.ListenAndServe(listenAddress, h.Mux); err != nil {
		h.Logger.Fatal("Server error:", zap.Any("error =>", err))
	}
}

func (h *Handler) Shutdown() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	server := &http.Server{
		Addr:    h.Config.GetString("server.listen"),
		Handler: h.Mux,
	}

	if err := server.Shutdown(ctx); err != nil {
		h.Logger.Error("server shutdown error:", zap.Error(err))
	}

	h.Logger.Info("handler shutdown complete")
}
