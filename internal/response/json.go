package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func DecodeJSON[T any](r *http.Request, destination *T) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.New("failed to read request body")
	}

	if len(strings.TrimSpace(string(body))) == 0 {
		return errors.New("request body must not be empty")
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(body, &raw); err != nil {
		return errors.New("request body must be a valid JSON object")
	}

	if len(raw) == 0 {
		return errors.New("request body must not be empty")
	}

	if err := json.Unmarshal(body, destination); err != nil {
		return errors.New("request body contains invalid field values")
	}

	return nil
}
