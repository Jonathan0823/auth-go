package handler

import (
	"net/http"
	"strconv"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (h *MainHandler) GetUserByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.Error(errors.BadRequest("Invalid user ID", nil))
		return
	}

	user, err := h.svc.GetUserByID(id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "user": user})
}

func (h *MainHandler) GetAllUsers(c *gin.Context) {
	users, err := h.svc.GetAllUsers()
	if err != nil {
		c.Error(err)
		return
	}

	if users == nil {
		users = []*models.User{}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users retrieved successfully", "users": users})
}

func (h *MainHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")

	user, err := h.svc.GetUserByEmail(email)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "user": user})
}

func (h *MainHandler) UpdateUser(c *gin.Context) {
	var user models.UpdateUserRequest
	if isValid := utils.BindJSONWithValidation(c, &user); !isValid {
		return
	}

	if err := h.svc.UpdateUser(user, c); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *MainHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.Error(errors.BadRequest("Invalid user ID", nil))
		return
	}

	if err := h.svc.DeleteUser(id, c); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *MainHandler) GetCurrentUser(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		c.Error(errors.Unauthorized("User is not authenticated", err))
		return
	}

	data, err := h.svc.GetUserByID(user.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Current user retrieved successfully", "user": data})
}
