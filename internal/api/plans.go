package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
)

type plansHandler struct {
	pool *datastore.Pool
}

func newPlansHandler(pool *datastore.Pool) *plansHandler {
	return &plansHandler{pool: pool}
}

// listPlansHandler 返回所有活跃的订阅计划及其定价，供用户首次进入系统时选择。
func (h *plansHandler) listPlansHandler(c *gin.Context) {
	ctx := c.Request.Context()

	rows, err := h.pool.Query(ctx,
		queries.PlansListActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type pricingResp struct {
		ID             uuid.UUID `json:"id"`
		DurationMonths int       `json:"duration_months"`
		Price          float64   `json:"price"`
		Currency       string    `json:"currency"`
	}

	plans := make(map[uuid.UUID]*gin.H)
	for rows.Next() {
		var planID uuid.UUID
		var name, description string
		var pricingID *uuid.UUID
		var durationMonths *int
		var price *float64
		var currency *string

		if err := rows.Scan(&planID, &name, &description, &pricingID, &durationMonths, &price, &currency); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, ok := plans[planID]; !ok {
			plans[planID] = &gin.H{
				"id":          planID,
				"name":        name,
				"description": description,
				"pricing":     []pricingResp{},
			}
		}

		if pricingID != nil {
			existing := (*plans[planID])["pricing"].([]pricingResp)
			existing = append(existing, pricingResp{
				ID:             *pricingID,
				DurationMonths: *durationMonths,
				Price:          *price,
				Currency:       *currency,
			})
			(*plans[planID])["pricing"] = existing
		}
	}

	result := make([]gin.H, 0, len(plans))
	for _, p := range plans {
		result = append(result, *p)
	}

	c.JSON(http.StatusOK, gin.H{"plans": result})
}
