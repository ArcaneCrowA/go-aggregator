package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mockService struct {
	addFn  func(ctx context.Context, m models.AddSubscriptionDTO) error
	calcFn func(ctx context.Context, q models.CalculateQuery) (int, error)
}

func (m *mockService) AddSubscription(ctx context.Context, dto models.AddSubscriptionDTO) error {
	return m.addFn(ctx, dto)
}

func (m *mockService) Calculate(ctx context.Context, q models.CalculateQuery) (int, error) {
	return m.calcFn(ctx, q)
}

func setupTest() (*handler, *mockService) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := New(logger, svc)
	return h, svc
}

func TestAddSubscriptionHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, svc := setupTest()
		svc.addFn = func(_ context.Context, m models.AddSubscriptionDTO) error {
			return nil
		}

		body, _ := json.Marshal(map[string]any{
			"service_name": "netflix",
			"price":        1999,
			"user_id":      uuid.New().String(),
			"start_date":   "01-2026",
		})
		req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.AddSubscription(c)

		if w.Code != http.StatusAccepted {
			t.Errorf("got status %d, want %d", w.Code, http.StatusAccepted)
		}
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["status"] != "accepted" {
			t.Errorf("got status %q, want %q", resp["status"], "accepted")
		}
	})

	t.Run("bad json body", func(t *testing.T) {
		h, _ := setupTest()

		req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader([]byte("{bad")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.AddSubscription(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		h, _ := setupTest()

		body, _ := json.Marshal(map[string]any{
			"price":   1999,
			"user_id": uuid.New().String(),
			// missing service_name (required) and start_date (required)
		})
		req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.AddSubscription(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error for bad date format", func(t *testing.T) {
		h, _ := setupTest()

		body, _ := json.Marshal(map[string]any{
			"service_name": "netflix",
			"price":        1999,
			"user_id":      uuid.New().String(),
			"start_date":   "2026-01-15T00:00:00Z",
		})
		req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.AddSubscription(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("service error", func(t *testing.T) {
		h, svc := setupTest()
		svc.addFn = func(_ context.Context, m models.AddSubscriptionDTO) error {
			return errors.New("db unavailable")
		}

		body, _ := json.Marshal(map[string]any{
			"service_name": "netflix",
			"price":        1999,
			"user_id":      uuid.New().String(),
			"start_date":   "01-2026",
		})
		req := httptest.NewRequest(http.MethodPost, "/service", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.AddSubscription(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestGetSubscriptionsFilterHandler(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("success with all params", func(t *testing.T) {
		h, svc := setupTest()
		svc.calcFn = func(_ context.Context, q models.CalculateQuery) (int, error) {
			if q.Name != "netflix" {
				t.Errorf("got name %q, want %q", q.Name, "netflix")
			}
			if q.StartDate == nil || !q.StartDate.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
				t.Errorf("unexpected start_date: %v", q.StartDate)
			}
			return 5000, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/service?service_name=netflix&user_id="+validUUID+"&start_date=01-2026", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.GetSubscriptionsFilter(c)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["status"] != "accepted" {
			t.Errorf("got status %q, want %q", resp["status"], "accepted")
		}
		if total, ok := resp["total"].(float64); !ok || total != 5000 {
			t.Errorf("got total %v, want 5000", resp["total"])
		}
	})

	t.Run("omits optional date params when not provided", func(t *testing.T) {
		h, svc := setupTest()
		svc.calcFn = func(_ context.Context, q models.CalculateQuery) (int, error) {
			if q.StartDate != nil {
				t.Errorf("expected nil start_date, got %v", q.StartDate)
			}
			if q.EndDate != nil {
				t.Errorf("expected nil end_date, got %v", q.EndDate)
			}
			return 0, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/service?service_name=netflix&user_id="+validUUID, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.GetSubscriptionsFilter(c)
	})

	t.Run("invalid user_id", func(t *testing.T) {
		h, _ := setupTest()

		req := httptest.NewRequest(http.MethodGet, "/service?service_name=netflix&user_id=not-a-uuid", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.GetSubscriptionsFilter(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing user_id", func(t *testing.T) {
		h, _ := setupTest()

		req := httptest.NewRequest(http.MethodGet, "/service?service_name=netflix", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.GetSubscriptionsFilter(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("service error", func(t *testing.T) {
		h, svc := setupTest()
		svc.calcFn = func(_ context.Context, q models.CalculateQuery) (int, error) {
			return 0, errors.New("calc failed")
		}

		req := httptest.NewRequest(http.MethodGet, "/service?service_name=netflix&user_id="+validUUID, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h.GetSubscriptionsFilter(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}
