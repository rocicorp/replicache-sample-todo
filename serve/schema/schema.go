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
	_, err := svc.ExecuteSql(&rdsdataservice.ExecuteSqlInput{
		DbClusterOrInstanceArn: aws.String("replicache-demo-notes"),
		SqlStatements: aws.String("CREATE TABLE User (ID INT AUTO_INCREMENT PRIMARY KEY)"),
	})
	return err
}
