package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/ryotaro612/ecr-backup/internal"
	"os"
)

var awsProfile = flag.String("p", "", "the aws profile to use")
var beVerbose = flag.Bool("v", false, "be verbose")
var repositoryName = flag.String("r", "", "repository")

func main() {
	ctx := context.Background()
	logger := internal.NewLogger(*beVerbose)

	flag.Parse()
	err := internal.VerifyFlag(*repositoryName)

	if err != nil {
		logger.ErrorContext(ctx, "invalid command line arguments were passed", "error", err)
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(*awsProfile))
	if err != nil {
		logger.ErrorContext(ctx, "loading aws configuration file was failed", "error", err)
	}

	cli := ecr.NewFromConfig(cfg)

	regi, err := cli.DescribeRegistry(ctx, &ecr.DescribeRegistryInput{})
	if err != nil {
		logger.ErrorContext(ctx, "inquiring information of the registry was failed", "error", err)
		os.Exit(1)
	}
	fmt.Println(*regi.RegistryId)
	fmt.Println(cfg.Region)
	repo, err := cli.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{*repositoryName},
	})
	if err != nil {
		logger.ErrorContext(ctx, "inquiring information of the repository was failed", "error", err, "repository", *repositoryName)
		os.Exit(1)
	}
	if len(repo.Repositories) != 1 {
		logger.ErrorContext(ctx, "the repository was not found", "repository", repositoryName)
		os.Exit(1)
	}

	uri := repo.Repositories[0].RepositoryUri

	images, err := cli.ListImages(ctx, &ecr.ListImagesInput{RepositoryName: repositoryName})
	if err != nil {
		logger.ErrorContext(ctx, "listing the images was failed", "error", err, "repository", repositoryName)
	}

	token, err := cli.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		logger.ErrorContext(ctx, "getting the authorization token was failed", "error", err)
		os.Exit(1)
	}
	fmt.Println(*token.AuthorizationData[0].AuthorizationToken)
	dec, err := base64.StdEncoding.DecodeString(*token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		logger.ErrorContext(ctx, "decoding the authorization token was failed", "error", err)
		os.Exit(1)
	}
	for _, image := range images.ImageIds {
		if tag := image.ImageTag; tag != nil {
			fmt.Println(*tag)
			fmt.Println(*uri + ":" + *tag)
		}

	}
	fmt.Println(string(dec[4:]))
	docker, err := client.NewClientWithOpts(client.WithVersion("1.45"))
	if err != nil {
		logger.ErrorContext(ctx, "creating a docker client was failed", "error", err)
		os.Exit(1)
	}

	body, err := docker.RegistryLogin(ctx, registry.AuthConfig{
		Username:      "AWS",
		Password:      string(dec)[4:],
		ServerAddress: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", *regi.RegistryId, cfg.Region),
	})
	if err != nil {
		logger.ErrorContext(ctx, "logging in to the registry was failed", "error", err)
	}
	fmt.Println(body.Status)
}
