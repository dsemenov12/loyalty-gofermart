package models

import "time"

type UserRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserAuth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID       int
	Login    string
	Password string
}

type Order struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type OrderResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawal struct {
    Order       string  `json:"order"`
    Sum         float64 `json:"sum"`
    ProcessedAt string  `json:"processed_at"`
}

type AccrualInfo struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}