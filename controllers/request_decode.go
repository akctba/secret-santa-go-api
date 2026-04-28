package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// decodeRequestJSON decodes a single JSON object and rejects unknown fields.
func decodeRequestJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only one json object")
	}

	return nil
}