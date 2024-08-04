package main

import (
	"log"
	"net/http"
	"os"

	"go_final_project/api"
	"go_final_project/auth"
	"go_final_project/database"
)

func main() {
	// Инициализация базы данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Регистрация маршрутов и передача экземпляра базы данных в обработчики
	http.HandleFunc("/api/signin", auth.SigninHandler(db))

	http.HandleFunc("/api/task", auth.Auth(api.AddTaskHandler(db), db))
	http.HandleFunc("/api/tasks", auth.Auth(api.GetTasksHandler(db), db))
	http.HandleFunc("/api/task/done", auth.Auth(api.MarkTaskDoneHandler(db), db))
	http.HandleFunc("/api/task/delete", auth.Auth(api.DeleteTaskHandler(db), db)) // Регистрация обработчика удаления

	// Регистрация файлового сервера для фронтенда
	webDir := "web"
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	log.Printf("Запуск сервера на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
