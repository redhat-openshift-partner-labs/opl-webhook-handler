package libs

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

// Validate checks a proper LabRequest has been submitted.
// Validate uses the go-playground/validator library to ensure required fields
// are populated and all fields have expected content i.e. valid email address.
func Validate(CurrentLabRequest string) LabRequest {
  // create labRequest struct for this validation request
	var labRequest LabRequest

	// unmarshal the incoming lab request and put into the struct
	err := json.Unmarshal([]byte(CurrentLabRequest), &labRequest)
	if err != nil {
		log.Fatal(err)
	}

	// generate a lab ID and set the labRequest.ID field to generated UUID
	labRequest.ID = uuid.New()

	// validate the request
	validate := validator.New()
	err = validate.Struct(labRequest)
	if err != nil {
		fmt.Printf("Unable to validate the request: %v", err)
		log.Fatal(err)
	}

	return labRequest
}
