package veris

import (
	"strconv"
	"strings"
)

type APIError struct {
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	var sb strings.Builder

	sb.WriteString("VerisLabs API error")

	if e.StatusCode > 0 {
		sb.WriteString(" (")
		sb.WriteString(strconv.Itoa(e.StatusCode))
		sb.WriteRune(')')
	}

	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}

	return sb.String()
}
