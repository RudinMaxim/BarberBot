package helper

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
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

func IsValidPhone(phone string) bool {
	re := regexp.MustCompile(`^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$`)
	return re.MatchString(phone)
}

func GetText(key string) string {
	return config.Texts[key]
}

type TemplateData struct {
	Name string
}

func GetFormattedMessage(key string, name string) string {
	text := GetText(key)

	tmpl, err := template.New("message").Parse(text)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return ""
	}

	// Подготовка буфера для записи результата
	var result bytes.Buffer
	data := TemplateData{Name: name}

	// Выполняем подстановку значений
	err = tmpl.Execute(&result, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return result.String()
}
