package application

import (
	"accts-api/domain"

	"encoding/json"
	"errors"
	"strings"
)

func normalizeDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func mustJSON(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return data
}

func publicFailureReason(err error) string {
	var appErr domain.AppError
	if errors.As(err, &appErr) {
		if appErr.Kind == domain.ErrorKindInternal {
			return "internal processing error"
		}
		return appErr.Error()
	}
	return "internal processing error"
}
