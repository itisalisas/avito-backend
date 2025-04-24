package models

import (
	"github.com/google/uuid"
	"time"
)

/* TODO - uncomment when use
var itemTypes = map[string]interface{}{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}
*/

type Product struct {
	Id          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	ItemType    string    `json:"itemType"`
	ReceptionId uuid.UUID `json:"receptionId"`
}
