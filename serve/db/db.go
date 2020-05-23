package db

import (
	"fmt"
	"log"
	"os"
	"reflect"

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

// Transact executes a function inside an atomic transaction.
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
func (db *DB) Transact(f func(ExecFunc) (commit bool)) (bool, error) {
	rArn := aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes")
	sArn := aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia")
	input := &rdsdataservice.BeginTransactionInput{
		ResourceArn: rArn,
		SecretArn:   sArn,
	}
	if db.name != "" {
		input.Database = aws.String(db.name)
	}
	out, err := db.svc.BeginTransaction(input)
	if err != nil {
		return false, fmt.Errorf("could not BEGIN: %w", err)
	}
	rollbackInput := rdsdataservice.RollbackTransactionInput{
		ResourceArn:   rArn,
		SecretArn:     sArn,
		TransactionId: out.TransactionId,
	}
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		log.Printf("caught panic: %#v", r)
		_, err := db.svc.RollbackTransaction(&rollbackInput)
		if err != nil {
			log.Printf("ERROR: Could not rollback transaction: %v", err)
		}
		panic(r)
	}()

	execFunc := func(sql string, args Params) (*rdsdataservice.ExecuteStatementOutput, error) {
		return db.ExecInTransaction(out.TransactionId, sql, args)
	}
	ok := f(execFunc)

	if !ok {
		_, err = db.svc.RollbackTransaction(&rollbackInput)
		if err != nil {
			return false, fmt.Errorf("could not ROLLBACK: %w", err)
		}
		return false, nil
	}

	cInput := &rdsdataservice.CommitTransactionInput{
		ResourceArn:   rArn,
		SecretArn:     sArn,
		TransactionId: out.TransactionId,
	}
	_, err = db.svc.CommitTransaction(cInput)
	if err != nil {
		return false, fmt.Errorf("could not COMMIT: %w", err)
	}
	return true, nil
}

type Params map[string]interface{}

type ExecFunc func(sql string, args Params) (*rdsdataservice.ExecuteStatementOutput, error)

func (db *DB) Exec(sql string, args Params) (*rdsdataservice.ExecuteStatementOutput, error) {
	return db.ExecInTransaction(nil, sql, args)
}

func (db *DB) ExecInTransaction(transactionID *string, sql string, args Params) (*rdsdataservice.ExecuteStatementOutput, error) {
	fmt.Printf("Executing: %s\n", sql)

	var params []*rdsdataservice.SqlParameter
	if args != nil {
		for n, arg := range args {
			f := argToField(arg)
			params = append(params, &rdsdataservice.SqlParameter{
				Name:  aws.String(n),
				Value: &f,
			})
		}
	}

	input := &rdsdataservice.ExecuteStatementInput{
		ResourceArn:   aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes"),
		SecretArn:     aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia"),
		Sql:           aws.String(sql),
		Parameters:    params,
		TransactionId: transactionID,
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

func argToField(arg interface{}) (f rdsdataservice.Field) {
	rv := reflect.ValueOf(arg)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			f.SetIsNull(true)
			return f
		}
		arg = reflect.Indirect(rv).Interface()
	}
	switch arg := arg.(type) {
	case bool:
		f.SetBooleanValue(arg)
	case int:
		f.SetLongValue(int64(arg))
	case int64:
		f.SetLongValue(arg)
	case float32:
		f.SetDoubleValue(float64(arg))
	case float64:
		f.SetDoubleValue(arg)
	case string:
		f.SetStringValue(arg)
	default:
		panic(fmt.Sprintf("Unknown argument type: %#v", arg))
	}
	return f
}
