package helper

import (
	"fmt"
	"regexp"

	"github.com/RudinMaxim/BarberBot.git/config"
)

func NormalizePhoneNumber(phone string) string {
	re := regexp.MustCompile("[^0-9]")
	numbers := re.ReplaceAllString(phone, "")

	if len(numbers) == 0 {
		return ""
	}

	if len(numbers) == 11 && numbers[0] == '8' {
		numbers = "7" + numbers[1:]
	}
	if len(numbers) == 10 {
		numbers = "7" + numbers
	}

	if len(numbers) != 11 {
		return ""
	}

	return fmt.Sprintf("+7 (%s) %s-%s-%s", numbers[1:4], numbers[4:7], numbers[7:9], numbers[9:])
}

func IsValidName(name string) bool {
	re := regexp.MustCompile(`^[а-яА-Яa-zA-Z]{2,}$`)
	return re.MatchString(name)
}

func IsValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$`)
	return re.MatchString(phone)
}

func GetText(key string) string {
	return config.Texts[key]
}
