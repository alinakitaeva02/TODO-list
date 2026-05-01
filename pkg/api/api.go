// Пакет api принимает запросы от браузера/тестов, вызывает функции из пакета db,
// считает даты через NextDate() и возвращает ответы клиенту.
package api

import "net/http"

// Init регистрирует все конечные точки API в HTTP mux по умолчанию.
// Ничего не возвращает и ожидает, что пакет базы данных будет инициализирован до того, как
// запросы дойдут до обработчиков, которые считывают или изменяют задачи
func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
}
