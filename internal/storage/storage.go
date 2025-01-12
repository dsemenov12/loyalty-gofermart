package storage

import (
	"context"

	"github.com/dsemenov12/loyalty-gofermart/internal/models"
)

type Storage interface {
	CreateUser(login, hashedPassword string) error
	GetUserByLogin(login string) (*models.User, error)
	SaveOrder(ctx context.Context, orderNumber string) (bool, error)
	GetOrdersByUser(ctx context.Context) ([]models.Order, error)
	GetBalance(ctx context.Context) (*models.Balance, error)
	WithdrawUserBalance(ctx context.Context, orderNumber string, amount float64) error
	GetUserWithdrawals(ctx context.Context) ([]models.Withdrawal, error)
	UpdateOrderStatus(orderNumber, status string, accrual float64) error
	UpdateUserBalance(ctx context.Context, sum float64) error
}