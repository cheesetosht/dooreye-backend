package api

import (
	"dooreye-backend/internal/model"
	"dooreye-backend/internal/store"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateVisitRequest struct {
	Phone       string            `json:"phone" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	PhotoURL    string            `json:"photo_url"`
	Type        model.VisitorType `json:"type" binding:"required"`
	Purpose     string            `json:"purpose"`
	ResidenceID *int64            `json:"residence_id"`
}

func (h *Handler) createVisitAsSecurity(c *gin.Context) {
	var req CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("user_id")

	tx, err := h.db.BeginTx(c)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback(c)

	var visitorID string
	err = tx.QueryRow(c, `
		INSERT INTO visitors (name, phone, photo_url, type, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.Name, req.Phone, req.PhotoURL, req.Type, userID).Scan(&visitorID)

	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	var visit model.Visit
	err = tx.QueryRow(c, `
		INSERT INTO visits (residence_id, visitor_id, checked_in_by, check_in_time, purpose)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, residence_id, visitor_id, checked_in_by, check_in_time, purpose
	`, req.ResidenceID, visitorID, userID, time.Now(), req.Purpose).Scan(
		&visit.ID, &visit.ResidenceID, &visit.VisitorID, &visit.CheckedInBy,
		&visit.CheckInTime, &visit.Purpose,
	)

	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(c); err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, visit)
}

func (h *Handler) getVisitorByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
		return
	}

	visitor, err := h.db.GetVisitorByPhone(c.Request.Context(), phone)
	if err != nil {
		if err == store.ErrNotFound {
			h.respondError(c, http.StatusNotFound, err)
			return
		}
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": visitor})
}

type CreatePreApprovedVisitorRequest struct {
	Name            string     `json:"name" binding:"required"`
	Phone           string     `json:"phone" binding:"required"`
	PhotoURL        string     `json:"photo_url"`
	Type            string     `json:"type" binding:"required"`
	PreApprovedTill *time.Time `json:"pre_approved_till"`
}

func (h *Handler) createPreApprovedVisitor(c *gin.Context) {
	var req CreatePreApprovedVisitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, err)
		return
	}

	userID := c.GetString("user_id")

	input := store.PreApprovedVisitor{
		Name:            req.Name,
		Phone:           req.Phone,
		PhotoURL:        req.PhotoURL,
		Type:            req.Type,
		PreApprovedTill: req.PreApprovedTill,
		CreatedBy:       userID,
	}

	visitor, err := h.db.CreatePreApprovedVisitor(c.Request.Context(), input)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": visitor})
}

func (h *Handler) getVisits(c *gin.Context) {
	var filter store.VisitFilter

	if residenceID := c.Query("residence_id"); residenceID != "" {
		id, err := strconv.ParseInt(residenceID, 10, 64)
		if err != nil {
			h.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid residence_id: %w", err))
			return
		}
		filter.ResidenceID = &id
	}

	filter.OnlyOngoing = c.Query("ongoing") == "true"

	visits, err := h.db.GetVisits(c.Request.Context(), filter)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": visits})
}
