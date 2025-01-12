package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"fmt"
	"reflect"

	"github.com/dsemenov12/loyalty-gofermart/internal/auth"
	"github.com/dsemenov12/loyalty-gofermart/internal/config"
	"github.com/dsemenov12/loyalty-gofermart/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

type StorageDB struct {
    conn *sql.DB
}

func NewStorage(conn *sql.DB) *StorageDB {
    return &StorageDB{conn: conn}
}

// Добавляет нового пользователя в базу данных
func (s *StorageDB) CreateUser(login, hashedPassword string) error {
	_, err := s.conn.Exec(`
		INSERT INTO users (login, password)
		VALUES ($1, $2)
	`, login, hashedPassword)

	if err != nil {
		// Проверка ошибки на наличие уникального ограничения
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
            return errors.New("user already exists")
        }
		return err
	}
	return nil
}

// Возвращает пользователя по логину
func (s *StorageDB) GetUserByLogin(login string) (*models.User, error) {
	row := s.conn.QueryRow(`
		SELECT id, login, password
		FROM users
		WHERE login = $1
	`, login)

	var user models.User
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Сохранение заказа
func (s *StorageDB) SaveOrder(ctx context.Context, orderNumber string) (bool, error) {
	userID := ctx.Value(auth.UserIDKey)

	// Проверка существующего номера заказа
	var existingUserID string
	err := s.conn.QueryRow(`
		SELECT user_id FROM orders WHERE number = $1
	`, orderNumber).Scan(&existingUserID)

	if err == nil {
		fmt.Println(reflect.TypeOf(existingUserID))
		fmt.Println(reflect.TypeOf(userID))
		if existingUserID == userID {
			return false, errors.New("order already exists for the same user")
		}
		return false, errors.New("order already exists for another user")
	}

	if err != sql.ErrNoRows {
		return false, err
	}

	// Вставка нового заказа
	_, err = s.conn.Exec(`
		INSERT INTO orders (user_id, number, status, created_at)
		VALUES ($1, $2, 'PROCESSING', NOW())
	`, userID, orderNumber)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Список заказов пользователя
func (s *StorageDB) GetOrdersByUser(ctx context.Context) ([]models.Order, error) {
	userID := ctx.Value(auth.UserIDKey)
	
	rows, err := s.conn.Query(`
		SELECT number, status, accrual, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Получение баланса пользователя
func (s *StorageDB) GetBalance(ctx context.Context) (*models.Balance, error) {
	var balance models.Balance
	userID := ctx.Value(auth.UserIDKey)

	err := s.conn.QueryRowContext(ctx, `
		SELECT current, withdrawn
		FROM balance
		WHERE user_id = $1
	`, userID).Scan(&balance.Current, &balance.Withdrawn)

	
	if err == sql.ErrNoRows {
		balance.Current = 0
		balance.Withdrawn = 0
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &balance, nil
}

// Списание средств
func (s *StorageDB) WithdrawUserBalance(ctx context.Context, orderNumber string, amount float64) error {
	userID := ctx.Value(auth.UserIDKey)

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверка текущего баланса
	var currentBalance float64
	err = tx.QueryRowContext(ctx, `
		SELECT current
		FROM balance
		WHERE user_id = $1
	`, userID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return errors.New("balance not found")
	}
	if err != nil {
		return err
	}

	if currentBalance < amount {
		return errors.New("insufficient funds")
	}

	// Обновление баланса
	_, err = tx.ExecContext(ctx, `
		UPDATE balance
		SET current = current - $2, withdrawn = withdrawn + $2
		WHERE user_id = $1
	`, userID, amount)
	if err != nil {
		return err
	}

	// Добавление записи в таблицу withdraw
	_, err = tx.ExecContext(ctx, `
		INSERT INTO withdraw (user_id, order_number, sum, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, orderNumber, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Получение списка списаний
func (s *StorageDB) GetUserWithdrawals(ctx context.Context) ([]models.Withdrawal, error) {
	userID := ctx.Value(auth.UserIDKey)

	rows, err := s.conn.QueryContext(ctx, `
		SELECT order_number, sum, created_at
		FROM withdraw
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		var createdAt time.Time
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &createdAt)
		if err != nil {
			return nil, err
		}
		withdrawal.ProcessedAt = createdAt.Format(time.RFC3339)
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

// Получает статус заказа и количество начисленных баллов из стороннего сервиса
func (s *StorageDB) GetAccrualInfo(orderNumber string) (*models.AccrualInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", config.FlagAccrualSystemAddress, orderNumber), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

// Обновляет статус заказа и количество начисленных баллов
func (s *StorageDB) UpdateOrderStatus(orderNumber, status string, accrual float64) error {
	query := `
		UPDATE orders
		SET status = $2, accrual = $3, updated_at = NOW()
		WHERE order_number = $1
	`
	_, err := s.conn.Exec(query, orderNumber, status, accrual)
	return err
}