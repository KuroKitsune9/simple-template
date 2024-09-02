package model

import "time"

type CategoryReq struct {
	Id           []int  `json:"id"`
	CategoryName string `form:"category_name"`
}

type CategoryRes struct {
	Id           int        `json:"id"`
	CategoryName string     `json:"category_name"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
