package utils

import (
	"strconv"
	"strings"
)

func ValidateOrderNumber(number string) bool {
	number = strings.ReplaceAll(number, " ", "")

	if _, err := strconv.ParseInt(number, 10, 64); err != nil {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
