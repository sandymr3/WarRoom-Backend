package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"war-room-backend/internal/broadcast"
	"war-room-backend/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// ============================================
// BATCH HANDLER
// ============================================

type BatchHandler struct {
	BatchService *services.BatchService
}

func NewBatchHandler(bs *services.BatchService) *BatchHandler {
	return &BatchHandler{BatchService: bs}
}

// POST /api/batches/validate
// Public endpoint - validates a batch code before registration/login.
func (h *BatchHandler) ValidateCode(c echo.Context) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	batch, err := h.BatchService.ValidateCode(req.Code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusOK, map[string]any{"valid": false, "error": "Invalid or inactive batch code"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"valid": true,
		"batch": map[string]any{
			"code":  batch.Code,
			"name":  batch.Name,
			"level": batch.Level,
		},
	})
}

// ============================================
// ADMIN CRUD
// ============================================

// POST /api/admin/batches  (admin only)
func (h *BatchHandler) CreateBatch(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	claims := userToken.Claims.(jwt.MapClaims)
	adminID := claims["user_id"].(string)

	var req struct {
		Code     string     `json:"code"`
		Name     string     `json:"name"`
		Level    int        `json:"level"`
		StartsAt *time.Time `json:"startsAt"`
		EndsAt   *time.Time `json:"endsAt"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "code is required"})
	}
	if req.Level == 0 {
		req.Level = 1
	}
	batch, err := h.BatchService.CreateBatch(req.Code, req.Name, req.Level, adminID, req.StartsAt, req.EndsAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create batch: " + err.Error()})
	}
	return c.JSON(http.StatusCreated, batch)
}

// GET /api/admin/batches  (admin only)
func (h *BatchHandler) ListBatches(c echo.Context) error {
	batches, err := h.BatchService.ListBatches()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not list batches"})
	}
	return c.JSON(http.StatusOK, batches)
}

// GET /api/admin/batches/:id  (admin only)
func (h *BatchHandler) GetBatchDetail(c echo.Context) error {
	id := c.Param("id")
	batch, err := h.BatchService.GetBatch(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Batch not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	return c.JSON(http.StatusOK, batch)
}

// PATCH /api/admin/batches/:id  (admin only)
func (h *BatchHandler) UpdateBatch(c echo.Context) error {
	id := c.Param("id")
	var input services.UpdateBatchInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	batch, err := h.BatchService.UpdateBatch(id, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Batch not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update batch"})
	}
	return c.JSON(http.StatusOK, batch)
}

// DELETE /api/admin/batches/:id  (admin only)
func (h *BatchHandler) DeleteBatch(c echo.Context) error {
	id := c.Param("id")
	if err := h.BatchService.DeleteBatch(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not delete batch"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Batch deleted"})
}

// GET /api/admin/batches/:id/participants  (admin only)
func (h *BatchHandler) GetBatchParticipants(c echo.Context) error {
	id := c.Param("id")
	batch, err := h.BatchService.GetBatch(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Batch not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	participants, err := h.BatchService.GetBatchParticipants(batch.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not load participants"})
	}
	return c.JSON(http.StatusOK, participants)
}

// GET /api/admin/batches/:id/stats  (admin only)
func (h *BatchHandler) GetBatchStats(c echo.Context) error {
	id := c.Param("id")
	batch, err := h.BatchService.GetBatch(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Batch not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	stats, err := h.BatchService.GetBatchStats(batch.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not load stats"})
	}
	return c.JSON(http.StatusOK, stats)
}

// ============================================
// LEADERBOARD
// ============================================
func (h *BatchHandler) GetLeaderboard(c echo.Context) error {
	code := strings.ToUpper(c.Param("code"))
	entries, err := h.BatchService.GetLeaderboard(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"batchCode": code,
		"entries":   entries,
	})
}

// ============================================
// WEBSOCKET LEADERBOARD
// ============================================

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WS /api/batches/:code/live
func (h *BatchHandler) LiveLeaderboard(c echo.Context) error {
	code := strings.ToUpper(c.Param("code"))

	conn, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Register client
	broadcast.Register(code, conn)
	defer broadcast.Unregister(code, conn)

	// Send initial snapshot
	if entries, err := h.BatchService.GetLeaderboard(code); err == nil {
		iEntries := make([]interface{}, len(entries))
		for i, e := range entries {
			iEntries[i] = e
		}
		msg, _ := json.Marshal(map[string]any{
			"type":      "leaderboard",
			"batchCode": code,
			"entries":   iEntries,
			"updatedAt": time.Now().UTC(),
		})
		conn.WriteMessage(websocket.TextMessage, msg)
	}

	done := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(done)
				return
			}
		}
	}()

	<-done
	return nil
}
