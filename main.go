// Пакет main запускает веб-сервер todo листа и подключает его к
// настроенной базе данных SQLite
package main

import (
	"final_project/pkg/db"
	"final_project/pkg/server"
	"log"
	"os"
)

// main считывает путь к базе данных из файла TODO_DBFILE, инициализирует хранилище и
// запускает HTTP-сервер. Если файл TODO_DBFILE пуст, main использует scheduler.db в
// текущем каталоге
func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	if err := db.Init(dbFile); err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	if err := server.StartServer("7540"); err != nil {
		log.Fatal(err)
	}
}
