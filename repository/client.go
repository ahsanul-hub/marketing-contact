package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
	merchantCache = cache.New(15*time.Minute, 18*time.Minute)
}

func FindClient(ctx context.Context, clientAppKey, clientID string) (*model.Client, error) {
	span, _ := apm.StartSpan(ctx, "FindClient", "repository")
	defer span.End()
	db := database.DB

	cacheKey := fmt.Sprintf("client:%s:%s", clientAppKey, clientID)
	if cachedClient, found := merchantCache.Get(cacheKey); found {
		// log.Println("data diambil dari cache")
		return cachedClient.(*model.Client), nil
	}

	var client model.Client
	// Mencari client berdasarkan clientAppKey dan clientID
	if err := db.Joins("JOIN client_apps ON client_apps.client_id = clients.uid").
		Where("client_apps.app_id = ? AND client_apps.app_key = ?", clientID, clientAppKey).
		Preload("ClientApps").Preload("PaymentMethods").Preload("ChannelRouteWeight").
		First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching client: %w", err)
	}

	merchantCache.Set(cacheKey, &client, cache.DefaultExpiration)
	// log.Println("data diambil dari database")

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

	if input.ClientName == nil || input.AppName == nil || input.Mobile == nil ||
		input.ClientStatus == nil || input.Testing == nil || input.Lang == nil ||
		input.Phone == nil || input.Email == nil || input.CallbackURL == nil || input.FailCallback == nil || input.Isdcb == nil {
		log.Println("Error: Missing required fields in input")
		return fmt.Errorf("missing required fields in input")
	}

	clientSecret, err := generateClientSecret()
	if err != nil {
		log.Println("Error generating client secret:", err)
		return fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Generate keys & UUID
	clientAppKey, err := generateUniqueKey()
	if err != nil {
		log.Println("Error generating client app key:", err)
		return fmt.Errorf("failed to generate client app key: %w", err)
	}

	clientAppID, err := generateUniqueKey()
	if err != nil {
		log.Println("Error generating client app ID:", err)
		return fmt.Errorf("failed to generate client app ID: %w", err)
	}

	uuidClient, err := uuid.NewV7()
	if err != nil {
		log.Println("Error generating UUID:", err)
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	client := model.Client{
		UID:          uuidClient.String(),
		ClientName:   *input.ClientName,
		ClientAppkey: clientAppKey,
		ClientSecret: clientSecret,
		ClientID:     clientAppID,
		AppName:      *input.AppName,
		Address:      *input.Address,
		Mobile:       *input.Mobile,
		ClientStatus: *input.ClientStatus,
		Testing:      *input.Testing,
		Lang:         *input.Lang,
		Phone:        *input.Phone,
		Email:        *input.Email,
		CallbackURL:  *input.CallbackURL,
		FailCallback: *input.FailCallback,
		Isdcb:        *input.Isdcb,
	}

	if err := database.DB.Create(&client).Error; err != nil {
		log.Println("Error creating client in database:", err)
		return fmt.Errorf("unable to create client: %w", err)
	}

	for _, pm := range input.PaymentMethods {
		if err := AddPaymentMethod(client.UID, &pm); err != nil {
			log.Printf("Failed to add payment method for client %s: %s", client.UID, err)
			return err
		}
	}

	// Process Settlements
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

	for _, weight := range input.ChannelRouteWeight {
		weight.ClientID = client.UID
		if err := database.DB.Create(&weight).Error; err != nil {
			log.Printf("Failed to insert supplier route weight: %+v, error: %v", weight, err)
			return fmt.Errorf("failed to create supplier route weight: %w", err)
		}
	}

	// Add new client to cache for immediate availability
	var newClient model.Client
	if err := database.DB.Where("client_id = ?", client.ClientID).
		Preload("ClientApps").
		Preload("PaymentMethods").
		Preload("Settlements").
		Preload("ChannelRouteWeight").
		First(&newClient).Error; err != nil {
		log.Printf("Failed to load new client for cache: %s", err)
		// Don't return error here, just log it as cache update is not critical
	} else {
		// Cache with all possible app_key combinations
		for _, app := range newClient.ClientApps {
			cacheKey := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Set(cacheKey, &newClient, cache.DefaultExpiration)
		}
		log.Printf("Cache added for new client: %s", client.ClientID)
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
	if input.Phone != nil {
		updateData["phone"] = *input.Phone
	}
	if input.Email != nil {
		updateData["email"] = *input.Email
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
			// Update status and msisdn if provided (including 0 values)
			existingPM.Status = pm.Status
			existingPM.Msisdn = pm.Msisdn

			// Save the updated payment method
			if err := db.Save(&existingPM).Error; err != nil {
				log.Printf("Failed to update payment method for client %s: %s", existingClient.UID, err)
				return err
			}
		}
	}

	for _, app := range input.ClientApp {
		var existingApps model.ClientApp

		// Check if app has ID (for existing apps) or find by app_name (for new apps)
		var query *gorm.DB
		if app.AppID != "" {
			// Update existing app by app_id
			query = db.Where("client_id = ? AND app_id = ?", existingClient.UID, app.AppID)
		} else {
			// Find by app_name for new or existing apps
			query = db.Where("client_id = ? AND app_name = ?", existingClient.UID, app.AppName)
		}

		if err := query.First(&existingApps).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new client app if it doesn't exist
				app.ClientID = existingClient.UID
				// Clear ID to let database auto-generate
				app.ID = 0
				if err := AddClientApps(existingClient.UID, &app); err != nil {
					log.Printf("Failed to add client app for client %s: %s", existingClient.UID, err)
					return err
				}
			} else {
				log.Printf("Failed to check existing client app: %s", err)
				return err
			}
		} else {
			// Update existing app only if properties are provided
			if app.AppName != "" {
				existingApps.AppName = app.AppName
			}
			if app.CallbackURL != "" {
				existingApps.CallbackURL = app.CallbackURL
			}
			// Update testing and status if provided (including 0 values)
			existingApps.Testing = app.Testing
			existingApps.Status = app.Status
			if app.FailCallback != "" {
				existingApps.FailCallback = app.FailCallback
			}
			if app.Mobile != "" {
				existingApps.Mobile = app.Mobile
			}

			// Save the updated client app
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

	for _, weight := range input.ChannelRouteWeight {
		var existingWeight model.ChannelRouteWeight
		if err := db.Where("client_id = ? AND payment_method = ? AND route = ?", existingClient.UID, weight.PaymentMethod, weight.Route).First(&existingWeight).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				weight.ClientID = existingClient.UID
				if err := db.Create(&weight).Error; err != nil {
					log.Printf("Failed to create supplier weight: %s", err)
					return err
				}
			} else {
				log.Printf("Failed to find supplier weight: %s", err)
				return err
			}
		} else {
			existingWeight.Weight = weight.Weight
			if err := db.Save(&existingWeight).Error; err != nil {
				log.Printf("Failed to update supplier weight: %s", err)
				return err
			}
		}
	}

	// Refresh cache with updated client data including all related data
	var updatedClient model.Client
	if err := db.Where("client_id = ?", clientID).
		Preload("ClientApps").
		Preload("PaymentMethods").
		Preload("Settlements").
		Preload("ChannelRouteWeight").
		First(&updatedClient).Error; err != nil {
		log.Printf("Failed to reload updated client for cache: %s", err)
		// Don't return error here, just log it as cache refresh is not critical
	} else {
		// Clear all cache entries for this client and set new cache entries for all app combos
		ClearClientCacheByClientUID(updatedClient.UID)
		for _, app := range updatedClient.ClientApps {
			newCacheKey := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Set(newCacheKey, &updatedClient, cache.DefaultExpiration)
		}
		log.Printf("Cache refreshed for client: %s", clientID)
	}

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
		return fmt.Errorf("failed to generate client app key: %w", err)
	}

	clients.ClientID = clientID
	clients.AppID = clientAppID
	clients.AppKey = clientAppKey
	// Ensure ID is 0 for auto-generation
	clients.ID = 0

	err = database.DB.Create(clients).Error
	if err != nil {
		return fmt.Errorf("failed to add client app: %w", err)
	}

	return nil
}

