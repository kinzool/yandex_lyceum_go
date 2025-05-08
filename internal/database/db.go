package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Expression struct {
	Id     int     `json:"id"`
	Result float64 `json:"result"`
}

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to SQLite database")
	createTables(context.TODO(), db)
	return db, nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	const usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT UNIQUE,
		password TEXT
	);`

	const expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		expression TEXT NOT NULL,
		result FLOAT,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`
	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}
	log.Println("Successfully added a tables to SQlite database")
	return nil
}

func InsertUsers(ctx context.Context, login, password string, db *sql.DB) error {
	var exists bool
	err := db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)",
		login,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf(`{"error": "Failed to check user existence: %w"}`, err)
	}

	if exists {
		return fmt.Errorf("user '%s' already exists", login)
	}
	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	_, err = db.ExecContext(ctx, q, login, password)
	if err != nil {
		return errors.New(`{"error": "Something went wrong"}`)
	}
	return nil
}

func GetUserID(ctx context.Context, login string, db *sql.DB) int {
	var id int
	var q = `SELECT id FROM users WHERE login = $1`
	err := db.QueryRowContext(ctx, q, login).Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

func IsAuth(ctx context.Context, login, password string, db *sql.DB) bool {
	var storedPassword string
	var q = `SELECT password FROM users WHERE login = $1`
	err := db.QueryRowContext(ctx, q, login).Scan(&storedPassword)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	return err == nil
}

func AddExpression(ctx context.Context, user_id int, expression string, db *sql.DB) (int, error) {
	var q = `INSERT INTO expressions (user_id, expression) values ($1, $2)`
	result, err := db.ExecContext(ctx, q, user_id, expression)
	if err != nil {
		return 0, errors.New(`{"error": "Something went wrong"}`)
	}
	id, _ := result.LastInsertId()
	log.Printf("Expression '%s' successfully added", expression)
	return int(id), nil
}

func AddAnswer(ctx context.Context, id int, result float64, db *sql.DB) error {
	var q = `UPDATE expressions
	SET result = $1
	WHERE id = $2`
	_, err := db.ExecContext(ctx, q, result, id)
	if err != nil {
		return errors.New(`{"error": "Something went wrong"}`)
	}
	return nil
}

func GetExpressions(user_id int, db *sql.DB) ([]Expression, error) {
	var answ []Expression
	var q = `SELECT id, result FROM expressions
	WHERE user_id = $1`
	rows, err := db.Query(q, user_id)
	if err != nil {
		return nil, errors.New(`{"error": "Something went wrong"}`)
	}
	for rows.Next() {
		var id int
		var result float64
		if err := rows.Scan(&id, &result); err != nil {
			return nil, errors.New(`{"error": "Something went wrong"}`)
		}
		answ = append(answ, Expression{Id: id, Result: result})
	}
	return answ, nil
}

func GetExpressionByID(ctx context.Context, user_id, id int, db *sql.DB) (Expression, error) {
	var answ_id int
	var answ_result float64
	var q = `SELECT id, result FROM expressions
	WHERE user_id = $1 AND id = $2`
	err := db.QueryRowContext(ctx, q, user_id, id).Scan(&answ_id, &answ_result)
	if err != nil {
		return Expression{}, errors.New(`{"error": "Empty answer"}`)
	}
	return Expression{Id: answ_id, Result: answ_result}, nil
}
