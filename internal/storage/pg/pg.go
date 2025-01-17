package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dsemenov12/loyalty-gofermart/internal/auth"
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
func (s *StorageDB) CreateUser(ctx context.Context, login, hashedPassword string) error {
	_, err := s.conn.ExecContext(ctx, `
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
func (s *StorageDB) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	row := s.conn.QueryRowContext(ctx, `
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
	err := s.conn.QueryRowContext(ctx, `
		SELECT user_id FROM orders WHERE number = $1
	`, orderNumber).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return false, errors.New("order already exists for the same user")
		}
		return false, errors.New("order already exists for another user")
	}

	if err != sql.ErrNoRows {
		return false, err
	}

	// Вставка нового заказа
	_, err = s.conn.ExecContext(ctx, `
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
	
	rows, err := s.conn.QueryContext(ctx, `
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

// Обновляет статус заказа и количество начисленных баллов
func (s *StorageDB) UpdateOrderStatus(ctx context.Context, orderNumber, status string, accrual float64) error {
	_, err := s.conn.ExecContext(ctx, `
		UPDATE orders
		SET status = $2, accrual = $3, updated_at = NOW()
		WHERE number = $1
	`, orderNumber, status, accrual)
	return err
}

// Обновление баланса пользователя
func (s *StorageDB) UpdateUserBalance(ctx context.Context, sum float64) error {
	userID := ctx.Value(auth.UserIDKey)
    
    // Проверяем, существует ли запись с балансом для данного пользователя
    var currentBalance float64
    err := s.conn.QueryRowContext(ctx, `
        SELECT current FROM balance WHERE user_id = $1
    `, userID).Scan(&currentBalance)

    if err != nil {
        if err.Error() == sql.ErrNoRows.Error() {
            // Если записи нет, то создаем новую запись с начальным значением баланса
            _, err := s.conn.ExecContext(ctx, `
                INSERT INTO balance (user_id, current, withdrawn) 
                VALUES ($1, $2, $3)
            `, userID, sum, 0.0)
            if err != nil {
                return fmt.Errorf("failed to insert balance record: %v", err)
            }
            return nil
        }
        // Другие ошибки
        return fmt.Errorf("failed to check balance: %v", err)
    }

    // Если запись с балансом существует, выполняем обновление
    _, err = s.conn.ExecContext(ctx, `
        UPDATE balance
		SET current = current + $2
		WHERE user_id = $1
    `, userID, sum)
    
    return err
}