package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"app/database"
	"app/dto/model"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
	"gorm.io/gorm"
)

type PaymentMethodRepository struct {
	DB *gorm.DB
}

var merchantCache *cache.Cache

func init() {
	merchantCache = cache.New(30*time.Minute, 35*time.Minute)
}

func FindClient(ctx context.Context, clientAppKey, clientID string) (*model.Client, error) {
	span, _ := apm.StartSpan(ctx, "FindClient", "repository")
	defer span.End()
	db := database.DB

	cacheKey := fmt.Sprintf("client:%s:%s", clientAppKey, clientID)
	if cachedClient, found := merchantCache.Get(cacheKey); found {
		return cachedClient.(*model.Client), nil
	}

	var client model.Client
	// Mencari client berdasarkan clientAppKey dan clientID
	if err := db.Joins("JOIN client_apps ON client_apps.client_id = clients.uid").
		Where("client_apps.app_id = ? AND client_apps.app_key = ?", clientID, clientAppKey).
		Preload("ClientApps").Preload("PaymentMethods").
		First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching client: %w", err)
	}

	merchantCache.Set(cacheKey, &client, cache.DefaultExpiration)

	return &client, nil
}

const (
	clientSecretLength = 15
)

func generateUniqueKey() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func generateClientSecret() (string, error) {

	secretBytes := make([]byte, clientSecretLength)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	encodedSecret := base64.RawURLEncoding.EncodeToString(secretBytes)
	if len(encodedSecret) > clientSecretLength {
		encodedSecret = encodedSecret[:clientSecretLength]
	}
	return encodedSecret, nil
}

func AddMerchant(ctx context.Context, input *model.InputClientRequest) error {

	clientSecret, err := generateClientSecret()
	if err != nil {
		return fmt.Errorf("failed to generate client secret: %w", err)
	}

	clientAppKey, err := generateUniqueKey()
	if err != nil {
		return fmt.Errorf("failed to generate client app key: %w", err)
	}

	clientAppID, err := generateUniqueKey()
	if err != nil {
		return fmt.Errorf("failed to generate client app ID: %w", err)
	}

	uuidClient, err := uuid.NewV7()
	if err != nil {
		log.Fatal(err)
	}
	client := model.Client{
		UID:          uuidClient.String(),
		ClientName:   *input.ClientName,
		ClientAppkey: clientAppKey,
		ClientSecret: clientSecret,
		ClientID:     clientAppID,
		AppName:      *input.AppName,
		Mobile:       *input.Mobile,
		ClientStatus: *input.ClientStatus,
		Testing:      *input.Testing,
		Lang:         *input.Lang,
		CallbackURL:  *input.CallbackURL,
		FailCallback: *input.FailCallback,
		Isdcb:        *input.Isdcb,
	}

	if err := database.DB.Create(&client).Error; err != nil {

		return fmt.Errorf("unable to create client: %w", err)
	}

	for _, pm := range input.PaymentMethods {
		if err := AddPaymentMethod(client.UID, &pm); err != nil {
			log.Printf("Failed to add payment method for client %s: %s", client.UID, err)
			return err
		}
	}
	// log.Println("settlement: ", input.Settlements)

	for _, settlements := range input.Settlements {
		if err := AddSettlements(client.UID, &settlements); err != nil {
			log.Printf("Failed to add settlement for client %s: %s", client.UID, err)
			return err
		}
	}

	for _, clients := range input.ClientApp {
		if err := AddClientApps(client.UID, &clients); err != nil {
			log.Printf("Failed to add client app for client %s: %s", client.UID, err)
			return err
		}
	}

	return nil
}

