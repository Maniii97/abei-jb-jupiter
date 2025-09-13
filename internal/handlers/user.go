package handlers

import (
	"api/internal/services"
	"api/pkg/errors"
	"api/pkg/request"
	"api/pkg/response"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserServiceInterface
	jwtService  services.JWTServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface, jwtService services.JWTServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	user, err := h.userService.Register(context.Background(), req.Email, req.Password, req.FirstName, req.LastName, req.Phone, req.IsAdmin)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userResp := response.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		IsAdmin:   user.IsAdmin,
	}

	response.Success(c, http.StatusCreated, "user registered successfully", userResp)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := request.BindJSON(c, &req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	user, err := h.userService.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.IsAdmin)
	if err != nil {
		h.handleError(c, err)
		return
	}

	loginResp := response.LoginResponse{
		Token: token,
		User: response.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			IsAdmin:   user.IsAdmin,
		},
	}

	response.JSON(c, http.StatusOK, loginResp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := h.userService.GetByID(context.Background(), userID.(uint))
	if err != nil {
		h.handleError(c, err)
		return
	}

	userResp := response.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		IsAdmin:   user.IsAdmin,
	}

	response.JSON(c, http.StatusOK, userResp)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	// This would be an admin-only endpoint
	// For now, just return a placeholder
	response.Success(c, http.StatusOK, "admin endpoint working", map[string]string{
		"message": "This would list all users",
	})
}

// handleError converts application errors to appropriate HTTP responses
func (h *UserHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case "BAD_REQUEST":
			response.Error(c, http.StatusBadRequest, appErr.Message)
		case "UNAUTHORIZED":
			response.Error(c, http.StatusUnauthorized, appErr.Message)
		case "NOT_FOUND":
			response.Error(c, http.StatusNotFound, appErr.Message)
		case "CONFLICT":
			response.Error(c, http.StatusConflict, appErr.Message)
		case "INTERNAL_ERROR":
			response.Error(c, http.StatusInternalServerError, "internal server error")
		default:
			response.Error(c, http.StatusInternalServerError, "internal server error")
		}
	} else {
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
