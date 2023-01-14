package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func pgConn() string {
	pgConfig := getConfig()

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgConfig.Postgres.Host, pgConfig.Postgres.Port, pgConfig.Postgres.User, pgConfig.Postgres.Password, pgConfig.Postgres.Dbname)
}

func getPassword(u string) (string, error) {
	var password string
	var db *sql.DB
	pgConfig := getConfig()

	db, err := sql.Open("postgres", pgConn())
	if err != nil {
		return "", err
	}

	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT password FROM %s WHERE username = '%s'", pgConfig.Postgres.Table, u))
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&password)
		if err != nil {
			return "", err
		}
	}

	return password, nil
}

func createUser(u string, p string) error {
	var db *sql.DB
	pgConfig := getConfig()

	db, err := sql.Open("postgres", pgConn())
	if err != nil {
		return err
	}

	defer db.Close()

	insertStmt := fmt.Sprintf("INSERT INTO public.%s(username, password) VALUES('%s', '%s')", pgConfig.Postgres.Table, u, p)
	_, err = db.Exec(insertStmt)
	if err != nil {
		return err
	}

	return nil
}