func UpdateMerchant(ctx context.Context, clientID string, input *model.InputClientRequest) error {
	db := database.DB

	var existingClient model.Client
	if err := db.Where("client_id = ?", clientID).First(&existingClient).Error; err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	cacheKey := fmt.Sprintf("client:%s:%s", existingClient.ClientAppkey, existingClient.ClientID)
	merchantCache.Delete(cacheKey)

	updateData := map[string]interface{}{}

	if input.ClientName != nil {
		updateData["client_name"] = *input.ClientName
	}
	if input.AppName != nil {
		updateData["app_name"] = *input.AppName
	}
	if input.Mobile != nil {
		updateData["mobile"] = *input.Mobile
	}
	if input.ClientStatus != nil {
		updateData["client_status"] = *input.ClientStatus
	}
	if input.Testing != nil {
		updateData["testing"] = *input.Testing
	}
	if input.Lang != nil {
		updateData["lang"] = *input.Lang
	}
	if input.CallbackURL != nil {
		updateData["callback_url"] = *input.CallbackURL
	}
	if input.FailCallback != nil {
		updateData["fail_callback"] = *input.FailCallback
	}
	if input.Isdcb != nil {
		updateData["isdcb"] = *input.Isdcb
	}

	if len(updateData) > 0 {
		if err := db.Model(&existingClient).Updates(updateData).Error; err != nil {
			return fmt.Errorf("unable to update client: %w", err)
		}
	}

	for _, pm := range input.PaymentMethods {
		var existingPM model.PaymentMethodClient

		// Check if the payment method exists
		if err := db.Where("client_id = ? AND name = ?", existingClient.UID, pm.Name).First(&existingPM).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new payment method if it doesn't exist
				pm.ClientID = existingClient.UID
				if err := AddPaymentMethod(existingClient.UID, &pm); err != nil {
					log.Printf("Failed to add payment method for client %s: %s", existingClient.UID, err)
					return err
				}
			} else {
				log.Printf("Failed to check existing payment method: %s", err)
				return err
			}
		} else {
			// Update existing payment method only if properties are provided
			if pm.Name != "" {
				existingPM.Name = pm.Name
			}
			if pm.Route != nil {
				existingPM.Route = pm.Route
			}
			if pm.Status != 0 {
				existingPM.Status = pm.Status
			}
			if pm.Msisdn != 0 {
				existingPM.Msisdn = pm.Msisdn
			}

			// Save the updated payment method
			if err := db.Save(&existingPM).Error; err != nil {
				log.Printf("Failed to update payment method for client %s: %s", existingClient.UID, err)
				return err
			}
		}
	}

	for _, app := range input.ClientApp {
		var existingApps model.ClientApp

		// Check if the payment method exists
		if err := db.Where("client_id = ? AND app_id = ?", existingClient.UID, app.AppID).First(&existingApps).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new payment method if it doesn't exist
				app.ClientID = existingClient.UID
				if err := AddClientApps(existingClient.UID, &app); err != nil {
					log.Printf("Failed to add client app for client %s: %s", existingClient.UID, err)
					return err
				}
			} else {
				log.Printf("Failed to check existing payment method: %s", err)
				return err
			}
		} else {
			// Update existing payment method only if properties are provided
			if app.AppName != "" {
				existingApps.AppName = app.AppName
			}
			if app.CallbackURL != "" {
				existingApps.CallbackURL = app.CallbackURL
			}
			if app.Testing != 0 {
				existingApps.Testing = app.Testing
			}
			if app.Status != 0 {
				existingApps.Status = app.Status
			}
			if app.FailCallback != "" {
				existingApps.FailCallback = app.FailCallback
			}
			if app.Mobile != "" {
				existingApps.Mobile = app.Mobile
			}

			// Save the updated payment method
			if err := db.Save(&existingApps).Error; err != nil {
				log.Printf("Failed to update app for client %s: %s", existingClient.UID, err)
				return err
			}
		}
	}

	// Update Settlements
	for _, settlement := range input.Settlements {
		var extSettlement model.SettlementClient

		if err := db.Where("client_id = ? AND name = ?", existingClient.UID, settlement.Name).First(&extSettlement).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				settlement.ClientID = existingClient.UID
				if err := AddSettlements(existingClient.UID, &settlement); err != nil {
					log.Printf("Failed to add settlement for client %s: %s", existingClient.UID, err)
					return err
				}
			} else {
				log.Printf("Failed to check existing settlement: %s", err)
				return err
			}
		} else {
			if settlement.IsBhpuso != "" {
				extSettlement.IsBhpuso = settlement.IsBhpuso
			}
			if settlement.ServiceCharge != nil {
				extSettlement.ServiceCharge = settlement.ServiceCharge
			}
			if settlement.Tax23 != nil {
				extSettlement.Tax23 = settlement.Tax23
			}
			if settlement.Ppn != nil {
				extSettlement.Ppn = settlement.Ppn
			}
			if settlement.Mdr != "" {
				extSettlement.Mdr = settlement.Mdr
			}
			if settlement.MdrType != "" {
				extSettlement.MdrType = settlement.MdrType
			}
			if settlement.AdditionalFee != nil {
				extSettlement.AdditionalFee = settlement.AdditionalFee
			}
			if settlement.AdditionalFeeType != nil {
				extSettlement.AdditionalFeeType = settlement.AdditionalFeeType
			}
			if settlement.PaymentType != "" {
				extSettlement.PaymentType = settlement.PaymentType
			}
			if settlement.ShareRedision != nil {
				extSettlement.ShareRedision = settlement.ShareRedision
			}
			if settlement.AdditionalPercent != nil {
				extSettlement.AdditionalPercent = settlement.AdditionalPercent
			}
			if settlement.SharePartner != nil {
				extSettlement.SharePartner = settlement.SharePartner
			}
			if settlement.IsDivide1Poin1 != "" {
				extSettlement.IsDivide1Poin1 = settlement.IsDivide1Poin1
			}

			if err := db.Save(&extSettlement).Error; err != nil {
				return fmt.Errorf("failed to update settlement for client %s: %w", existingClient.UID, err)
			}
		}
	}

	merchantCache.Set(cacheKey, &existingClient, cache.DefaultExpiration)

	return nil
}

