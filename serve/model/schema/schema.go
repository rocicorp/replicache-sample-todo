package schema

import (
	"fmt"

	"roci.dev/replicache-sample-todo/serve/db"
)

const (
	schemaVersion = 4
)

func Create(db *db.DB, name string) (err error) {
	execStatementOutput, err := db.Exec(fmt.Sprintf("SELECT IntVal FROM %s.Meta WHERE Name = 'Version' LIMIT 1", name))
	if err != nil {
		fmt.Printf("ERROR: Invalid database: %s\n", err)
	} else if len(execStatementOutput.Records) == 1 && *(execStatementOutput.Records[0][0].LongValue) == schemaVersion {
		return nil
	}

	statements := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", name),
		fmt.Sprintf("CREATE DATABASE %s", name),
		fmt.Sprintf("CREATE TABLE %s.Meta (Name VARCHAR(16) PRIMARY KEY NOT NULL, IntVal INT)", name),
		fmt.Sprintf("INSERT INTO %s.Meta Values ('Version', %d)", name, schemaVersion),
		fmt.Sprintf("CREATE TABLE %s.User (Id INT PRIMARY KEY NOT NULL)", name),
		fmt.Sprintf("CREATE TABLE %s.TodoList (Id INT PRIMARY KEY NOT NULL, OwnerUserId INT NOT NULL, FOREIGN KEY (OwnerUserId) REFERENCES %s.User (Id))", name, name),
		fmt.Sprintf("CREATE TABLE %s.Todo (Id INT PRIMARY KEY NOT NULL, TodoListId INT NOT NULL, Title VARCHAR(255) NOT NULL, Complete BIT NOT NULL, SortOrder FLOAT(53) NOT NULL, FOREIGN KEY (TodoListId) REFERENCES %s.TodoList (Id))", name, name),
	}

	for _, s := range statements {
		_, err = db.Exec(s)
		if err != nil {
			return err
		}
	}

	return nil
}
