package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func DefaultPriority(priority, index int) int {
	if priority != 0 {
		return priority
	}
	return index + 1
}

func DefaultMap(value map[string]interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	return value
}

func ParseOccurredAt(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed
}

func ParseOccurredAtStrict(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, InvalidError("occurredAt must be an RFC3339 timestamp")
	}
	return parsed, nil
}

func EnsureActiveStatus(entity, status string) error {
	if status != StatusActive {
		return InvariantError(fmt.Sprintf("%s is not active", entity))
	}
	return nil
}

func HashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return "", InvalidError("password must be at least 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, passwordHash string) bool {
	if strings.TrimSpace(password) == "" || strings.TrimSpace(passwordHash) == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

func ValidateTransactionType(value string) error {
	switch value {
	case EventPurchase, EventRefund:
		return nil
	default:
		return InvalidError(fmt.Sprintf("unsupported transaction type %q", value))
	}
}

func SupportedRuleStrategy(value string) bool {
	switch value {
	case RuleStrategyStack, RuleStrategyMaxOf, RuleStrategyWaterfall:
		return true
	default:
		return false
	}
}

func SupportedRuleType(value string) bool {
	switch value {
	case RuleTypePointsPerDollar, RuleTypeFixedPerTransaction, RuleTypeFirstPurchaseBonus, RuleTypeSpendWindowBonus:
		return true
	default:
		return false
	}
}

func SupportedLimitPeriod(value string) bool {
	switch value {
	case "", "lifetime", "day", "calendar_month", "calendar_year":
		return true
	default:
		return false
	}
}

func SupportedLimitScope(value string) bool {
	return value == "" || value == "member"
}

func HashTransactionPayload(body TransactionIngestRequest) string {
	payload, err := json.Marshal(body)
	if err != nil {
		payload = []byte(`{}`)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func SanitizedTransactionPayload(body TransactionIngestRequest) JSONMap {
	return JSONMap{
		"externalTransactionId":         body.ExternalTransactionID,
		"originalExternalTransactionId": body.OriginalExternalTransactionID,
		"type":                          body.Type,
		"currency":                      body.Currency,
		"subtotalMinor":                 body.SubtotalMinor,
		"taxMinor":                      body.TaxMinor,
		"totalMinor":                    body.TotalMinor,
		"eligibleMinor":                 body.EligibleMinor,
		"occurredAt":                    body.OccurredAt,
		"lineItems":                     body.LineItems,
	}
}

func HashIdentifierValue(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(value))))
	return fmt.Sprintf("%x", sum[:])
}

func HashSecret(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func GenerateToken(prefix string) string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%s", prefix, base64.RawURLEncoding.EncodeToString(bytes))
}

func NormalizeIdentifierValue(identifierType, value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(strings.TrimSpace(identifierType)) {
	case "email":
		return strings.ToLower(value)
	case "phone":
		digits := ""
		for _, r := range value {
			if r >= '0' && r <= '9' {
				digits += string(r)
			}
		}
		if len(digits) == 10 {
			return "+1" + digits
		}
		if len(digits) == 11 && strings.HasPrefix(digits, "1") {
			return "+" + digits
		}
		return digits
	case "qr_code", "square_customer_id", "toast_guest_id":
		return value
	default:
		return strings.ToLower(value)
	}
}

func HashIdentifier(identifierType, value string) string {
	return HashIdentifierValue(NormalizeIdentifierValue(identifierType, value))
}

func PeriodKey(period string, occurredAt time.Time) string {
	switch period {
	case "day":
		return occurredAt.Format("2006-01-02")
	case "calendar_month":
		return occurredAt.Format("2006-01")
	case "calendar_year":
		return occurredAt.Format("2006")
	default:
		return "lifetime"
	}
}

func ConfigInt(config JSONMap, keys ...string) int {
	for _, key := range keys {
		if value, ok := config[key]; ok {
			return IntFromAny(value)
		}
	}
	return 0
}

func ConfigFloat(config JSONMap, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := config[key]; ok {
			return FloatFromAny(value)
		}
	}
	return 0
}

func BoolConfig(config JSONMap, key string) bool {
	value, ok := config[key]
	if !ok {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	if s, ok := value.(string); ok {
		return s == "true"
	}
	return false
}

func StringSliceConfig(config JSONMap, key string) []string {
	value, ok := config[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []interface{}:
		out := []string{}
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		return out
	case []string:
		return typed
	case string:
		if typed == "" {
			return nil
		}
		return strings.Split(typed, ",")
	default:
		return nil
	}
}

func ContainsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func IntFromAny(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		i, _ := strconv.Atoi(typed.String())
		return i
	case string:
		i, _ := strconv.Atoi(typed)
		return i
	default:
		return 0
	}
}

func FloatFromAny(value interface{}) float64 {
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case float64:
		return typed
	case json.Number:
		f, _ := strconv.ParseFloat(typed.String(), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(typed, 64)
		return f
	default:
		return 0
	}
}

func StringFromAny(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func Prorate(value, numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return (value * numerator) / denominator
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
