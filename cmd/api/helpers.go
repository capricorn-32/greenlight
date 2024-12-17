package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	var maxBytes int64 = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// Decoding request body
	dec := json.NewDecoder(r.Body)

	// disallowing unknown field in the request body,
	dec.DisallowUnknownFields()

	// Decode the request body to the destination
	if err := dec.Decode(dst); err != nil {
		var ErrSyntax *json.SyntaxError
		var ErrUnmarshalType *json.UnmarshalTypeError
		var ErrInvalidUnmarshal *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &ErrSyntax):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", ErrSyntax.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &ErrUnmarshalType):
			if ErrUnmarshalType.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", ErrUnmarshalType.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", ErrUnmarshalType.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &ErrInvalidUnmarshal):
			panic(err)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only contain a single json value")
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
