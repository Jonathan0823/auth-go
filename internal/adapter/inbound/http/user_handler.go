package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func unauthenticatedError() error {
	return fmt.Errorf("user is not authenticated: %w", domain.ErrUnauthenticated)
}

// GetUserByID returns a user by numeric ID.
// @Summary Get user by ID
// @ID getUserByID
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path int true "User ID"
// @Success 200 {object} UserResponseEnvelope
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/user/{id} [get]
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

// GetAllUsers returns all users.
// @Summary List users
// @ID getAllUsers
// @Tags users
// @Produce json
// @Security CookieAuth
// @Success 200 {object} UsersResponseEnvelope
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/user/get-all [get]
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

// GetUserByEmail returns a user by email address.
// @Summary Get user by email
// @ID getUserByEmail
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param email query string true "User email address"
// @Success 200 {object} UserResponseEnvelope
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/user/email [get]
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

// UpdateUser updates the authenticated user's profile.
// @Summary Update current user
// @ID updateUser
// @Tags users
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body dto.UpdateUserRequest true "Profile updates"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/user/update [patch]
func (h *Handler) UpdateUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.UpdateUserRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}
	currentUser, err := GetUser(c)
	if err != nil {
		c.Error(unauthenticatedError())
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

// DeleteUser deletes a user by ID when authorized by the authenticated user.
// @Summary Delete user
// @ID deleteUser
// @Tags users
// @Produce json
// @Security CookieAuth
// @Param id path int true "User ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/user/delete/{id} [delete]
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
		c.Error(unauthenticatedError())
		return
	}
	if err := h.Svc.User.Delete(ctx, id, currentUser.ID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetCurrentUser returns the authenticated user's profile.
// @Summary Get current user
// @ID getCurrentUser
// @Tags users
// @Produce json
// @Security CookieAuth
// @Success 200 {object} UserResponseEnvelope
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/user/me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	user, err := GetUser(c)
	if err != nil {
		c.Error(unauthenticatedError())
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
