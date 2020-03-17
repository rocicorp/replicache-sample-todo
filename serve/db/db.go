package db

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
)

const (
	awsAccessKeyId     = "REPLICANT_AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "REPLICANT_AWS_SECRET_ACCESS_KEY"
	awsRegion          = "us-west-2"
)

type DB struct {
	name string
	svc  *rdsdataservice.RDSDataService
}

func New() *DB {
	cfg := aws.NewConfig().WithRegion(awsRegion)
	envKey := os.Getenv(awsAccessKeyId)
	if envKey != "" {
		// Have to do this wackiness because not allowed to set AWS env variables in Now for some reason.
		cfg = cfg.WithCredentials(credentials.NewStaticCredentials(
			envKey,
			os.Getenv(awsSecretAccessKey), ""))
	}
	sess := session.Must(session.NewSession(cfg))
	return &DB{
		svc: rdsdataservice.New(sess),
	}
}

func (db *DB) Using() string {
	return db.name
}

func (db *DB) Use(dbName string) {
	db.name = dbName
}

func (db *DB) Begin() error {
	_, err := db.Exec("BEGIN")
	return err
}

func (db *DB) Commit() error {
	_, err := db.Exec("COMMIT")
	return err
}

func (db *DB) Exec(sql string) (*rdsdataservice.ExecuteStatementOutput, error) {
	// TODO: Figure out named params.
	fmt.Printf("Executing: %s\n", sql)
	input := &rdsdataservice.ExecuteStatementInput{
		ResourceArn: aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes"),
		SecretArn:   aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia"),
		Sql:         aws.String(sql),
	}
	if db.name != "" {
		input.Database = aws.String(db.name)
	}
	out, err := db.svc.ExecuteStatement(input)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	return out, err
}
