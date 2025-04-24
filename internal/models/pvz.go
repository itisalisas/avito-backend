package models

import (
	"time"

	"github.com/google/uuid"
)

/* TODO - uncomment when use
var cities = map[string]interface{}{
	"Москва":          true,
	"Санкт-Петербург": true,
	"Казань":          true,
}
*/

type Pvz struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}
