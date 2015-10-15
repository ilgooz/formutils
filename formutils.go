package formutils

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/schema"
	"github.com/ilgooz/eres"
	"gopkg.in/go-playground/validator.v8"
)

var (
	// gorilla/schema
	// Any type that you decode your form fields into
	ParsingErrorMessages = map[reflect.Type][]string{
		reflect.TypeOf(time.Time{}): []string{"must be a UTC date", "must be list of UTC dates"},
		reflect.TypeOf(int(0)):      []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(int8(0)):     []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(int16(0)):    []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(int32(0)):    []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(int64(0)):    []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(uint8(0)):    []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(uint16(0)):   []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(uint32(0)):   []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(uint64(0)):   []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(float32(0)):  []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(float64(0)):  []string{"must be a number", "must be list of numbers"},
		reflect.TypeOf(false):       []string{"must be boolean", "must be list of booleans"},
	}

	// go-playground/validator
	// Refer: https://godoc.org/gopkg.in/go-playground/validator.v8#hdr-Baked_In_Validators_and_Tags
	ValidationErrorMessages = map[string][]string{
		"email": []string{"must be a valid email address", "must be group of valid email addresses"},
		"min":   []string{"must be min %s chars length", "items must be min %s chars length"},
	}
)

var (
	decoder  *schema.Decoder
	validate *validator.Validate
)

func init() {
	decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	config := &validator.Config{
		TagName:      "validate",
		FieldNameTag: "schema",
	}
	validate = validator.New(config)
}

// Parse parses and validates your *http.Request.Form
// and returns invalids fields within a map.
// error returned if the parsing process fails
func Parse(r *http.Request, out interface{}) (invalids map[string]string, err error) {
	invalids, err = parseForm(out, r)

	if err != nil {
		return invalids, err
	}

	for key, val := range validateForm(out) {
		if _, exists := invalids[key]; exists {
			continue
		}

		invalids[key] = val
	}

	return invalids, nil
}

// ParseSend does the same thing with Parse but responses the http request
// with a formatted json error mesages for invalids fields and returns a bool
// according to any invalid fields found or not
func ParseSend(w http.ResponseWriter, r *http.Request, out interface{}) (invalid bool) {
	invalids, err := Parse(r, out)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}

	return eres.New(w).SetFields(invalids).WeakSend()
}

func parseForm(out interface{}, r *http.Request) (invalids map[string]string, err error) {
	invalids = make(map[string]string)

	if err := r.ParseForm(); err != nil {
		return invalids, err
	}

	err = decoder.Decode(out, r.Form)
	if err == nil {
		return invalids, err
	}

	multiErrs, ok := err.(schema.MultiError)
	if !ok {
		return invalids, err
	}

	for _, multiErr := range multiErrs {
		convErr, ok := multiErr.(schema.ConversionError)

		if !ok {
			return invalids, err
		}

		var message string

		if messages, exists := ParsingErrorMessages[convErr.Type]; exists {
			if convErr.Index == -1 {
				message = messages[0]
			} else {
				message = messages[1]
			}
		} else {
			message = "invalid"
		}

		invalids[convErr.Key] = message
	}

	return invalids, nil
}

func validateForm(out interface{}) (invalids map[string]string) {
	invalids = make(map[string]string)

	err := validate.Struct(out)

	if err == nil {
		return invalids
	}

	for _, e := range err.(validator.ValidationErrors) {
		var message string

		if messages, exists := ValidationErrorMessages[e.Tag]; exists {
			if e.Kind == reflect.Slice {
				message = messages[1]
			} else {
				message = messages[0]
			}

			if e.Param != "" {
				message = fmt.Sprintf(message, e.Param)
			}
		} else {
			if e.Param == "" {
				message = e.Tag
			} else {
				message = fmt.Sprintf("%s: %s", e.Tag, e.Param)
			}
		}

		invalids[e.Name] = message
	}

	return invalids
}
