package helper

import "strings"

var ValidPrefixes = map[string][]string{
	"telkomsel_airtime": {"62811", "62812", "62813", "62852", "62853", "62821", "62822", "62823", "62851"},
	"three_airtime":     {"62895", "62896", "62897", "62898", "62899"},
	"indosat_airtime":   {"62815", "62816", "62856", "62857", "62858", "62814", "62855"},
	"smartfren_airtime": {"62888", "62889", "62881", "62882", "62883", "62884", "62885", "62886", "62887"},
	"xl_airtime":        {"62817", "62818", "62819", "62859", "62877", "62878", "62879", "62831", "62832", "62833", "62838", "62839", "62834", "62835", "62836", "62837"},
}

// 3	Telkomsel	62811, 62812, 62813, 62852, 62853, 62821, 62822, 62823, 62851
// 4	Axis	62831, 62832, 62833, 62838, 62839, 62834, 62835, 62836, 62837
// 5	Three	62895, 62896, 62897, 62898, 62899
// 6	Indosat	62815, 62816, 62856, 62857, 62858, 62814, 62855
// 8	SmartFren	62888, 62889, 62881, 62882, 62883, 62884, 62885, 62886, 62887
// 10	XL	62817, 62818, 62819, 62859, 62877, 62878, 62879

// Fungsi untuk mengecek apakah prefix valid
func IsValidPrefix(userMDN, paymentMethod string) bool {
	// Jika payment method ada di map, cek prefix sesuai payment method
	if validPrefixList, exists := ValidPrefixes[paymentMethod]; exists {
		for _, prefix := range validPrefixList {
			if strings.HasPrefix(userMDN, prefix) {
				return true
			}
		}
		return false
	}

	// Jika payment method tidak ada di map, cek semua prefix
	for _, prefixList := range ValidPrefixes {
		for _, prefix := range prefixList {
			if strings.HasPrefix(userMDN, prefix) {
				return true
			}
		}
	}
	return false
}
