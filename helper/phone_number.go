package helper

import (
	"app/repository"
	"strings"
)

func BeautifyIDNumber(mdn string, zero bool) string {
	check := true

	if mdn == "" {
		return ""
	}

	for check {
		check = false

		// Remove non-numeric prefix
		if len(mdn) > 0 && !isNumeric(string(mdn[0])) {
			mdn = mdn[1:]
			check = true
		}

		// Remove '62' prefix
		if strings.HasPrefix(mdn, "62") {
			mdn = mdn[2:]
			check = true
		}

		// Remove leading '0's
		for strings.HasPrefix(mdn, "0") {
			mdn = mdn[1:]
			check = true
		}
	}

	if zero {
		mdn = "0" + mdn
	} else {
		mdn = "62" + mdn
	}

	return mdn
}

func ByPrefixNumber(method string, userMdn string) (bool, error) {
	// Log start time (optional)
	// log.Printf("Settings@byPrefixNumber starts at %s", time.Now().Format("2006-01-02 15:04:05"))

	exist := false
	charging, err := repository.FindPaymentMethodBySlug(method, "") //getSettingsBySlug(method)
	if err != nil {
		return false, err
	}

	if charging != nil {
		prefixes := charging.Prefix
		for _, prefix := range prefixes {
			if strings.HasPrefix(userMdn, prefix) {
				exist = true
				break
			}
		}
	}

	// Log end time (optional)
	// log.Printf("Settings@byPrefixNumber ends at %s", time.Now().Format("2006-01-02 15:04:05"))
	return exist, nil
}

// isNumeric checks if a string is numeric
func isNumeric(str string) bool {
	return str >= "0" && str <= "9"
}
