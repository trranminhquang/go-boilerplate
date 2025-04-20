package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func sendJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error encoding json response: %v", data))
	}

	w.WriteHeader(status)
	_, err = w.Write(b)

	return err
}
