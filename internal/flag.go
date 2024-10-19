package internal

import "fmt"

func VerifyFlag(repositoryName string) error {
	if repositoryName == "" {
		return fmt.Errorf("-r <repository name> is required")
	}
	return nil
}
