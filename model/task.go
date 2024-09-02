package model

import "time"

type TaskReq struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	Status      string `form:"status"`
	Date        string `form:"date"`
	CategoryId  int    `form:"category_id"`
	Important   bool   `form:"important"`
}

type TaskRes struct {
	Id           int64      `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	Date         time.Time  `json:"date"`
	Image        *string    `json:"image"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	IdUser       int64      `json:"id_user"`
	CategoryId   *int       `json:"category_id"`
	CategoryName *string    `json:"category_name"`
	Important    *bool      `json:"important"`
}
