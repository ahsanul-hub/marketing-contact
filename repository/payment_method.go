package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"app/database"
	"app/dto/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindPaymentMethodBySlug(slug string, defaultValue interface{}) (*model.PaymentMethod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection("dcb", "settings")

	var paymentMethod model.PaymentMethod

	filter := bson.M{"slug": slug}

	err := collection.FindOne(ctx, filter).Decode(&paymentMethod)
	if err != nil {
		return nil, err
	}
	// log.Printf("paymentMethod: %+v\n", paymentMethod)

	return &paymentMethod, nil
}

func GetPrice(prefix string, amount float32) (float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection("dcb", "settings")

	// Build slug
	slug := fmt.Sprintf("%s_charging", prefix)

	// Query MongoDB
	var paymentMethod model.PaymentMethod
	filter := bson.M{"slug": slug}
	err := collection.FindOne(ctx, filter).Decode(&paymentMethod)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, fmt.Errorf("charging settings not found for prefix: %s", prefix)
		}
		return 0, err
	}

	// Validate if Denom exists
	if len(paymentMethod.Denom) == 0 {
		return 0, fmt.Errorf("no denominated values available for prefix: %s", prefix)
	}

	// Loop through Denom to find the price for the given amount
	for denom, price := range paymentMethod.Denom {
		// Convert denom (key) to float64 and compare
		denomFloat := convertDenomToFloat(fmt.Sprintf("%d", denom))
		if float32(denomFloat) == amount {
			priceFloat, err := strconv.ParseFloat(price, 32)
			if err != nil {
				return 0, err
			}
			return float32(priceFloat), nil
		}
	}

	// Return error if no matching denom is found
	return 0, fmt.Errorf("amount %.2f not found in denominated values for prefix: %s", amount, prefix)
}

func GetChannelRouteWeight(clientID, paymentMethod string) ([]model.ChannelRouteWeight, error) {
	var weights []model.ChannelRouteWeight
	err := database.DB.
		Where("client_id = ? AND payment_method = ?", clientID, paymentMethod).
		Find(&weights).Error
	return weights, err
}

func InsertSupplierRouteWeight(route *model.ChannelRouteWeight) error {
	return database.DB.Create(route).Error
}

func GetTotalWeightByClientAndSlug(clientID, paymentSlug string) (int, error) {
	var total int
	err := database.DB.Model(&model.ChannelRouteWeight{}).
		Where("client_id = ? AND payment_method = ?", clientID, paymentSlug).
		Select("COALESCE(SUM(weight), 0)").Scan(&total).Error
	return total, err
}

func convertDenomToFloat(denom string) float32 {
	var denomFloat float32
	fmt.Sscanf(denom, "%f", &denomFloat)
	return denomFloat
}

// Route fee CRUD
func CreateRouteFee(fee *model.PaymentMethodRouteFee) error {
	return database.DB.Create(fee).Error
}

func UpdateRouteFee(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&model.PaymentMethodRouteFee{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteRouteFee(id uint) error {
	return database.DB.Where("id = ?", id).Delete(&model.PaymentMethodRouteFee{}).Error
}

func ListRouteFees(slug string) ([]model.PaymentMethodRouteFee, error) {
	var fees []model.PaymentMethodRouteFee
	q := database.DB.Model(&model.PaymentMethodRouteFee{})
	if slug != "" {
		q = q.Where("payment_method_slug = ?", slug)
	}
	if err := q.Find(&fees).Error; err != nil {
		return nil, err
	}
	return fees, nil
}
