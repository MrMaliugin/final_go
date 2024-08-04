package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go_final_project/database"
	"io"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

type Response struct {
	ID       string `json:"id,omitempty"`
	Error    string `json:"error,omitempty"`
	Tasks    []Task `json:"tasks,omitempty"`
	NextDate string `json:"nextDate,omitempty"`
}

func AddTaskHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task Task

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %s", err)
			http.Error(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
			return
		}
		log.Printf("Received body: %s", body)

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		err = json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			log.Printf("JSON Decode Error: %s", err)
			http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
			return
		}

		if task.Title == "" {
			http.Error(w, "Не указан заголовок задачи", http.StatusBadRequest)
			return
		}

		const layout = "20060102"
		now := time.Now()

		if task.Date == "" {
			task.Date = now.Format(layout)
		} else {
			parsedDate, err := time.Parse(layout, task.Date)
			if err != nil {
				http.Error(w, "Дата указана в неправильном формате", http.StatusBadRequest)
				return
			}
			if parsedDate.Before(now) {
				if task.Repeat == "" {
					task.Date = now.Format(layout)
				} else {
					nextDate, err := NextDate(now, task.Date, task.Repeat)
					if err != nil {
						http.Error(w, "Ошибка вычисления следующей даты: "+err.Error(), http.StatusBadRequest)
						return
					}
					task.Date = nextDate
				}
			}
		}

		res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
			task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, "Ошибка добавления задачи в базу данных", http.StatusInternalServerError)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Ошибка получения ID новой задачи", http.StatusInternalServerError)
			return
		}

		response := Response{ID: fmt.Sprintf("%d", id)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetTasksHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Queryx("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50")
		if err != nil {
			http.Error(w, "Ошибка выборки задач из базы данных", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var task Task
			if err := rows.StructScan(&task); err != nil {
				http.Error(w, "Ошибка сканирования задачи", http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, task)
		}

		if tasks == nil {
			tasks = []Task{}
		}

		response := Response{Tasks: tasks}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func MarkTaskDoneHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		var task Task
		err := db.QueryRowx("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).StructScan(&task)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			} else {
				http.Error(w, `{"error": "Ошибка при получении задачи"}`, http.StatusInternalServerError)
			}
			return
		}

		if task.Repeat == "" {
			_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
			if err != nil {
				http.Error(w, `{"error": "Ошибка при удалении задачи"}`, http.StatusInternalServerError)
				return
			}
		} else {
			now := time.Now()
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error": "Ошибка при вычислении следующей даты"}`, http.StatusInternalServerError)
				return
			}

			_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
			if err != nil {
				http.Error(w, `{"error": "Ошибка при обновлении задачи"}`, http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{})
	}
}

func DeleteTask(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		log.Printf("Attempting to delete task with id: %s", id)

		err := database.DeleteTask(db, id)
		if err != nil {
			log.Printf("Error deleting task with id %s: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Task with id %s successfully deleted", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
