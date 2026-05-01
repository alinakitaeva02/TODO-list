package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Task описывает одну задачу планировщика, хранящуюся в базе данных и возвращаемую API
// обработчиками. Идентификатор хранится в виде текста в ответах JSON
type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

// AddTask вставляет задачу в таблицу планировщика.
// Возвращает сгенерированный идентификатор базы данных или ошибку из SQLite
func AddTask(task *Task) (int64, error) {
	res, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Tasks возвращает максимальное количество задач, упорядоченных по дате.
// Возвращает пустой слайс, если задач не существует, и возвращает ошибки запроса или проверки строк из SQLite
func Tasks(limit int) ([]*Task, error) {
	rows, err := DB.Query(
		`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// SearchTasks возвращает задачи, соответствующие поиску, упорядоченному по дате.
// Если поиск выполняется в формате 02.01.2006, он преобразуется в 20060102 и сопоставляется
// со столбцом даты; в противном случае строка поиска сопоставляется как аналогичная
// подстрока в заголовке или комментарии
func SearchTasks(search string, limit int) ([]*Task, error) {
	date, err := time.Parse("02.01.2006", search)
	if err == nil {
		rows, err := DB.Query(
			`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`,
			date.Format("20060102"),
			limit,
		)
		if err != nil {
			return nil, err
		}
		return scanTasks(rows)
	}

	like := "%" + search + "%"
	rows, err := DB.Query(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`,
		like,
		like,
		limit,
	)
	if err != nil {
		return nil, err
	}
	return scanTasks(rows)
}

// scanTasks преобразует строки планировщика в значения Task.
// Программа закрывает строки перед возвратом и сообщает об ошибках сканирования строк или итерации
func scanTasks(rows *sql.Rows) ([]*Task, error) {
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		task := &Task{}
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// getTask возвращает задачу по идентификатору.
// Возвращает "задача не найдена", когда для идентификатора не существует строки
func GetTask(id string) (*Task, error) {
	task := &Task{}
	err := DB.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	return task, nil
}

// UpdateTask обновляет все поля задачи для существующего идентификатора.
// Возвращает ошибку при сбое SQLite или при отсутствии обновленной строки
func UpdateTask(task *Task) error {
	res, err := DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}

	return nil
}

// UpdateDate обновляет только столбец даты для задачи с идентификатором.
// Он возвращает ошибку при сбое SQLite или когда задача не найдена
func UpdateDate(next string, id string) error {
	res, err := DB.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, next, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask удаляет задачу с идентификатором из таблицы планировщика.
// Она возвращает значение nil после успешного удаления и возвращает ошибку, если задача
// отсутствует или операция с базой данных завершается сбоем
func DeleteTask(id string) error {
	res, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
