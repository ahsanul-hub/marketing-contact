package handler

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"app/dto/model"
	"app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

func GetTransactionSummary(c *fiber.Ctx) error {
	layout := "2006-01-02"
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	start, _ := time.Parse(layout, startStr)
	end, _ := time.Parse(layout, endStr)

	merchant := c.Query("merchant")
	status := c.Query("status")
	payment := c.Query("payment_method")
	route := c.Query("route")
	format := c.Query("format")

	data, err := repository.GetTransactionSummaryDaily(start, end, merchant, status, payment, route)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	switch strings.ToLower(format) {
	case "csv":
		return exportCSVSummaryDaily(c, data)
	case "excel":
		return exportExcelSummaryDaily(c, data)
	default:
		return c.JSON(fiber.Map{
			"success": true,
			"data":    data,
		})
	}
}

func exportExcelSummaryDaily(c *fiber.Ctx, data []model.TransactionDailySummary) error {
	f := excelize.NewFile()
	sheet := "Sheet1"

	headers := []string{
		"Date", "Status", "PaymentMethod", "Amount", "Route", "MerchantName", "Total", "Revenue",
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for row, d := range data {
		values := []interface{}{
			d.Date,
			d.Status,
			d.PaymentMethod,
			d.Amount,
			d.Route,
			d.MerchantName,
			d.Total,
			d.Revenue,
			// d.FirstCreatedAt.Format(time.RFC3339),
			// d.LastCreatedAt.Format(time.RFC3339),
		}

		for col, val := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=transaction_summary.xlsx")

	return f.Write(c.Context().Response.BodyWriter())
}

func exportCSVSummaryDaily(c *fiber.Ctx, data []model.TransactionDailySummary) error {
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment;filename=transaction_summary.csv")

	writer := csv.NewWriter(c.Context().Response.BodyWriter())

	// Header
	writer.Write([]string{
		"Date", "Status", "PaymentMethod", "Amount", "Route", "MerchantName", "Total", "Revenue",
	})

	// Data
	for _, d := range data {
		writer.Write([]string{
			d.Date,
			d.Status,
			d.PaymentMethod,
			fmt.Sprintf("%d", d.Amount),
			d.Route,
			d.MerchantName,
			fmt.Sprintf("%d", d.Total),
			fmt.Sprintf("%.2f", d.Revenue),
		})
	}

	writer.Flush()
	return nil
}
