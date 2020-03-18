package schema

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"roci.dev/replicache-sample-todo/serve/db"
)

func Create(db *db.DB, name string) (err error) {
	statements := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", name),
		fmt.Sprintf("CREATE DATABASE %s", name),
		fmt.Sprintf("CREATE TABLE %s.Meta (Name VARCHAR(255) PRIMARY KEY NOT NULL, Value VARCHAR(255) NOT NULL)", name),
		fmt.Sprintf("CREATE TABLE %s.User (Id INT PRIMARY KEY NOT NULL)", name),
		fmt.Sprintf("CREATE TABLE %s.TodoList (Id INT PRIMARY KEY NOT NULL, OwnerUserId INT NOT NULL, FOREIGN KEY (OwnerUserId) REFERENCES %s.User (Id))", name, name),
		fmt.Sprintf("CREATE TABLE %s.Todo (Id INT PRIMARY KEY NOT NULL, TodoListId INT NOT NULL, Title VARCHAR(255) NOT NULL, Complete BIT NOT NULL, SortOrder FLOAT(53) NOT NULL, FOREIGN KEY (TodoListId) REFERENCES %s.TodoList (Id))", name, name),
	}

	schemaHash := sha1.Sum([]byte(strings.Join(statements, "\n")))
	schemaHashStr := hex.EncodeToString(schemaHash[:])

	execStatementOutput, err := db.Exec(fmt.Sprintf("SELECT Value FROM %s.Meta WHERE Name = 'Version' AND Value = '%s' LIMIT 1", name, schemaHashStr), nil)
	if err != nil {
		fmt.Printf("ERROR: Invalid database: %s\n", err)
	} else if len(execStatementOutput.Records) == 1 {
		return nil
	}

	statements = append(statements, fmt.Sprintf("INSERT INTO %s.Meta (Name, Value) VALUES ('Version', '%s')", name, schemaHashStr))

	for _, s := range statements {
		_, err = db.Exec(s, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
