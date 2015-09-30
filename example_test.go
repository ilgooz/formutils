package formutils

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ilgooz/formutils"
	"github.com/ilgooz/httpres"
)

func Example() {
	http.HandleFunc("/users", CreateUserHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	fields := createUserForm{}

	// parse & validate your form and response errors with a pretty json error message
	// if not all fields are valid
	// e.g.
	//
	// HTTP 400
	// {
	// "message": "Invalid Data",
	//   "fields": {
	//     "email": "must be a valid email address",
	//     "password": "must be min 3 chars length"
	//   }
	// }
	if formutils.ParseSend(w, r, &fields) {
		// oh! some fields are not valid, exit your handler
		return
	}

	// OR use formutils.Parse(r, &fields) instead if you don't want to response
	// with an error message automatically.
	// Handle your invalids manually
	// invalids, err := formutils.Parse(r, &fields)

	// everything is OK, fields should be filled with their values
	fmt.Println(fields)
	httpres.Json(w, http.StatusCreated, fields)
}

type createUserForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email" validate:"email,required"`
	Password string `schema:"password" validate:"min=3,required"`
}
