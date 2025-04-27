package models

import "errors"

var (
	ErrIncorrectProductType = errors.New("incorrect product type")
	ErrIncorrectCity        = errors.New("incorrect city")

	ErrIncorrectUserRole = errors.New("incorrect user role")
	ErrEmailAlreadyInUse = errors.New("user with this email already exists")
	ErrWrongPassword     = errors.New("wrong password")
	ErrUserNotFound      = errors.New("user not found")

	ErrReceptionNotFound     = errors.New("reception not found")
	ErrReceptionClosed       = errors.New("reception closed")
	ErrNoProductsInReception = errors.New("reception is empty")

	ErrReceptionNotClosed = errors.New("previous reception not closed")
)
