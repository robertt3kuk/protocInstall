package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var version = "dev"

// GetModuleName возвращает имя модуля, указанное в файле go.mod.
// Функция открывает файл go.mod, читает его построчно и ищет строку, начинающуюся с "module ".
// Если строка с именем модуля найдена, функция возвращает его; в противном случае возвращается ошибка.
func GetModuleName(dir string) (string, error) {
	goModPath := filepath.Join(dir, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", errors.New("module name not found in go.mod")
}

// RunDockerCompose запускает docker-compose с указанными аргументами.
func RunDockerCompose(envFiles []string, composeFile string, args ...string) error {
	// Предварительно выделяем память для cmdArgs, исходя из количества env-файлов и других аргументов
	cmdArgs := make([]string, 0, len(envFiles)*2+len(args)+3) //nolint:mnd // len(envFiles)*2 для --env-file и его значения, +2 для -f и composeFile
	cmdArgs = append(cmdArgs, "compose")

	// Добавляем флаги --env-file и соответствующие файлы
	for _, envFile := range envFiles {
		cmdArgs = append(cmdArgs, "--env-file", envFile)
	}

	// Добавляем флаг -f и имя compose-файла
	cmdArgs = append(cmdArgs, "-f", composeFile)

	// Добавляем оставшиеся аргументы
	cmdArgs = append(cmdArgs, args...)

	// Создаем команду для запуска docker-compose
	cmd := exec.Command("docker", cmdArgs...)

	// Запускаем команду и проверяем ошибки
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	err := cmd.Run()
	if err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// RunDockerComposeWithCustomEnv запускает docker-compose один сервис с env переменными из map[string]string.
func RunDockerComposeOneService(env map[string]string, composeFile string, serviceName string) error {
	// Pre-allocate memory for cmdArgs, considering the compose file and other arguments
	cmdArgs := []string{"compose", "-f", composeFile, "up", serviceName, "-d", "--build"}

	// Create the environment slice to pass to the command
	envVars := os.Environ() // Get current environment variables
	for key, value := range env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	// Create the command to run docker-compose
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Env = envVars // Pass the environment variables

	// Attach the command's output to logs
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	// Run the command and handle errors
	err := cmd.Run()
	if err != nil {
		return err // nolint:wrapcheck
	}

	return nil
}

// StopDockerComposeOneService останавливает docker-compose один сервис.
func StopDockerComposeOneService(serviceName string) {
	// Создаем команду для остановки docker-compose
	cmd := exec.Command("docker", "rm", "-f", serviceName)

	// Запускаем команду и проверяем ошибки
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to stop docker-compose: %v", err)

		return
	}
}

// GenerateUniquePort генерирует уникальный порт.
func GenerateUniquePort() int {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
	var port int
	for {
		port = rand.Intn(64511) + 1024 //nolint:gosec,mnd

		if !isPortInUse(port) {
			break
		}
	}

	return port
}

// IsPortInUse проверяет, используется ли данный порт.
func isPortInUse(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true // Порт используется
	}
	defer listener.Close()

	return false
}

// GetLastSegment Функция для извлечения последнего значения после /.
func GetLastSegment(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// GetSchemaName извлекает имя схемы из ссылки $ref в объекте OpenAPI3.
func GetSchemaName(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil || schemaRef.Ref == "" {
		return ""
	}

	// Извлекаем имя структуры из $ref, например, "#/components/schemas/YourSchemaName"
	parts := strings.Split(schemaRef.Ref, "/")

	return parts[len(parts)-1]
}

// IsAlpine проверяет, работает ли система на базе дистрибутива Alpine Linux.
func IsAlpine() bool {
	output, err := exec.Command("cat", "/etc/os-release").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Alpine")
}

// GetGitBranchName returns the current Git branch name.
func GetGitBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "unknown", err
	}

	// Trim whitespace characters like newline
	branchName := strings.TrimSpace(out.String())
	return branchName, nil
}

func GetFieldPackageMapFromFile(filePath string, structName string) (map[string]string, error) {
	fieldMap := make(map[string]string)

	// Открываем файл
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	// Создаем FileSet для управления позициями в исходнике
	fset := token.NewFileSet()

	// Парсим файл и создаем AST
	fileNode, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("cannot parse Go file: %v", err)
	}

	// Проходим по всем объявлениям в файле
	ast.Inspect(fileNode, func(n ast.Node) bool {
		// Ищем объявление типа (в данном случае структуры)
		ts, ok := n.(*ast.TypeSpec)
		if ok && ts.Name.Name == structName { //nolint:nestif
			// Ищем поля структуры
			st, ok := ts.Type.(*ast.StructType)
			if ok {
				for _, field := range st.Fields.List {
					var packageName string

					// Проверяем тип поля: обычный тип или указатель
					switch t := field.Type.(type) {
					case *ast.StarExpr: // Указатель на тип
						if sel, ok := t.X.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok {
								packageName = ident.Name
							}
						}
					case *ast.SelectorExpr: // Обычный тип без указателя
						if ident, ok := t.X.(*ast.Ident); ok {
							packageName = ident.Name
						}
					}

					// Если имя пакета получено, ищем его в импортированных пакетах
					if packageName != "" {
						for _, imp := range fileNode.Imports {
							// Извлекаем путь импорта
							importPath := strings.Trim(imp.Path.Value, `"`)

							// Если псевдоним у импорта явно указан, используем его. Если нет, используем последний элемент пути как имя пакета.
							var importName string
							if imp.Name != nil {
								importName = imp.Name.Name // Используем указанный псевдоним
							} else {
								// Используем последний элемент пути как имя пакета, если псевдоним не задан
								importName = filepath.Base(importPath)
							}

							// Проверяем, совпадает ли имя пакета (или псевдоним) с именем, использованным в селекторе
							if importName == packageName {
								// Проходим по полям структуры и записываем их в map
								for _, fieldName := range field.Names {
									fieldMap[fieldName.Name] = importPath
								}
								break
							}
						}
					}
				}
			}
		}
		return true
	})

	return fieldMap, nil
}
