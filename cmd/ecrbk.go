package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	_ "github.com/docker/docker/client"
	"log/slog"
	"os"
)

var awsProfile = flag.String("p", "", "the aws profile to use")
var beVerbose = flag.Bool("v", false, "be verbose")
var repositoryName = flag.String("r", "", "repository")

func main() {
	ctx := context.Background()
	logger := newLogger(*beVerbose)

	flag.Parse()
	err := verifyFlag()
	if err != nil {
		logger.ErrorContext(ctx, "invalid command line arguments were passed", "error", err)
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(*awsProfile))
	if err != nil {
		logger.ErrorContext(ctx, "loading aws configuration file was failed", "error", err)
	}

	cli := ecr.NewFromConfig(cfg)

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
	fmt.Println(token)

	for _, image := range images.ImageIds {
		if tag := image.ImageTag; tag != nil {
			fmt.Println(*tag)
			fmt.Println(*uri + ":" + *tag)
		}

	}
}

func newLogger(verbose bool) *slog.Logger {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

func verifyFlag() error {
	if *repositoryName == "" {
		return fmt.Errorf("-r <repository name> is required")
	}
	return nil
}
