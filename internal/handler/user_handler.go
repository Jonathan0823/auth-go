package handler

import (
	"net/http"
	"strconv"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (h *MainHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.svc.GetUserByID(idInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "user": user})
}

func (h *MainHandler) GetAllUsers(c *gin.Context) {
	users, err := h.svc.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *MainHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.svc.DeleteUser(idInt, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *MainHandler) GetCurrentUser(c *gin.Context) {
	user, err := utils.GetUser(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current user"})
		return
	}

	data, err := h.svc.GetUserByID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Current user retrieved successfully", "user": data})
}
