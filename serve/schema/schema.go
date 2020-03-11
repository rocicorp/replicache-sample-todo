package schema

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
	schemaVersion      = 1
)

var (
	resourceArn = aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes")
	secretArn   = aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia")
)

func dbName() (string, error) {
	n := "REPLICANT_SAMPLE_TODO_ENV"
	env := os.Getenv(n)
	if env == "" {
		return "", fmt.Errorf("Required environment variable %s not found", n)
	} else {
		return fmt.Sprintf("replicache_sample_todo__%s", env), nil
	}
}

func Create() (err error) {
	sess := session.Must(session.NewSession(
		aws.NewConfig().WithRegion(awsRegion).WithCredentials(
			// Have to do this wackiness because not allowed to set AWS env variables in Now for some reason.
			credentials.NewStaticCredentials(
				os.Getenv(awsAccessKeyId),
				os.Getenv(awsSecretAccessKey), ""))))
	svc := rdsdataservice.New(sess)

	name, err := dbName()
	if err != nil {
		return err
	}

	execStatementOutput, err := exec(svc, fmt.Sprintf("SELECT IntVal FROM %s.Meta WHERE Name = 'Version' LIMIT 1", name))
	if err != nil {
		fmt.Printf("ERROR: Invalid database: %s\n", err)
	} else if len(execStatementOutput.Records) == 1 && *(execStatementOutput.Records[0][0].LongValue) == schemaVersion {
		return nil
	}

	statements := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", name),
		fmt.Sprintf("CREATE DATABASE %s", name),
		"START TRANSACTION",
		fmt.Sprintf("CREATE TABLE %s.Meta (Name VARCHAR(16) PRIMARY KEY NOT NULL, IntVal INT)", name),
		fmt.Sprintf("INSERT INTO %s.Meta Values ('Version', %d)", name, schemaVersion),
		fmt.Sprintf("CREATE TABLE %s.User (Id INT PRIMARY KEY)", name),
		"COMMIT",
	}

	for _, s := range statements {
		_, err = exec(svc, s)
		if err != nil {
			fmt.Println("Abandoning transaction")
			return err
		}
	}

	return nil
}

func exec(svc *rdsdataservice.RDSDataService, sql string) (*rdsdataservice.ExecuteStatementOutput, error) {
	fmt.Printf("Executing: %s\n", sql)
	out, err := svc.ExecuteStatement(&rdsdataservice.ExecuteStatementInput{
		ResourceArn: resourceArn,
		SecretArn:   secretArn,
		Sql:         aws.String(sql),
	})
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	return out, err
}
