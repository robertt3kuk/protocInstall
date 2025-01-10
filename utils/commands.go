package utils

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// RunCommand запускает команду и возвращает ошибку, если она возникает.
// Перед выполнением команда и её аргументы логируются.
func RunCommand(command string, args ...string) error {
	log.Printf("Running command: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s with args %v: %w", command, args, err)
	}

	return nil
}

// RunCommandWithOutput запускает команду и возвращает ошибку, если она возникает, иначе возвращает стандартный вывод.
// Перед выполнением команда и её аргументы логируются.
func RunCommandWithOutput(command string, args ...string) ([]byte, error) {
	log.Printf("Running command: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get command output: %w", err)
	}

	return output, nil
}

// CheckGitlabEnvVariables проверяет наличие необходимых переменных окружения.
func CheckGitlabEnvVariables(cmd *cobra.Command, args []string) error {
	projectGitlabToken := os.Getenv("PROJECT_GITLAB_TOKEN")
	if projectGitlabToken == "" {
		gitlabLogin := os.Getenv("GITLAB_LOGIN")
		if gitlabLogin == "" {
			return errors.New("GITLAB_LOGIN environment variable must be set")
		}
		gitlabToken := os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			return errors.New("GITLAB_TOKEN environment variable must be set")
		}
	}
	goPrivate := os.Getenv("GOPRIVATE")
	if goPrivate == "" {
		return errors.New("GOPRIVATE environment variable must be set")
	}
	return nil
}

// IsContainerRunning проверяет, запущен ли указанный контейнер.
func IsContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("ошибка при выполнении docker ps: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, c := range containers {
		if c == containerName {
			return true, nil
		}
	}

	return false, nil
}
