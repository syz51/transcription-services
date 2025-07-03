package handler

import (
	"net/http"

	"github.com/syz51/media-validation-service/internal/config"

	"github.com/labstack/echo/v4"
)

// Handler contains all the handlers
type Handler struct {
	config *config.Config
}

// New creates a new handler instance
func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

func (h *Handler) Events(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}
