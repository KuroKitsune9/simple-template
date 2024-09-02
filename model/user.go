package model

import (
	"mime/multipart"
	"time"
)

type User struct {
	Id          int64                 `json:"id"`
	Name        string                `json:"name"`
	Email       string                `json:"email"`
	Umur        int64                 `json:"umur"`
	Address     *string               `json:"address"`
	PhoneNumber *string               `json:"phone_number"`
	Gender      *string               `json:"gender"`
	Status      *string               `json:"status"`
	City        *string               `json:"city"`
	Province    *string               `json:"province"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   *time.Time            `json:"updated_at"`
	Image       *multipart.FileHeader `json:"image"`
}

type UserReq struct {
	Name        string `form:"name" validate:"required"`
	Email       string `form:"email" validate:"required,email"`
	Umur        int    `form:"umur" validate:"required,numeric"`
	Password    string `form:"password" validate:"required"`
	Address     string `form:"address" validate:"required"`
	PhoneNumber string `form:"phone_number" validate:"required"`
	Gender      string `form:"gender"`
	Status      string `form:"status"`
	City        string `form:"city" validate:"required"`
	Province    string `form:"province" validate:"required"`
}

type UserDel struct {
	Id []int64 `json:"id"`
}
