package entity

import "errors"

var (

	ErrUserAlreadyExsist = errors.New("user with this email already exisist")
	ErrInternalServer	  = errors.New("internal server error")
	ErrNotFound 		  = errors.New("data not found")
)