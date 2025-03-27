package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.abhishek/internal/validator"
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

// The readString() helper returns a string value from the query string or the provided
// default value if no matching key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return the empty string "".

	s := qs.Get(key)

	// If no key exists (or the value is empty) the return the default value.
	if s == "" {
		return defaultValue
	}

	// Otherwise return the string
	return s
}

// The readCSV() help reads a string value from the query string and the splits it
// into a slice on the comma charachter. If no matching key could be found, it returns
// the provided default value.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// Extract the value from the query string.
	csv := qs.Get(key)

	// if no key exists (or the value is empty) the return the default value
	if csv == "" {
		return defaultValue
	}

	// Otherwise parse the value into a []string slice and return it.
	return strings.Split(csv, ",")
}

// The readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
func (app *application) readInts(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from the query string.
	s := qs.Get(key)

	// if no key exists (or the value is empty) the return the default value.
	if s == "" {
		return defaultValue
	}

	// Try to convert the value to an int, If this fails, add an error message to the
	// validator instance and return the default value.
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	// Otherwise, return the converted interger value.
	return i
}
