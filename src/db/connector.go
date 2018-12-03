package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// ConnectorConfig model
type ConnectorConfig struct {
	Endpoint  string
	Key       string
	Secret    string
	TableName string
	Region    string
}

// Connector model
type Connector struct {
	Session *session.Session
	Service *dynamodb.DynamoDB
}

// NewConnector creates a new db connector
func NewConnector(c ConnectorConfig) (*Connector, error) {
	connector := &Connector{}

	config := &aws.Config{
		Region: aws.String(c.Region),
	}

	if c.Endpoint != "" {
		config.Endpoint = aws.String(c.Endpoint)
	}

	if c.Key != "" && c.Secret != "" {
		config.Credentials = credentials.NewStaticCredentials(c.Key, c.Secret, "")
	}

	sess, err := session.NewSession(config)

	if err != nil {
		return nil, err
	}

	connector.Session = sess
	connector.Service = dynamodb.New(sess)

	return connector, nil
}
