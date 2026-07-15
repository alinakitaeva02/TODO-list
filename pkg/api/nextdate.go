package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

// NextDate вычисляет дату следующего запуска для повторяющейся задачи.
// Возвращает дату в формате 20060102 и поддерживает правила ежедневного, ежегодного, еженедельного
// и ежемесячного повторения. Для пустого правила возвращается ошибка, недопустимое значение
// даты начала, неподдерживаемое правило или недопустимые значения правил
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("repeat rule is empty")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid repeat rule")
	}

	if parts[0] == "d" {
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repeat rule")
		}

		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("invalid repeat rule")
		}

		date = date.AddDate(0, 0, interval)
		for {
			if afterNow(date, now) {
				break
			}
			date = date.AddDate(0, 0, interval)
		}

		return date.Format(dateFormat), nil
	}

	if parts[0] == "y" {
		if len(parts) != 1 {
			return "", fmt.Errorf("invalid repeat rule")
		}

		date = date.AddDate(1, 0, 0)
		for {
			if afterNow(date, now) {
				break
			}
			date = date.AddDate(1, 0, 0)
		}

		return date.Format(dateFormat), nil
	}

	if parts[0] == "w" {
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repeat rule")
		}

		weekdays, err := parseNumbers(parts[1], 1, 7)
		if err != nil {
			return "", err
		}

		next := laterDate(date, now).AddDate(0, 0, 1)
		for i := 0; i < 7; i++ {
			if contains(weekdays, weekdayNumber(next)) {
				return next.Format(dateFormat), nil
			}
			next = next.AddDate(0, 0, 1)
		}
	}

	if parts[0] == "m" {
		if len(parts) != 2 && len(parts) != 3 {
			return "", fmt.Errorf("invalid repeat rule")
		}

		days, err := parseMonthDays(parts[1])
		if err != nil {
			return "", err
		}

		months := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		if len(parts) == 3 {
			months, err = parseNumbers(parts[2], 1, 12)
			if err != nil {
				return "", err
			}
		}

		next, err := nextMonthDate(laterDate(date, now).AddDate(0, 0, 1), days, months)
		if err != nil {
			return "", err
		}
		return next.Format(dateFormat), nil
	}

	return "", fmt.Errorf("invalid repeat rule")
}

// afterNow сообщает, является ли дата строго более поздней,
// сравнивая только календарный год, месяц и день
func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return date.After(now)
}

// laterDate возвращает более позднюю календарную дату между a и b.
// Перед сравнением значения времени суток нормализуются до полуночи
func laterDate(a, b time.Time) time.Time {
	a = time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.Local)
	b = time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.Local)
	if a.After(b) {
		return a
	}
	return b
}

// parseNumbers выполняет синтаксический анализ списка целых чисел, разделенных запятыми, в пределах min и max.
// Возвращает обработанные числа или недопустимую ошибку правила повтора для пустых,
// нечисловых или выходящих за пределы диапазона значений
func parseNumbers(value string, min int, max int) ([]int, error) {
	parts := strings.Split(value, ",")
	numbers := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid repeat rule")
		}

		number, err := strconv.Atoi(part)
		if err != nil || number < min || number > max {
			return nil, fmt.Errorf("invalid repeat rule")
		}
		numbers = append(numbers, number)
	}
	return numbers, nil
}

// parseMonthDays анализирует список значений дней месяца, разделенных запятыми.
// Принимает значения от 1 до 31, -1 для последнего дня месяца и -2 для
// предпоследнего дня, возвращая ошибку для любого другого значения
func parseMonthDays(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid repeat rule")
		}

		day, err := strconv.Atoi(part)
		if err != nil || !validMonthDay(day) {
			return nil, fmt.Errorf("invalid repeat rule")
		}
		days = append(days, day)
	}
	return days, nil
}

// validMonthDay сообщает, разрешен ли день в правиле ежемесячного повторения.
// Допустимые значения от 1 до 31, -1 и -2
func validMonthDay(day int) bool {
	return (day >= 1 && day <= 31) || day == -1 || day == -2
}

// Функция weekdayNumber преобразует обычный день недели в нумерацию по повторяющемуся правилу.
// Возвращает значение от 1 для понедельника до 7 для воскресенья
func weekdayNumber(date time.Time) int {
	if date.Weekday() == time.Sunday {
		return 7
	}
	return int(date.Weekday())
}

// содержит данные о том, нужны ли numbers.
// Используется для небольших обработанных списков правил, где достаточно линейного сканирования
func contains(numbers []int, want int) bool {
	for _, number := range numbers {
		if number == want {
			return true
		}
	}
	return false
}

// nextMonthDate находит первую разрешенную месячную дату в начале или после нее.
// Проверяет подходящие дни только в разрешенных месяцах и возвращает ошибку, если
// в окне ограниченного поиска не удается найти ни одного подходящего
func nextMonthDate(start time.Time, days []int, months []int) (time.Time, error) {
	for i := 0; i < 120; i++ {
		month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, i, 0)
		if !contains(months, int(month.Month())) {
			continue
		}

		candidates := monthCandidates(month, days)
		for _, candidate := range candidates {
			if !candidate.Before(start) {
				return candidate, nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("invalid repeat rule")
}

// monthCandidates создает допустимые даты для запрошенных дней в месяце.
// Невозможные календарные даты, например 31 февраля, пропускаются; значения -1 и -2
// преобразуются в последний и предпоследний день месяца. Возвращаемыми датами являются
// отсортировано от самого раннего к самому последнему
func monthCandidates(month time.Time, days []int) []time.Time {
	lastDay := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	candidates := make([]time.Time, 0, len(days))
	for _, day := range days {
		switch day {
		case -1:
			day = lastDay
		case -2:
			day = lastDay - 1
		}

		if day < 1 || day > lastDay {
			continue
		}
		candidates = append(candidates, time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.Local))
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Before(candidates[j])
	})
	return candidates
}

// nextDateHandler обрабатывает запросы /api/nextdate и возвращает дату в виде обычного текста.
// Принимает параметры запроса now, date и repeat; параметр now является необязательным и используется по умолчанию
// на текущую дату. Неверный ввод возвращается как ответ об ошибке HTTP
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowValue := r.FormValue("now")
	dateValue := r.FormValue("date")
	repeatValue := r.FormValue("repeat")

	now := time.Now()
	var err error
	if nowValue != "" {
		now, err = time.Parse(dateFormat, nowValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, dateValue, repeatValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(next))
}
