package handler

import (
	"app/repository"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetTrafficMonitoringChart returns 5-minute bucketed traffic grouped per merchant, route, and payment method
// Query params: start, end (RFC3339), client_uid, app_id, merchant_name, payment_method, route
func GetTrafficMonitoringChart(c *fiber.Ctx) error {
	ctx := c.Context()

	startStr := c.Query("start")
	endStr := c.Query("end")
	clientUID := c.Query("client_uid")
	appID := c.Query("app_id")
	merchant := c.Query("merchant_name")
	paymentMethod := c.Query("payment_method")
	route := c.Query("route")

	var (
		start time.Time
		end   time.Time
		err   error
	)

	if startStr == "" || endStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "start and end are required in RFC3339"})
	}

	start, err = parseTimeParam(startStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start format, use RFC3339 (e.g. 2025-10-13T11:00:00+07:00)"})
	}
	end, err = parseTimeParam(endStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end format, use RFC3339 (e.g. 2025-10-13T12:00:00+07:00)"})
	}

	data, err := repository.GetTrafficMonitoring(ctx, start, end, clientUID, appID, merchant, paymentMethod, route)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": data})
}

// GetTrafficSummary returns overall counts for given filters
func GetTrafficSummary(c *fiber.Ctx) error {
	ctx := c.Context()

	startStr := c.Query("start")
	endStr := c.Query("end")
	clientUID := c.Query("client_uid")
	appID := c.Query("app_id")
	merchant := c.Query("merchant_name")
	paymentMethod := c.Query("payment_method")
	route := c.Query("route")

	var (
		start time.Time
		end   time.Time
		err   error
	)

	if startStr == "" || endStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "start and end are required in RFC3339"})
	}

	start, err = parseTimeParam(startStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start format, use RFC3339 (e.g. 2025-10-13T11:00:00+07:00)"})
	}
	end, err = parseTimeParam(endStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end format, use RFC3339 (e.g. 2025-10-13T12:00:00+07:00)"})
	}

	sum, err := repository.GetTrafficSummary(ctx, start, end, clientUID, appID, merchant, paymentMethod, route)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(sum)
}

// parseTimeParam tries multiple layouts and also fixes space-before-offset cases (e.g., due to '+' not URL-encoded)
func parseTimeParam(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, fiber.ErrBadRequest
	}
	// If '+' wasn't encoded, in query it may become space. Normalize it.
	// Only replace the last space before timezone offset pattern if any; simple approach: replace all spaces with '+'
	normalized := strings.ReplaceAll(v, " ", "+")

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, normalized); err == nil {
			return t, nil
		}
	}

	// Try without timezone using WIB as default
	if t, err := time.Parse("2006-01-02T15:04:00", normalized); err == nil {
		if loc, e := time.LoadLocation("Asia/Jakarta"); e == nil {
			// Interpret parsed wall clock as WIB
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
		}
		return t, nil
	}

	return time.Time{}, fiber.ErrBadRequest
}
