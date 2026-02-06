package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// AWSConfig configures the AWS Secrets Manager provider.
type AWSConfig struct {
	// Region is the AWS region (e.g., "us-east-1").
	Region string

	// Endpoint is an optional custom endpoint for local testing.
	Endpoint string

	// Prefix is prepended to all secret names.
	Prefix string
}

// AWSProvider implements secrets.Provider for AWS Secrets Manager.
type AWSProvider struct {
	client *secretsmanager.Client
	prefix string
}

// NewAWSProvider creates a new AWS Secrets Manager provider.
func NewAWSProvider(ctx context.Context, cfg AWSConfig) (*AWSProvider, error) {
	var opts []func(*config.LoadOptions) error

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOpts := []func(*secretsmanager.Options){}
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	client := secretsmanager.NewFromConfig(awsCfg, clientOpts...)

	return &AWSProvider{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

// Get retrieves a secret from AWS Secrets Manager.
func (p *AWSProvider) Get(ctx context.Context, key string) (*Secret, error) {
	secretName := p.prefixKey(key)

	output, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %q: %w", key, err)
	}

	var value string
	if output.SecretString != nil {
		value = *output.SecretString
	}

	return &Secret{
		Key:     key,
		Value:   value,
		Version: aws.ToString(output.VersionId),
	}, nil
}

// Put stores a secret in AWS Secrets Manager.
func (p *AWSProvider) Put(ctx context.Context, key string, value string, opts ...PutOption) error {
	options := &PutOptions{}
	for _, opt := range opts {
		opt(options)
	}

	secretName := p.prefixKey(key)

	// Try to update existing secret first
	_, err := p.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(value),
	})

	if err != nil {
		// Secret doesn't exist, create it
		tags := []secretsManagerTag{}
		for k, v := range options.Metadata {
			tags = append(tags, secretsManagerTag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}

		createInput := &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(value),
		}

		if len(tags) > 0 {
			createInput.Tags = convertTags(tags)
		}

		_, err = p.client.CreateSecret(ctx, createInput)
		if err != nil {
			return fmt.Errorf("failed to create secret %q: %w", key, err)
		}
	}

	return nil
}

// Delete removes a secret from AWS Secrets Manager.
func (p *AWSProvider) Delete(ctx context.Context, key string) error {
	secretName := p.prefixKey(key)

	_, err := p.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret %q: %w", key, err)
	}

	return nil
}

// List returns all secret names matching the prefix.
func (p *AWSProvider) List(ctx context.Context, prefix string) ([]string, error) {
	fullPrefix := p.prefixKey(prefix)

	var keys []string
	var nextToken *string

	for {
		output, err := p.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
			NextToken: nextToken,
			Filters: []types.Filter{
				{
					Key:    types.FilterNameStringTypeName,
					Values: []string{fullPrefix},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range output.SecretList {
			if secret.Name != nil {
				// Remove prefix from returned keys
				name := *secret.Name
				if len(p.prefix) > 0 && len(name) > len(p.prefix)+1 {
					name = name[len(p.prefix)+1:]
				}
				keys = append(keys, name)
			}
		}

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return keys, nil
}

// Health checks AWS Secrets Manager connectivity.
func (p *AWSProvider) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to list secrets with a limit of 1 to verify connectivity
	_, err := p.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("AWS Secrets Manager health check failed: %w", err)
	}

	return nil
}

func (p *AWSProvider) prefixKey(key string) string {
	if p.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s/%s", p.prefix, key)
}

// GetJSON retrieves a secret and unmarshals it as JSON.
func (p *AWSProvider) GetJSON(ctx context.Context, key string, v interface{}) error {
	secret, err := p.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(secret.Value), v); err != nil {
		return fmt.Errorf("failed to unmarshal secret %q: %w", key, err)
	}

	return nil
}

// Helper types for AWS SDK
type secretsManagerTag struct {
	Key   *string
	Value *string
}

func convertTags(tags []secretsManagerTag) []types.Tag {
	result := make([]types.Tag, len(tags))
	for i, t := range tags {
		result[i] = types.Tag{
			Key:   t.Key,
			Value: t.Value,
		}
	}
	return result
}
