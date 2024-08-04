package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"

	"go_final_project/api"
	"go_final_project/auth"
	"go_final_project/database"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/signin", auth.SigninHandler(db)).Methods("POST")

	r.HandleFunc("/api/task", auth.Auth(api.AddTaskHandler(db), db)).Methods("POST")
	r.HandleFunc("/api/tasks", auth.Auth(api.GetTasksHandler(db), db)).Methods("GET")
	r.HandleFunc("/api/task/done", auth.Auth(api.MarkTaskDoneHandler(db), db)).Methods("POST")
	r.HandleFunc("/api/task/delete/{id}", auth.Auth(api.DeleteTask(db), db)).Methods("DELETE")

	webDir := "web"
	fs := http.FileServer(http.Dir(webDir))
	r.PathPrefix("/").Handler(fs)

	log.Printf("Запуск сервера на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
