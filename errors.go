package muumuu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ErrorCodeNotFound          = "not_found"
	ErrorCodeRateLimitExceeded = "rate_limit_exceeded"
	ErrorCodeValidationError   = "validation_error"
	ErrorCodeConflict          = "conflict"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type Error struct {
	StatusCode int           `json:"-"`
	RetryAfter time.Duration `json:"-"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
	PurchaseID string        `json:"purchase-id,omitempty"`
	FQDN       string        `json:"fqdn,omitempty"`
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "muumuu-domain: %d", e.StatusCode)
	if e.Code != "" {
		fmt.Fprintf(&b, " %s", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	for _, d := range e.Details {
		fmt.Fprintf(&b, " [%s: %s]", d.Field, d.Message)
	}
	return b.String()
}

func newError(resp *http.Response) error {
	e := &Error{StatusCode: resp.StatusCode}

	if s := resp.Header.Get("Retry-After"); s != "" {
		if sec, err := strconv.Atoi(s); err == nil {
			e.RetryAfter = time.Duration(sec) * time.Second
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Join(e, err)
	}

	var r struct {
		Error *Error `json:"error"`
	}
	if json.Unmarshal(body, &r) == nil && r.Error != nil {
		r.Error.StatusCode = e.StatusCode
		r.Error.RetryAfter = e.RetryAfter
		return r.Error
	}

	e.Message = strings.TrimSpace(string(body))
	if e.Message == "" {
		e.Message = http.StatusText(resp.StatusCode)
	}
	return e
}

func IsNotFound(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusNotFound
}

func IsRateLimited(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.StatusCode == http.StatusTooManyRequests
}
