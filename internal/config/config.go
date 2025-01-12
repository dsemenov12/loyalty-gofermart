package config

import (
	"flag"
	"os"
)

var FlagRunAddr string
var FlagLogLevel string
var FlagDatabaseURI string
var FlagAccrualSystemAddress string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&FlagDatabaseURI, "d", "", "адрес подключения к БД")
	flag.StringVar(&FlagAccrualSystemAddress, "r", "127.0.0.1:8081", "адрес системы расчёта начислений")

	flag.Parse()

    if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
        FlagRunAddr = envRunAddr
    }
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
        FlagDatabaseURI = envDatabaseURI
    }
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
        FlagAccrualSystemAddress = envAccrualSystemAddress
    }
}