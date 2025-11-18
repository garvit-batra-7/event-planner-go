package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token + user info
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body      LoginRequest  true  "Login credentials"
// @Success      200        {object}  AuthResponse
// @Failure      400        {object}  map[string]string  "Invalid request payload"
// @Failure      401        {object}  map[string]string  "Invalid email or password"
// @Failure      500        {object}  map[string]string  "Internal server error"
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		if err.Error() == "invalid email or password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Register godoc
// @Summary      User registration
// @Description  Create a new user account and return JWT token + user info
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body      RegisterRequest  true  "User registration details"
// @Success      201  {object}  AuthResponse
// @Failure      400  {object}  map[string]string  "Invalid payload"
// @Failure      409  {object}  map[string]string  "User with this email already exists"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetByID returns a user by ID
// @Summary      Get a user by ID
// @Description  Retrieves a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserDTO
// @Failure      400  {object}  map[string]string  "Invalid user ID"
// @Failure      404  {object}  map[string]string  "User not found"
// @Failure      500  {object}  map[string]string  "Failed to get user"
// @Router       /api/v1/users/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
