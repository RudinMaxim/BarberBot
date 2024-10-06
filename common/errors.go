package common

import "errors"

var (
	ErrTelegramTokenNotFound = errors.New("TELEGRAM_TOKEN not set in environment")
)
