package api

// Task представляет задачу
type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

// Response представляет ответ сервера
type Response struct {
	ID       string `json:"id,omitempty"`
	Error    string `json:"error,omitempty"`
	Tasks    []Task `json:"tasks,omitempty"`
	NextDate string `json:"nextDate,omitempty"`
}