func AddPaymentMethod(clientID string, paymentMethod *model.PaymentMethodClient) error {
	paymentMethod.ClientID = clientID

	err := database.DB.Create(paymentMethod).Error
	if err != nil {
		return fmt.Errorf("failed to add payment method: %w", err)
	}

	return nil
}

func AddSettlements(clientID string, settlements *model.SettlementClient) error {
	settlements.ClientID = clientID

	// log.Println(settlements)

	err := database.DB.Create(settlements).Error
	if err != nil {
		return fmt.Errorf("failed to add  method: %w", err)
	}

	return nil
}

func AddClientApps(clientID string, clients *model.ClientApp) error {

	clientAppID, err := generateUniqueKey()
	if err != nil {
		return fmt.Errorf("failed to generate client app ID: %w", err)
	}

	clientAppKey, err := generateUniqueKey()
	if err != nil {
		return fmt.Errorf("failed to generate client app ID: %w", err)
	}

	clients.ClientID = clientID
	clients.AppID = clientAppID
	clients.AppKey = clientAppKey

	// log.Println(settlements)

	err = database.DB.Create(clients).Error
	if err != nil {
		return fmt.Errorf("failed to add  method: %w", err)
	}

	return nil
}

func GetByClientID(clientAppID string) (model.Client, error) {
	var client model.Client
	if err := database.DB.Preload("PaymentMethods").Preload("Settlements").Preload("ClientApps").Where("id = ?", clientAppID).First(&client).Error; err != nil {
		return client, fmt.Errorf("client not found: %w", err)
	}
	return client, nil
}

func GetAllClients() ([]model.Client, error) {
	var clients []model.Client
	if err := database.DB.Preload("PaymentMethods").Preload("Settlements").Preload("ClientApps").Find(&clients).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch clients: %w", err)
	}

	return clients, nil
}

func DeleteMerchant(clientID string) error {
	db := database.DB

	var existingClient model.Client
	if err := db.Where("client_id = ?", clientID).First(&existingClient).Error; err != nil {
		return fmt.Errorf("client not found: %w", err)
	}

	cacheKey := fmt.Sprintf("client:%s:%s", existingClient.ClientAppkey, existingClient.ClientID)
	merchantCache.Delete(cacheKey)

	if err := db.Where("client_id = ?", existingClient.UID).Delete(&model.PaymentMethodClient{}).Error; err != nil {
		log.Printf("Failed to delete payment methods for client %s: %s", existingClient.UID, err)
		return err
	}

	if err := db.Where("client_id = ?", existingClient.UID).Delete(&model.SettlementClient{}).Error; err != nil {
		log.Printf("Failed to delete settlements for client %s: %s", existingClient.UID, err)
		return err
	}

	if err := db.Where("client_id = ?", existingClient.UID).Delete(&model.ClientApp{}).Error; err != nil {
		log.Printf("Failed to delete app for client %s: %s", existingClient.UID, err)
		return err
	}

	if err := db.Delete(&existingClient).Error; err != nil {
		return fmt.Errorf("unable to delete client: %w", err)
	}

	return nil
}

func GetSettlementConfig(clientID string) ([]model.SettlementClient, error) {
	cacheKey := fmt.Sprintf("settlement:%s", clientID)
	if cachedConfig, found := merchantCache.Get(cacheKey); found {
		return cachedConfig.([]model.SettlementClient), nil
	}

	var config []model.SettlementClient

	if err := database.DB.Where("client_id = ?", clientID).Find(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch settlement config: %w", err)
	}

	merchantCache.Set(cacheKey, config, cache.DefaultExpiration)

	return config, nil
}

func NewPaymentMethodRepository(db *gorm.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{DB: db}
}

func (r *PaymentMethodRepository) Create(paymentMethod *model.PaymentMethod) error {
	return r.DB.Create(paymentMethod).Error
}

func (r *PaymentMethodRepository) GetAll() ([]model.PaymentMethod, error) {
	var paymentMethods []model.PaymentMethod
	err := r.DB.Find(&paymentMethods).Error
	return paymentMethods, err
}

func (r *PaymentMethodRepository) GetBySlug(slug string) (*model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	err := r.DB.Where("slug = ?", slug).First(&paymentMethod).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *PaymentMethodRepository) Update(paymentMethod *model.PaymentMethod) error {
	return r.DB.Save(paymentMethod).Error
}

func (r *PaymentMethodRepository) Delete(slug string) error {
	var paymentMethod model.PaymentMethod

	if err := r.DB.Where("slug = ?", slug).First(&paymentMethod).Error; err != nil {
		return err
	}

	return r.DB.Delete(&paymentMethod).Error
}
