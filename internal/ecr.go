package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"log/slog"
)

// NewClient creates a new client for ECR.
func NewClient(l *slog.Logger, cfg aws.Config) client {
	cli := ecr.NewFromConfig(cfg)
	return client{cli: cli, cfg: cfg}
}

// ServerAddress returns the address that docker login command requires.
func (c client) ServerAddress(ctx context.Context) (string, error) {
	re, err := c.cli.DescribeRegistry(ctx, &ecr.DescribeRegistryInput{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", *re.RegistryId, c.cfg.Region), nil
}

// ListImages lists images in the repository.
// The image is in the format of <repository uri>:<tag>.
func (c client) ListImages(ctx context.Context, repositoryName string) ([]string, error) {
	repos, err := c.cli.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repositoryName},
	})

	if err != nil {
		return nil, err
	}
	if len(repos.Repositories) != 1 {
		return nil, fmt.Errorf("more than one repositoies were found by '%s'", repositoryName)
	}
	uri := repos.Repositories[0].RepositoryUri

	imgs, err := c.cli.ListImages(ctx, &ecr.ListImagesInput{RepositoryName: &repositoryName})
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, img := range imgs.ImageIds {
		res = append(res, fmt.Sprintf("%s:%s", uri, img.ImageTag))
	}

	return res, nil
}

// GetAuthorizationToken returns a password that `docker login --username AWS --password-stdin <serveraddress>` accepts.
func (c client) GetAuthorizationToken(ctx context.Context) (string, error) {
	token, err := c.cli.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}
	dec, err := base64.StdEncoding.DecodeString(*token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", err
	}
	// AWS:<password>
	return string(dec)[4:], nil
}

type client struct {
	cli *ecr.Client
	cfg aws.Config
}
