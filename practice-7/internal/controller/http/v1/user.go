package v1

import (
	"net/http"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	u usecase.UserInterface
}

func New(r *gin.Engine, u usecase.UserInterface) {
	h := &Handler{u}

	api := r.Group("/api/v1/users")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
		api.POST("/verify", h.Verify)
		api.POST("/refresh", h.Refresh)

		protected := api.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		protected.Use(utils.RateLimit())
		{
			protected.GET("/me", h.Me)
			protected.POST("/logout", h.Logout)
			protected.PATCH("/promote/:id",
				utils.RoleMiddleware("admin"),
				h.Promote)
		}
	}
}

func (h *Handler) Register(c *gin.Context) {
	var dto entity.CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := utils.HashPassword(dto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := entity.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: hash,
		Role:     "user",
	}

	if err := h.u.Register(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email for verification code.",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var dto entity.LoginUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, refresh, err := h.u.Login(&dto)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccess, err := utils.GenerateAccessTokenFromRefresh(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccess,
	})
}

func (h *Handler) Verify(c *gin.Context) {
	var dto entity.VerifyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.u.Verify(dto.Email, dto.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *Handler) Me(c *gin.Context) {
	id := c.GetString("userID")
	if id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user, err := h.u.GetMe(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"verified": user.Verified,
	})
}

func (h *Handler) Promote(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	if err := h.u.Promote(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User promoted to admin successfully"})
}

func (h *Handler) Logout(c *gin.Context) {
	id := c.GetString("userID")
	if id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	utils.RDB.Del(utils.Ctx, "auth:"+id)
	utils.RDB.Del(utils.Ctx, "refresh:"+id)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}