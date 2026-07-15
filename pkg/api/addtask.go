package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"final_project/pkg/db"
)

// taskHandler отправляет запросы /api/task методом HTTP.
// Поддерживает чтение, создание, обновление и удаление задач и возвращает JSON
// ошибки для неподдерживаемых методов
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJSONStatus(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// addTaskHandler обрабатывает POST /api/task и создает задачу из тела JSON.
// Проверяет обязательные поля и правила повторения, возвращает новый идентификатор в случае успеха
// и возвращает JSON с ошибкой при сбое проверки
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := decodeAndValidateTask(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := db.AddTask(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, map[string]any{"id": id})
}

// getTaskHandler обрабатывает GET /api/task?id=... и возвращает одну задачу в формате JSON.
// Требует непустого параметра запроса id и сообщает об ошибках поиска в поле JSON error
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, statusByError(err), err)
		return
	}

	writeJSON(w, task)
}

// updateTaskHandler обрабатывает PUT /api/task и обновляет все редактируемые поля задачи.
// Расшифровывает текст JSON, проверяет задачу, включая ее идентификатор, и возвращает
// пустой объект JSON в случае успеха или JSON с ошибкой в случае сбоя
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := decodeAndValidateTask(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if task.ID == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	if err := db.UpdateTask(task); err != nil {
		writeError(w, statusByError(err), err)
		return
	}

	writeJSON(w, map[string]any{})
}

// deleteTaskHandler обрабатывает DELETE /api/task?id=... и удаляет задачу.
// Возвращает пустой объект JSON после удаления и возвращает JSON с ошибкой, если
// id отсутствует или операция с базой данных завершается ошибкой
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeError(w, statusByError(err), err)
		return
	}

	writeJSON(w, map[string]any{})
}

// doneTaskHandler обрабатывает POST /api/task/done?id=... и помечает задачу выполненной.
// Одноразовые задачи удаляются; для повторных задач вычисляется следующая дата
// по NextDate и обновляется только столбец даты. Ответом будет пустой файл JSON
// в случае успеха или JSON с ошибкой в случае неудачи
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONStatus(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := r.FormValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, statusByError(err), err)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, statusByError(err), err)
			return
		}
		writeJSON(w, map[string]any{})
		return
	}

	next, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeError(w, statusByError(err), err)
		return
	}

	writeJSON(w, map[string]any{})
}

// decodeAndValidateTask десереализует задачу из текста запроса и проверяет ее.
// Возвращает нормализованную задачу или ошибку, если: не удается выполнить десереализацию в формате JSON,
// заголовок пуст или дата неверна
func decodeAndValidateTask(r *http.Request) (*db.Task, error) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		return nil, err
	}

	if task.Title == "" {
		return nil, errors.New("title is required")
	}

	if err := checkDate(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

// checkDate проверяет и нормализует дату задачи.
// Возвращается ошибка для недопустимых дат или повторяющихся правил
func checkDate(task *db.Task) error {
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return err
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format(dateFormat)
		} else {
			task.Date = next
		}
	}

	return nil
}

// writeJSON записывает данные в виде ответа в формате JSON с типом содержимого UTF-8
func writeJSON(w http.ResponseWriter, data any) {
	writeJSONStatus(w, http.StatusOK, data)
}

// writeError записывает ошибку в формате JSON с указанным HTTP-статусом.
// Текст ошибки возвращается клиенту в поле error
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSONStatus(w, status, map[string]string{"error": err.Error()})
}

// writeJSONStatus записывает данные в виде ответа JSON с указанным HTTP-статусом.
// Заголовки устанавливаются до кода ответа, поэтому клиент получает корректный
// Content-Type и статус
func writeJSONStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// statusByError возвращает HTTP-статус для известных ошибок приложения.
// Неизвестные ошибки считаются внутренними ошибками сервера
func statusByError(err error) int {
	if errors.Is(err, db.ErrTaskNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
