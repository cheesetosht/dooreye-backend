package api

import (
	"dooreye-backend/internal/model"
	"dooreye-backend/internal/store"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	AccessCode  string         `json:"access_code" binding:"required"`
	DeviceID    string         `json:"device_id" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	ResidenceID *int64         `json:"residence_id"`
	SocietyID   *int64         `json:"society_id"`
	Role        model.UserRole `json:"role" binding:"required"`
}

func (h *Handler) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, err)
		return
	}

	params := store.CreateUserParams{
		AccessCode:  req.AccessCode,
		DeviceID:    req.DeviceID,
		Name:        req.Name,
		ResidenceID: req.ResidenceID,
		SocietyID:   req.SocietyID,
		Role:        req.Role,
		ActivatedBy: c.GetString("user_id"),
	}

	user, err := h.db.CreateUser(c.Request.Context(), params)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, store.ErrInvalidUserType):
			status = http.StatusBadRequest
		case errors.Is(err, store.ErrDuplicateAccessCode):
			status = http.StatusConflict
		}
		h.respondError(c, status, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}
