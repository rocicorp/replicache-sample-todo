package schema

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
)

const (
	aws_access_key_id     = "REPLICANT_AWS_ACCESS_KEY_ID"
	aws_secret_access_key = "REPLICANT_AWS_SECRET_ACCESS_KEY"
	aws_region            = "us-west-2"
)

func Create() error {
	sess := session.Must(session.NewSession(
		aws.NewConfig().WithRegion(aws_region).WithCredentials(
				// Have to do this wackiness because not allowed to set AWS env variables in Now for some reason.
				credentials.NewStaticCredentials(
						os.Getenv(aws_access_key_id),
						os.Getenv(aws_secret_access_key), ""))))
	svc := rdsdataservice.New(sess)
	_, err := svc.ExecuteStatement(&rdsdataservice.ExecuteStatementInput{
		ResourceArn: aws.String("arn:aws:rds:us-west-2:712907626835:cluster:replicache-demo-notes"),
		SecretArn: aws.String("arn:aws:secretsmanager:us-west-2:712907626835:secret:rds-db-credentials/cluster-X5NALMLWZ34K55M5ZZVPN2IYOI/admin-65L3ia"),
		Sql: aws.String("CREATE DATABASE foo"),
	})
	return err
}
