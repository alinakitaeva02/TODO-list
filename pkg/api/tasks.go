package api

import (
	"net/http"

	"final_project/pkg/db"
)

// TasksResp - это текст ответа в формате JSON для /api/tasks.
// Tasks содержит список выбранных задач и пуст, если соответствующие задачи не найдены
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// TasksHandler обрабатывает GET /api/tasks и возвращает задачи в формате JSON.
// Без поиска он возвращает ближайшие задачи, упорядоченные по дате; при поиске он
// фильтрует по названию, комментарию или дате в формате 02.01.2006. Ошибки возвращаются
// в формате JSON с полем error
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONStatus(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	search := r.FormValue("search")
	tasks, err := db.Tasks(50)
	if search != "" {
		tasks, err = db.SearchTasks(search, 50)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}

// tasksHandler делегирует запросы /api/tasks в TasksHandler.
// Он существует для того, чтобы не экспортировать зарегистрированный обработчик
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	TasksHandler(w, r)
}
