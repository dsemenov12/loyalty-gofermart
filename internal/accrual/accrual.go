package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsemenov12/loyalty-gofermart/internal/logger"
	"github.com/dsemenov12/loyalty-gofermart/internal/models"
	"go.uber.org/zap"
)

type AccrualClient interface {
	GetAccrualInfo(orderNumber string) (*models.AccrualInfo, error)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 100,
				IdleConnTimeout: 10 * time.Second,
				DisableKeepAlives: false,
			},
		},
		baseURL: baseURL,
	}
}

// Получает статус заказа и количество начисленных баллов из стороннего сервиса
func (c *Client) GetAccrualInfo(orderNumber string) (*models.AccrualInfo, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	logger.Log.Info("Order processing", zap.String("status", strconv.Itoa(resp.StatusCode)))

	// Обработка кодов ответа
	switch resp.StatusCode {
	case http.StatusOK:
		var accrual models.AccrualInfo
		if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
			return nil, err
		}
		return &accrual, nil
	case http.StatusNoContent:
		return nil, errors.New("order not found")
	case http.StatusTooManyRequests:
		return nil, errors.New("too many requests")
	default:
		return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
}