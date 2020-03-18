package db

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
	"github.com/pkg/errors"
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

// Transact() executes the provided function inside an atomic transaction.
//
// If the function returns true, the transaction is committed.
// If the function returns false, the transaction is aborted.
// If the function panics, the transaction is aborted and the panic is propagated up the stack.
//
// The passed function is expected to do its own internal error handling (hence it doesn't
// return an error).
//
// Transact() itself returns an error only in the case where the transaction could not be
// initiated, committed, or aborted for some reason.
func (db *DB) Transact(f func() (commit bool)) (bool, error) {
	_, err := db.Exec("BEGIN", nil)
	if err != nil {
		return false, errors.Wrap(err, "Could not BEGIN")
	}
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		_, err = db.Exec("ROLLBACK", nil)
		if err != nil {
			log.Printf("ERROR: Could not rollback transaction: %s", err.Error())
		}
		panic(r)
	}()

	ok := f()

	if !ok {
		_, err = db.Exec("ROLLBACK", nil)
		if err != nil {
			return false, errors.Wrap(err, "Could not ROLLBACK")
		}
		return false, nil
	}

	_, err = db.Exec("COMMIT", nil)
	if err != nil {
		return false, errors.Wrap(err, "Could not COMMIT")
	}
	return true, nil
}

type Params map[string]interface{}

func (db *DB) Exec(sql string, args Params) (*rdsdataservice.ExecuteStatementOutput, error) {
	// TODO: Figure out named params.
	fmt.Printf("Executing: %s\n", sql)

	var params []*rdsdataservice.SqlParameter
	if args != nil {
		for n, v := range args {
			f := rdsdataservice.Field{}
			if v == nil {
				f.SetIsNull(true)
				continue
			}
			switch v := v.(type) {
			case bool:
				f.SetBooleanValue(v)
			case int:
				f.SetLongValue(int64(v))
			case int64:
				f.SetLongValue(v)
			case float32:
				f.SetDoubleValue(float64(v))
			case float64:
				f.SetDoubleValue(v)
			case string:
				f.SetStringValue(v)
			default:
				panic(fmt.Sprintf("Unknown argument type: %#v", v))
			}
			params = append(params, &rdsdataservice.SqlParameter{
				Name:  aws.String(n),
				Value: &f,
			})
		}
	}

	input := &rdsdataservice.ExecuteStatementInput{
		ResourceArn: aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes"),
		SecretArn:   aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia"),
		Sql:         aws.String(sql),
		Parameters:  params,
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
