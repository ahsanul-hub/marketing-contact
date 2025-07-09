package helper

import (
	"app/dto/model"
	"fmt"
	"math/rand"
	"time"

	"github.com/patrickmn/go-cache"
)

var channelRouteCache = cache.New(2*time.Hour, 2*time.Hour)

func ChooseRouteByWeight(weights []model.ChannelRouteWeight) string {
	total := 0
	for _, w := range weights {
		total += w.Weight
	}
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(total)

	accum := 0
	for _, w := range weights {
		accum += w.Weight
		if r < accum {
			return w.Route
		}
	}
	return ""
}

func GetRouteWeightFromClient(client *model.Client, paymentSlug string) ([]model.ChannelRouteWeight, error) {
	var routes []model.ChannelRouteWeight
	for _, r := range client.ChannelRouteWeight {
		if r.PaymentMethod == paymentSlug {
			routes = append(routes, r)
		}
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("no route found for paymentSlug: %s", paymentSlug)
	}
	return routes, nil
}
