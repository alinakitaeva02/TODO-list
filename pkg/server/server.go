// Пакет server настраивает и запускает HTTP-сервер для приложения scheduler.
// Регистрирует обработчики API и обрабатывает файлы из веб-каталога
package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"final_project/pkg/api"
)

const defaultPort = "7540"

// startServer запускает HTTP-сервер через порт и обслуживает веб-интерфейс и API.
// Если порт пуст, startServer использует порт пакета по умолчанию.
// Возвращается только при сбое ListenAndServe
func StartServer(port string) error {
	if port == "" {
		port = defaultPort
	}

	fmt.Println("Starting server on port", port)

	webDir, err := filepath.Abs("web")
	if err != nil {
		return fmt.Errorf("failed to resolve web directory: %w", err)
	}

	fileServer := http.FileServer(http.Dir(webDir))
	api.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	addr := ":" + port
	log.Printf("serving %s on http://localhost%s", webDir, addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
