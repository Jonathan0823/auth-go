package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func (h *Handler) GetUserByID(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.Error(domain.BadRequest("Invalid user ID", nil))
		return
	}
	user, err := h.Svc.User.GetByID(ctx, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "user": user})
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	users, err := h.Svc.User.GetAll(ctx)
	if err != nil {
		c.Error(err)
		return
	}
	if users == nil {
		users = []*domain.User{}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Users retrieved successfully", "users": users})
}

func (h *Handler) GetUserByEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	email := c.Query("email")
	if email == "" {
		c.Error(domain.BadRequest("Email query parameter is required", nil))
		return
	}
	user, err := h.Svc.User.GetByEmail(ctx, email)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "user": user})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req domain.UpdateUserRequest
	if isValid := BindJSONWithValidation(c, &req); !isValid {
		return
	}
	currentUser, err := GetUser(c)
	if err != nil {
		c.Error(domain.Unauthorized("User is not authenticated", err))
		return
	}
	if err := h.Svc.User.Update(ctx, currentUser.ID, req); err != nil {
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
		c.Error(domain.BadRequest("Invalid user ID", nil))
		return
	}
	currentUser, err := GetUser(c)
	if err != nil {
		c.Error(domain.Unauthorized("User is not authenticated", err))
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
		c.Error(domain.Unauthorized("User is not authenticated", err))
		return
	}
	data, err := h.Svc.User.GetByID(ctx, user.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Current user retrieved successfully", "user": data})
}
