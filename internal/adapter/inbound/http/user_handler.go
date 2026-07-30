package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func (h *Handler) GetUserByID(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	user, err := h.Svc.User.GetByID(ctx, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"user":    dto.UserResponseFromDomain(user),
	})
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	users, err := h.Svc.User.GetAll(ctx)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"users":   dto.UserResponsesFromDomain(users),
	})
}

func (h *Handler) GetUserByEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
		return
	}
	user, err := h.Svc.User.GetByEmail(ctx, email)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"user":    dto.UserResponseFromDomain(user),
	})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.UpdateUserRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}
	currentUser, err := GetUser(c)
	if err != nil {
		c.Error(fmt.Errorf("user is not authenticated: %w", domain.ErrUnauthenticated))
		return
	}
	command := domain.UpdateUserCommand{
		ID:        req.ID,
		Username:  req.Username,
		AvatarURL: req.AvatarURL,
		Email:     req.Email,
	}
	if err := h.Svc.User.Update(ctx, currentUser.ID, command); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	currentUser, err := GetUser(c)
	if err != nil {
		c.Error(fmt.Errorf("user is not authenticated: %w", domain.ErrUnauthenticated))
		return
	}
	if err := h.Svc.User.Delete(ctx, id, currentUser.ID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	user, err := GetUser(c)
	if err != nil {
		c.Error(fmt.Errorf("user is not authenticated: %w", domain.ErrUnauthenticated))
		return
	}
	data, err := h.Svc.User.GetByID(ctx, user.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Current user retrieved successfully",
		"user":    dto.UserResponseFromDomain(data),
	})
}
