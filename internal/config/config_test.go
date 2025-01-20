package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	// Сначала очистим флаги и переменные окружения
	defer func() {
		os.Clearenv()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	}()

	// Убедимся, что начальные значения пустые
	assert.Equal(t, FlagRunAddr, "")
	assert.Equal(t, FlagLogLevel, "")
	assert.Equal(t, FlagDatabaseURI, "")
	assert.Equal(t, FlagAccrualSystemAddress, "")

	// Пример 1: Тестируем флаги
	os.Args = []string{"cmd", "-a", "192.168.0.1:9090", "-l", "debug", "-d", "postgres://user:pass@localhost/db", "-r", "192.168.0.1:8081"}
	ParseFlags()

	// Проверяем, что флаги были правильно распарсены
	assert.Equal(t, FlagRunAddr, "192.168.0.1:9090")
	assert.Equal(t, FlagLogLevel, "debug")
	assert.Equal(t, FlagDatabaseURI, "postgres://user:pass@localhost/db")
	assert.Equal(t, FlagAccrualSystemAddress, "192.168.0.1:8081")
}

func TestParseFlagsWithEnvVars(t *testing.T) {
	// Сначала очистим флаги и переменные окружения
	defer func() {
		os.Clearenv()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	}()

	// Установим переменные окружения
	os.Setenv("SERVER_ADDRESS", "192.168.1.1:8082")
	os.Setenv("DATABASE_URI", "postgres://user:pass@localhost/testdb")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "192.168.1.1:8083")

	// Симулируем, что флаги не были переданы, но переменные окружения присутствуют
	os.Args = []string{"cmd"}
	ParseFlags()

	// Проверяем, что переменные окружения переопределили флаги
	assert.Equal(t, FlagRunAddr, "192.168.1.1:8082")
	assert.Equal(t, FlagDatabaseURI, "postgres://user:pass@localhost/testdb")
	assert.Equal(t, FlagAccrualSystemAddress, "192.168.1.1:8083")
}

func TestParseFlagsPriority(t *testing.T) {
	// Сначала очистим флаги и переменные окружения
	defer func() {
		os.Clearenv()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	}()

	// Установим переменные окружения
	os.Setenv("SERVER_ADDRESS", "192.168.1.1:8082")
	os.Setenv("DATABASE_URI", "postgres://user:pass@localhost/testdb")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "192.168.1.1:8083")

	// Симулируем, что флаги были переданы
	os.Args = []string{"cmd", "-a", "192.168.0.1:9090", "-d", "postgres://user:pass@localhost/db", "-r", "192.168.0.1:8081"}
	ParseFlags()

	// Проверяем, что флаги имеют приоритет перед переменными окружения
	assert.Equal(t, FlagRunAddr, "192.168.1.1:8082")
	assert.Equal(t, FlagDatabaseURI, "postgres://user:pass@localhost/testdb")
	assert.Equal(t, FlagAccrualSystemAddress, "192.168.1.1:8083")
}

func TestParseFlagsEmptyArgs(t *testing.T) {
	// Сначала очистим флаги и переменные окружения
	defer func() {
		os.Clearenv()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	}()

	// Убедимся, что флаги не переданы, но установлены переменные окружения
	os.Setenv("SERVER_ADDRESS", "192.168.1.1:8082")
	os.Setenv("DATABASE_URI", "postgres://user:pass@localhost/testdb")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "192.168.1.1:8083")

	// Симулируем отсутствие флагов
	os.Args = []string{"cmd"}
	ParseFlags()

	// Проверяем, что переменные окружения используются
	assert.Equal(t, FlagRunAddr, "192.168.1.1:8082")
	assert.Equal(t, FlagDatabaseURI, "postgres://user:pass@localhost/testdb")
	assert.Equal(t, FlagAccrualSystemAddress, "192.168.1.1:8083")
}
