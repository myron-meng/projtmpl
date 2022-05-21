package handler

import (
	"net/http"

	"projtmpl/internal/repository"
	"projtmpl/internal/service"

	"github.com/gofiber/fiber/v2"
)

const (
	userCtxKey        = "user"
	userSessionCtxKey = "user_session"
)

type Handler struct {
	Repositories *repository.Repositories
	Services     *service.Services
}

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=1,max=16"`
	Password struct {
		Password string `json:"password" validate:"required,min=8,max=32"`
	} `json:"password" validate:"required"`
	DateOfBirth string `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Gender      string `json:"gender" validate:"required,oneof=MALE FEMALE"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return SimpleCodeResponse(c, http.StatusBadRequest, err)
	}
	if err := validate.Struct(&req); err != nil {
		return BadRequest(c, err)
	}

	return SimpleCodeResponse(c, http.StatusOK)
}
