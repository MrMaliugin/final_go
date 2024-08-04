package database

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sqlx.DB, error) {
	appPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")

	_, err = os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	if install {
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT,
            title TEXT,
            comment TEXT,
            repeat TEXT
        )`)
		if err != nil {
			return nil, err
		}
		log.Println("Таблица scheduler создана или уже существует.")

		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date)`)
		if err != nil {
			return nil, err
		}
		log.Println("Индекс idx_date создан или уже существует.")
	}

	return db, nil
}

var (
	tasks = make(map[string]string)
	mu    sync.Mutex
)

func DeleteTask(db *sqlx.DB, id string) error {
	mu.Lock()
	defer mu.Unlock()

	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("task not found")
	}

	return nil
}
