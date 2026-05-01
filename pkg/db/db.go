// Пакет db создает базу данных SQLite для задач планировщика.
// Отвечает за инициализацию базы данных, создание таблицы и выполнение CRUD-запросов к задачам, используемых
// с помощью HTTP API
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

var DB *sql.DB

// Init открывает базу данных SQLite в dbFile и создает таблицу для планировщика
// если файл еще не существует. Возвращает ошибку, если файл не может быть проверен,
// база данных не может быть открыта или создание таблицы завершается ошибкой
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)
	if err != nil && !install {
		return fmt.Errorf("failed to check database file: %w", err)
	}

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if install {
		if _, err := DB.Exec(schema); err != nil {
			return fmt.Errorf("failed to install database schema: %w", err)
		}
	}

	return nil
}