func GetByClientID(clientAppID string) (model.Client, error) {
	var client model.Client
	if err := database.DB.Preload("PaymentMethods").Preload("Settlements").Preload("ClientApps").Preload("ChannelRouteWeight").Where("client_id = ?", clientAppID).First(&client).Error; err != nil {
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

	if err := db.Where("client_id = ?", existingClient.UID).Delete(&model.ChannelRouteWeight{}).Error; err != nil {
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

func ConvertSelectedPaymentMethods(clientID string, selectedMethods []model.SelectedPaymentMethod) ([]model.PaymentMethodClient, []model.SettlementClient, []model.ChannelRouteWeight, error) {
	var paymentMethods []model.PaymentMethodClient
	var settlements []model.SettlementClient
	var channelRouteWeights []model.ChannelRouteWeight

	repo := PaymentMethodRepository{DB: database.DB}

	for _, selected := range selectedMethods {
		// Get payment method by slug to get the available routes
		paymentMethod, err := repo.GetBySlug(selected.PaymentMethodSlug)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("payment method with slug %s not found: %w", selected.PaymentMethodSlug, err)
		}

		// Validate selected routes against available routes
		availableRoutes := make(map[string]bool)
		log.Printf("Payment method %s has routes: %+v", selected.PaymentMethodSlug, paymentMethod.Route)
		// Convert pq.StringArray to []string for iteration
		for _, route := range []string(paymentMethod.Route) {
			availableRoutes[route] = true
		}

		// Extract route names and validate + create route weights
		totalWeight := 0
		firstRoute := ""

		// Only create ChannelRouteWeight if there are multiple routes
		if len(selected.SelectedRoutes) > 1 {
			for _, routeWeight := range selected.SelectedRoutes {
				if !availableRoutes[routeWeight.Route] {
					return nil, nil, nil, fmt.Errorf("route %s is not available for payment method %s", routeWeight.Route, selected.PaymentMethodSlug)
				}

				// Take only the first route for PaymentMethodClient
				if firstRoute == "" {
					firstRoute = routeWeight.Route
				}

				// Lookup default fee per route
				var routeFee model.PaymentMethodRouteFee
				_ = database.DB.Where("payment_method_slug = ? AND route = ?", selected.PaymentMethodSlug, routeWeight.Route).First(&routeFee).Error

				channelWeight := model.ChannelRouteWeight{
					ClientID:      clientID,
					PaymentMethod: selected.PaymentMethodSlug,
					Route:         routeWeight.Route,
					Weight:        routeWeight.Weight,
					Fee:           routeFee.Fee,
				}
				channelRouteWeights = append(channelRouteWeights, channelWeight)
				totalWeight += routeWeight.Weight
			}

			// Validate total weight (should be 100 for percentage-based systems)
			if totalWeight != 100 {
				return nil, nil, nil, fmt.Errorf("total weight for payment method %s must equal 100, got %d", selected.PaymentMethodSlug, totalWeight)
			}
		} else {
			// Single route - no weight validation needed
			if !availableRoutes[selected.SelectedRoutes[0].Route] {
				return nil, nil, nil, fmt.Errorf("route %s is not available for payment method %s", selected.SelectedRoutes[0].Route, selected.PaymentMethodSlug)
			}
			firstRoute = selected.SelectedRoutes[0].Route

			// Lookup default fee for single route (for downstream usage)
			var routeFee model.PaymentMethodRouteFee
			_ = database.DB.Where("payment_method_slug = ? AND route = ?", selected.PaymentMethodSlug, firstRoute).First(&routeFee).Error
		}

		// Create route object with denom array for the first route
		routeObject := map[string][]string{
			firstRoute: paymentMethod.Denom,
		}

		routeJSON, err := json.Marshal(routeObject)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to marshal route for payment method %s: %w", selected.PaymentMethodSlug, err)
		}

		paymentMethodClient := model.PaymentMethodClient{
			Name:     selected.PaymentMethodSlug,
			Route:    routeJSON,
			Flexible: paymentMethod.Flexible,
			Status:   selected.Status,
			Msisdn:   selected.Msisdn,
			ClientID: clientID,
		}

		paymentMethods = append(paymentMethods, paymentMethodClient)

		// Auto-generate settlement for this payment method
		settlement := model.SettlementClient{
			ClientID:    clientID,
			Name:        selected.PaymentMethodSlug, // Settlement name matches payment method slug
			PaymentType: paymentMethod.Type,         // Use payment method type
		}

		// Apply settlement config if provided
		if selected.SettlementConfig != nil {
			config := selected.SettlementConfig
			if config.IsBhpuso != "" {
				settlement.IsBhpuso = config.IsBhpuso
			}
			if config.ServiceCharge != nil {
				settlement.ServiceCharge = config.ServiceCharge
			}
			if config.Tax23 != nil {
				settlement.Tax23 = config.Tax23
			}
			if config.Ppn != nil {
				settlement.Ppn = config.Ppn
			}
			if config.Mdr != "" {
				settlement.Mdr = config.Mdr
			}
			if config.MdrType != "" {
				settlement.MdrType = config.MdrType
			}
			if config.AdditionalFee != nil {
				settlement.AdditionalFee = config.AdditionalFee
			}
			if config.AdditionalPercent != nil {
				settlement.AdditionalPercent = config.AdditionalPercent
			}
			if config.AdditionalFeeType != nil {
				settlement.AdditionalFeeType = config.AdditionalFeeType
			}
			if config.PaymentType != "" {
				settlement.PaymentType = config.PaymentType
			}
			if config.ShareRedision != nil {
				settlement.ShareRedision = config.ShareRedision
			}
			if config.SharePartner != nil {
				settlement.SharePartner = config.SharePartner
			}
			if config.IsDivide1Poin1 != "" {
				settlement.IsDivide1Poin1 = config.IsDivide1Poin1
			}
		} else {
			// Set default values if no config provided
			settlement.Mdr = "0"
			settlement.MdrType = "percent"
		}

		settlements = append(settlements, settlement)
	}

	return paymentMethods, settlements, channelRouteWeights, nil
}

func AddMerchantV2(ctx context.Context, input *model.InputClientRequestV2) error {
	if input.ClientName == nil || input.AppName == nil || input.Mobile == nil ||
		input.ClientStatus == nil || input.Testing == nil || input.Lang == nil ||
		input.Phone == nil || input.Email == nil || input.CallbackURL == nil || input.Isdcb == nil {
		log.Println("Error: Missing required fields in input")
		return fmt.Errorf("missing required fields in input")
	}

	clientSecret, err := generateClientSecret()
	if err != nil {
		log.Println("Error generating client secret:", err)
		return fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Generate keys & UUID
	clientAppKey, err := generateUniqueKey()
	if err != nil {
		log.Println("Error generating client app key:", err)
		return fmt.Errorf("failed to generate client app key: %w", err)
	}

	clientAppID, err := generateUniqueKey()
	if err != nil {
		log.Println("Error generating client app ID:", err)
		return fmt.Errorf("failed to generate client app ID: %w", err)
	}

	uuidClient, err := uuid.NewV7()
	if err != nil {
		log.Println("Error generating UUID:", err)
		return fmt.Errorf("failed to generate UUID: %w", err)
	}

	var failCallback string
	if input.FailCallback != nil {
		failCallback = *input.FailCallback
	}

	client := model.Client{
		UID:          uuidClient.String(),
		ClientName:   *input.ClientName,
		ClientAppkey: clientAppKey,
		ClientSecret: clientSecret,
		ClientID:     clientAppID,
		AppName:      *input.AppName,
		Address:      *input.Address,
		Mobile:       *input.Mobile,
		ClientStatus: *input.ClientStatus,
		Testing:      *input.Testing,
		Lang:         *input.Lang,
		Phone:        *input.Phone,
		Email:        *input.Email,
		CallbackURL:  *input.CallbackURL,
		FailCallback: failCallback,
		Isdcb:        *input.Isdcb,
	}

	if err := database.DB.Create(&client).Error; err != nil {
		log.Println("Error creating client in database:", err)
		return fmt.Errorf("unable to create client: %w", err)
	}

	// Convert selected payment methods to PaymentMethodClient
	paymentMethods, settlements, channelRouteWeights, err := ConvertSelectedPaymentMethods(client.UID, input.SelectedPaymentMethods)
	if err != nil {
		log.Printf("Failed to convert selected payment methods: %s", err)
		return err
	}

	// Validate consistency between payment methods and settlements
	if err := ValidatePaymentMethodSettlementConsistency(paymentMethods, settlements); err != nil {
		log.Printf("Payment method and settlement consistency validation failed: %s", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Add payment methods
	for _, pm := range paymentMethods {
		if err := AddPaymentMethod(client.UID, &pm); err != nil {
			log.Printf("Failed to add payment method for client %s: %s", client.UID, err)
			return err
		}
	}

	// Process Settlements
	for _, settlement := range settlements {
		if err := AddSettlements(client.UID, &settlement); err != nil {
			log.Printf("Failed to add settlement for client %s: %s", client.UID, err)
			return err
		}
	}

	for _, weight := range channelRouteWeights {
		if err := database.DB.Create(&weight).Error; err != nil {
			log.Printf("Failed to insert supplier route weight: %+v, error: %v", weight, err)
			return fmt.Errorf("failed to create supplier route weight: %w", err)
		}
	}

	for _, clients := range input.ClientApp {
		if err := AddClientApps(client.UID, &clients); err != nil {
			log.Printf("Failed to add client app for client %s: %s", client.UID, err)
			return err
		}
	}

	// Add new client to cache for immediate availability (override any stale entries)
	var newClient model.Client
	if err := database.DB.Where("client_id = ?", client.ClientID).
		Preload("ClientApps").
		Preload("PaymentMethods").
		Preload("Settlements").
		Preload("ChannelRouteWeight").
		First(&newClient).Error; err != nil {
		log.Printf("Failed to load new client for cache: %s", err)
	} else {
		// Hapus semua entry cache lama untuk client ini lalu isi ulang
		for _, app := range newClient.ClientApps {
			oldKey := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Delete(oldKey)
		}
		for _, app := range newClient.ClientApps {
			cacheKey := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Set(cacheKey, &newClient, cache.DefaultExpiration)
		}
		log.Printf("Cache added for new client: %s", client.ClientID)
	}

	return nil
}

func UpdateMerchantV2(ctx context.Context, clientID string, input *model.InputClientRequestV2) error {
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
	if input.Address != nil {
		updateData["address"] = *input.Address
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
	if input.Phone != nil {
		updateData["phone"] = *input.Phone
	}
	if input.Email != nil {
		updateData["email"] = *input.Email
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

	// Handle payment method updates with route weights
	if len(input.SelectedPaymentMethods) > 0 {
		// Convert selected payment methods
		paymentMethods, settlements, channelRouteWeights, err := ConvertSelectedPaymentMethods(existingClient.UID, input.SelectedPaymentMethods)
		if err != nil {
			log.Printf("Failed to convert selected payment methods: %s", err)
			return err
		}

		// Update payment methods
		for _, pm := range paymentMethods {
			var existingPM model.PaymentMethodClient

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
				// Update existing payment method
				existingPM.Route = pm.Route
				existingPM.Flexible = pm.Flexible
				existingPM.Status = pm.Status
				existingPM.Msisdn = pm.Msisdn

				if err := db.Save(&existingPM).Error; err != nil {
					log.Printf("Failed to update payment method for client %s: %s", existingClient.UID, err)
					return err
				}
			}
		}

		// Update settlements
		for _, settlement := range settlements {
			var existingSettlement model.SettlementClient

			if err := db.Where("client_id = ? AND name = ?", existingClient.UID, settlement.Name).First(&existingSettlement).Error; err != nil {
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
				// Update existing settlement
				if settlement.IsBhpuso != "" {
					existingSettlement.IsBhpuso = settlement.IsBhpuso
				}
				if settlement.ServiceCharge != nil {
					existingSettlement.ServiceCharge = settlement.ServiceCharge
				}
				if settlement.Tax23 != nil {
					existingSettlement.Tax23 = settlement.Tax23
				}
				if settlement.Ppn != nil {
					existingSettlement.Ppn = settlement.Ppn
				}
				if settlement.Mdr != "" {
					existingSettlement.Mdr = settlement.Mdr
				}
				if settlement.MdrType != "" {
					existingSettlement.MdrType = settlement.MdrType
				}
				if settlement.AdditionalFee != nil {
					existingSettlement.AdditionalFee = settlement.AdditionalFee
				}
				if settlement.AdditionalPercent != nil {
					existingSettlement.AdditionalPercent = settlement.AdditionalPercent
				}
				if settlement.AdditionalFeeType != nil {
					existingSettlement.AdditionalFeeType = settlement.AdditionalFeeType
				}
				if settlement.PaymentType != "" {
					existingSettlement.PaymentType = settlement.PaymentType
				}
				if settlement.ShareRedision != nil {
					existingSettlement.ShareRedision = settlement.ShareRedision
				}
				if settlement.SharePartner != nil {
					existingSettlement.SharePartner = settlement.SharePartner
				}
				if settlement.IsDivide1Poin1 != "" {
					existingSettlement.IsDivide1Poin1 = settlement.IsDivide1Poin1
				}

				if err := db.Save(&existingSettlement).Error; err != nil {
					return fmt.Errorf("failed to update settlement for client %s: %w", existingClient.UID, err)
				}
			}
		}

		// Update channel route weights - delete old ones and create new ones
		for _, weight := range channelRouteWeights {
			// Delete existing route weights for this payment method
			if err := db.Where("client_id = ? AND payment_method = ?", existingClient.UID, weight.PaymentMethod).Delete(&model.ChannelRouteWeight{}).Error; err != nil {
				log.Printf("Failed to delete existing route weights: %s", err)
				return err
			}
		}

		// Create new route weights
		for _, weight := range channelRouteWeights {
			if err := db.Create(&weight).Error; err != nil {
				log.Printf("Failed to create supplier route weight: %+v, error: %v", weight, err)
				return fmt.Errorf("failed to create supplier route weight: %w", err)
			}
		}
	}

	// Handle client app updates
	for _, app := range input.ClientApp {
		var existingApps model.ClientApp

		// Check if app has ID (for existing apps) or find by app_name (for new apps)
		var query *gorm.DB
		if app.AppID != "" {
			// Update existing app by app_id
			query = db.Where("client_id = ? AND app_id = ?", existingClient.UID, app.AppID)
		} else {
			// Find by app_name for new or existing apps
			query = db.Where("client_id = ? AND app_name = ?", existingClient.UID, app.AppName)
		}

		if err := query.First(&existingApps).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new client app if it doesn't exist
				app.ClientID = existingClient.UID
				// Clear ID to let database auto-generate
				app.ID = 0
				if err := AddClientApps(existingClient.UID, &app); err != nil {
					log.Printf("Failed to add client app for client %s: %s", existingClient.UID, err)
					return err
				}
			} else {
				log.Printf("Failed to check existing client app: %s", err)
				return err
			}
		} else {
			// Update existing app only if properties are provided
			if app.AppName != "" {
				existingApps.AppName = app.AppName
			}
			if app.CallbackURL != "" {
				existingApps.CallbackURL = app.CallbackURL
			}
			// Update testing and status if provided (including 0 values)
			existingApps.Testing = app.Testing
			existingApps.Status = app.Status
			if app.FailCallback != "" {
				existingApps.FailCallback = app.FailCallback
			}
			if app.Mobile != "" {
				existingApps.Mobile = app.Mobile
			}

			// Save the updated client app
			if err := db.Save(&existingApps).Error; err != nil {
				log.Printf("Failed to update app for client %s: %s", existingClient.UID, err)
				return err
			}
		}
	}

	// Refresh cache with updated client data including all related data
	var updatedClient model.Client
	if err := db.Where("client_id = ?", clientID).
		Preload("ClientApps").
		Preload("PaymentMethods").
		Preload("Settlements").
		Preload("ChannelRouteWeight").
		First(&updatedClient).Error; err != nil {
		log.Printf("Failed to reload updated client for cache: %s", err)
		// Don't return error here, just log it as cache refresh is not critical
	} else {
		// Clear old cache and set new cache with all app_key combinations
		merchantCache.Delete(cacheKey)
		for _, app := range updatedClient.ClientApps {
			newCacheKey := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Set(newCacheKey, &updatedClient, cache.DefaultExpiration)
		}
		log.Printf("Cache refreshed for client: %s", clientID)
	}

	return nil
}

func ValidatePaymentMethodSettlementConsistency(paymentMethods []model.PaymentMethodClient, settlements []model.SettlementClient) error {
	// Create a map of payment method names for quick lookup
	paymentMethodMap := make(map[string]bool)
	for _, pm := range paymentMethods {
		paymentMethodMap[pm.Name] = true
	}

	// Check that each settlement corresponds to a payment method
	for _, settlement := range settlements {
		if !paymentMethodMap[settlement.Name] {
			return fmt.Errorf("settlement '%s' does not have a corresponding payment method", settlement.Name)
		}
	}

	// Check that each payment method has a corresponding settlement
	settlementMap := make(map[string]bool)
	for _, settlement := range settlements {
		settlementMap[settlement.Name] = true
	}

	for _, pm := range paymentMethods {
		if !settlementMap[pm.Name] {
			return fmt.Errorf("payment method '%s' does not have a corresponding settlement", pm.Name)
		}
	}

	return nil
}

// ClearClientCache menghapus cache untuk client tertentu
func ClearClientCache(cacheKey string) {
	merchantCache.Delete(cacheKey)
}

func ClearClientCacheByClientUID(clientUID string) {
	// Ambil semua app milik client ini lalu delete berdasarkan kombinasi app_key + app_id
	var apps []model.ClientApp
	if err := database.DB.Where("client_id = ?", clientUID).Find(&apps).Error; err == nil {
		for _, app := range apps {
			key := fmt.Sprintf("client:%s:%s", app.AppKey, app.AppID)
			merchantCache.Delete(key)
		}
	}
}

// UpdateClientProfile mengupdate data client (email dan address)
func UpdateClientProfile(ctx context.Context, clientUID string, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return nil
	}

	if err := database.DB.Model(&model.Client{}).Where("uid = ?", clientUID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update client data: %w", err)
	}

	return nil
}

// GetClientAppByID mengambil client app berdasarkan client_id dan app_id
func GetClientAppByID(ctx context.Context, clientUID, appID string) (*model.ClientApp, error) {
	var clientApp model.ClientApp
	if err := database.DB.Where("client_id = ? AND app_id = ?", clientUID, appID).First(&clientApp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client app with AppID %s not found", appID)
		}
		return nil, fmt.Errorf("failed to get client app: %w", err)
	}

	return &clientApp, nil
}

// UpdateClientApp mengupdate data client app
func UpdateClientApp(ctx context.Context, clientApp *model.ClientApp) error {
	if err := database.DB.Save(clientApp).Error; err != nil {
		return fmt.Errorf("failed to update client app: %w", err)
	}

	return nil
}

// UpdateClientAppFields mengupdate field tertentu dari client app
func UpdateClientAppFields(ctx context.Context, clientUID, appID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	if err := database.DB.Model(&model.ClientApp{}).Where("client_id = ? AND app_id = ?", clientUID, appID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update client app fields: %w", err)
	}

	return nil
}

// UpdateClientApps mengupdate multiple client apps
func UpdateClientApps(ctx context.Context, clientUID string, appUpdates []model.ClientAppUpdate) error {
	for _, appUpdate := range appUpdates {
		if appUpdate.AppID == "" {
			return fmt.Errorf("AppID is required for each client app update")
		}

		// Validasi bahwa app yang diupdate adalah milik client ini
		clientApp, err := GetClientAppByID(ctx, clientUID, appUpdate.AppID)
		if err != nil {
			return err
		}

		// Update fields yang disediakan
		if appUpdate.CallbackURL != nil {
			clientApp.CallbackURL = *appUpdate.CallbackURL
		}

		if appUpdate.FailCallback != nil {
			clientApp.FailCallback = *appUpdate.FailCallback
		}

		if appUpdate.Mobile != nil {
			clientApp.Mobile = *appUpdate.Mobile
		}

		// Simpan perubahan
		if err := UpdateClientApp(ctx, clientApp); err != nil {
			return fmt.Errorf("failed to update client app %s: %w", appUpdate.AppID, err)
		}
	}

	return nil
}
