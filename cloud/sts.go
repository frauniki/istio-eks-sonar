package cloud

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func assumeRole(cfg *aws.Config, roleArn string) {
	stsClient := sts.NewFromConfig(*cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleArn)
	cfg.Credentials = aws.NewCredentialsCache(provider)
}
