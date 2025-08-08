package service

import (
	"app/dto/model"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateExcelReport(transactions []model.Transactions, merchantName string) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Sheet1"
	index, _ := f.NewSheet(sheetName)

	// Tulis header
	headers := []string{"ID", "Merchant Transaction ID", "Date", "MDN", "Merchant", "App", "Amount", "Price", "Item", "Method", "Status"}
	for i, header := range headers {
		cell := getColumnName(i+1) + "1"
		f.SetCellValue(sheetName, cell, header)
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Tulis data transaksi
	for rowIndex, transaction := range transactions {
		var status, paymentMethod string
		var price uint

		switch transaction.StatusCode {
		case 1005:
			status = "failed"
		case 1001:
			status = "pending"
		case 1003:
			status = "pending"
		case 1000:
			status = "success"
		}

		switch transaction.PaymentMethod {
		case "qris":
			price = transaction.Amount

			paymentMethod = transaction.PaymentMethod
		case "dana":
			price = transaction.Amount
			paymentMethod = transaction.PaymentMethod
		case "telkomsel_airtime":
			paymentMethod = "Telkomsel"
			price = transaction.Price
		case "xl_airtime":
			paymentMethod = "XL"
			price = transaction.Price
		case "indosat_airtime":
			paymentMethod = "Indosat"
			price = transaction.Price
		case "three_airtime":
			paymentMethod = "Tri"
			price = transaction.Price
		case "smartfren_airtime":
			paymentMethod = "Smartfren"
			price = transaction.Price
		default:
			price = transaction.Price
			paymentMethod = transaction.PaymentMethod
		}

		var createdAt string
		if transaction.AppName == "Zingplay games" {
			createdAt = transaction.CreatedAt.In(loc).Format("01/02/2006 15:04:05")
		} else {
			createdAt = transaction.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		}

		row := rowIndex + 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), transaction.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), transaction.MtTid)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), createdAt)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), transaction.UserMDN)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), transaction.MerchantName)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), transaction.AppName)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), transaction.Amount)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), price)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(row), transaction.ItemName)
		f.SetCellValue(sheetName, "J"+strconv.Itoa(row), paymentMethod)
		f.SetCellValue(sheetName, "K"+strconv.Itoa(row), status)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Kembalikan sebagai bytes
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Fungsi untuk mendapatkan nama kolom berdasarkan indeks
func getColumnName(index int) string {
	columnName := ""
	for index > 0 {
		index-- // Mengurangi 1 untuk mengubah indeks ke 0-based
		columnName = string(rune('A'+(index%26))) + columnName
		index /= 26
	}
	return columnName
}
