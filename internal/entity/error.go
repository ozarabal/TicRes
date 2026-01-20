package entity

import "errors"

var (

	ErrUserAlreadyExisist = errors.New("user with this email already exisist")
	ErrInternalServer	  = errors.New("internal server error")
	ErrNotFound 		  = errors.New("data not found")
)