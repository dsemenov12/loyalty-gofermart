package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
    "context"

	"github.com/dsemenov12/loyalty-gofermart/internal/auth"
	"github.com/dsemenov12/loyalty-gofermart/internal/helpers/luhn"
	"github.com/dsemenov12/loyalty-gofermart/internal/models"
	"github.com/dsemenov12/loyalty-gofermart/internal/storage"
    "github.com/dsemenov12/loyalty-gofermart/internal/accrual"
	"github.com/dsemenov12/loyalty-gofermart/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type app struct {
	storage storage.Storage
}

func NewApp(storage storage.Storage) *app {
    return &app{storage: storage}
}

// Регистрация пользователя
func (a *app) UserRegister(w http.ResponseWriter, r *http.Request) {
	// Чтение и декодирование тела запроса
	var req models.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя в базе данных
	err = a.storage.CreateUser(req.Login, string(hashedPassword))
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
		} else {
			http.Error(w, "Error saving user", http.StatusInternalServerError)
		}
		return
	}

	// Получение данных пользователя из хранилища
	user, err := a.storage.GetUserByLogin(req.Login)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	setCookieJWT(strconv.Itoa(user.ID), w)

	w.WriteHeader(http.StatusOK)
}

// Аутентификация пользователя
func (a *app) UserLogin(w http.ResponseWriter, r *http.Request) {
	// Чтение и декодирование тела запроса
	var req models.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// Получение данных пользователя из хранилища
	user, err := a.storage.GetUserByLogin(req.Login)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	setCookieJWT(strconv.Itoa(user.ID), w)

	w.WriteHeader(http.StatusOK)
}

// Обрабатывает загрузку номера заказа
func (a *app) UserUploadOrder(w http.ResponseWriter, r *http.Request) {
	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Преобразование и валидация номера заказа.
	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" || !luhn.ValidateLuhn(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Проверка уникальности номера заказа.
	status, err := a.storage.SaveOrder(r.Context(), orderNumber)
	if err != nil {
		if err.Error() == "order already exists for the same user" {
			http.Error(w, "Order number already uploaded by this user", http.StatusOK)
		} else if err.Error() == "order already exists for another user" {
			http.Error(w, "Order number already uploaded by another user", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Если номер принят в обработку
	if status {
		go a.checkOrderStatus(r.Context(), orderNumber)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Order number accepted")
	}
}

// Возвращает список заказов пользователя
func (a *app) UserGetOrders(w http.ResponseWriter, r *http.Request) {
	// Получение списка заказов из хранилища
	orders, err := a.storage.GetOrdersByUser(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Если данных нет, возвращаем 204
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Преобразование данных в формат ответа
	var response []models.OrderResponse
	for _, order := range orders {
		response = append(response, models.OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Возвращает баланс пользователя
func (a *app) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	// Получение баланса
	balance, err := a.storage.GetBalance(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

// Список списаний со счета пользователя
func (a *app) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Получение выводов средств из хранилища
	withdrawals, err := a.storage.GetUserWithdrawals(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Если данных нет
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Возврат успешного ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}

// Списание средств со счета пользователя
func (a *app) WithdrawUserBalance(w http.ResponseWriter, r *http.Request) {
	// Парсинг тела запроса
	var req models.WithdrawRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Order == "" || req.Sum <= 0 {
		http.Error(w, "Invalid request format", http.StatusUnprocessableEntity)
		return
	}

	// Получение баланса
	balance, err := a.storage.GetBalance(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Проверка баланса
	if balance.Current == 0 {
		http.Error(w, "Balance is empty", http.StatusUnprocessableEntity)
		return
	}

	// Списание средств
	err = a.storage.WithdrawUserBalance(r.Context(), req.Order, req.Sum)
	if err != nil {
		switch {
		case err.Error() == "invalid order number":
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		case err.Error() == "insufficient funds":
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		case err.Error() == "balance not found":
			http.Error(w, "Balance not found", http.StatusInternalServerError)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
}

// Запись JWT в куки
func setCookieJWT(userID string, w http.ResponseWriter) {
	tokenString, err := auth.BuildJWTString(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
    cookie := &http.Cookie{
		Name: "Authorization",
		Value: tokenString,
		Expires: time.Now().Add(24 * time.Hour),
		Path: "/",
	}
		
	http.SetCookie(w, cookie)
}

// Фоновая проверка статуса заказа
func (a *app) checkOrderStatus(ctx context.Context, orderNumber string) {
	client := accrual.NewClient(config.FlagAccrualSystemAddress)

	for {
		// Получаем информацию о начислениях
		accrualInfo, err := client.GetAccrualInfo(orderNumber)
		if err != nil {
			if err.Error() == "order not found" {
				// Если заказ не найден, завершаем процесс
				return
			}
			if err.Error() == "too many requests" {
				// Если превышено количество запросов, ждем перед повторной попыткой
				time.Sleep(1 * time.Minute)
				continue
			}
			// В случае других ошибок завершаем процесс
			return
		}

		// Проверяем статус заказа
		switch accrualInfo.Status {
		case "PROCESSED", "INVALID":
			// Если статус окончательный, обновляем в хранилище и завершаем
			_ = a.storage.UpdateOrderStatus(orderNumber, accrualInfo.Status, accrualInfo.Accrual)
            // Обновление баланса пользователя, если статус "PROCESSED"
			if accrualInfo.Status == "PROCESSED" {
				_ = a.storage.UpdateUserBalance(ctx, accrualInfo.Accrual)
			}
			return
		case "PROCESSING", "REGISTERED":
			// Если статус временный, продолжаем проверку через заданный интервал
			time.Sleep(30 * time.Second)
		default:
			// Неизвестный статус, завершаем процесс
			return
		}
	}
}