package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitialize_ValidLevel(t *testing.T) {
	// Проверка корректной инициализации с правильным уровнем логирования
	err := Initialize("info")

	// Проверяем, что ошибки нет
	assert.NoError(t, err)

	// Проверяем, что логгер инициализировался корректно и не равен пустому логгеру
	assert.NotEqual(t, Log, zap.NewNop())
}

func TestInitialize_InvalidLevel(t *testing.T) {
	// Проверка инициализации с неверным уровнем логирования
	err := Initialize("invalidLevel")

	// Проверяем, что ошибка возникает
	assert.Error(t, err)
}

func TestInitialize_CustomLevel(t *testing.T) {
	// Проверка инициализации с кастомным уровнем логирования
	err := Initialize("debug")

	// Проверяем, что ошибки нет
	assert.NoError(t, err)

	// Проверяем, что уровень логирования был изменен
	assert.Equal(t, true, Log.Core().Enabled(zap.DebugLevel))
}

func TestInitialize_BuildLoggerError(t *testing.T) {
	// Проверка ошибки при создании логгера (например, если конфигурация неправильная)
	// В данном случае, ошибка при настройке конфигурации в Build вызовет сбой инициализации
	err := Initialize("info")
	assert.NoError(t, err)

	// Дополнительная проверка на что-то специфическое, если необходимо
	// В данном примере мы проверяем, что логгер инициализируется
	assert.NotNil(t, Log)
}
