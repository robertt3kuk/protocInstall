package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

type BrewProtocVersion struct {
	Version BrewProtocStable `json:"versions"`
}

type BrewProtocStable struct {
	Stable string `json:"stable"`
}

// GetVersion возвращает версию сборки приложения.
// Если версия не соответствует "(devel)", возвращается основная версия сборки,
// в противном случае возвращается значение глобальной переменной `version`.
func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return version
}

// Slugify заменяет все специальные символы на дефисы, превращая строку в слаг.
// Используется для формирования урлов или корректных версий openapi по SemVer.
func Slugify(str string) string {
	// Приводим к нижнему регистру
	str = strings.ToLower(str)

	// Заменяем все символы, кроме букв, цифр и дефисов, на дефис
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	str = reg.ReplaceAllString(str, "-")

	// Удаляем начальные и конечные дефисы
	str = strings.Trim(str, "-")

	// Заменяем множественные дефисы на один
	reg = regexp.MustCompile(`-+`)
	str = reg.ReplaceAllString(str, "-")

	return str
}

// BumpVersion повышает семантическую версию.
// bumpType должен быть major, minor или patch.
func BumpVersion(version, bumpType string) (string, error) {
	// Отбрасываем часть после дефиса
	versionParts := strings.SplitN(version, "-", 2)
	version = versionParts[0]

	// Разделяем версию на части major.minor.patch
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("неверный формат версии: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("неверный формат major версии: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("неверный формат minor версии: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("неверный формат patch версии: %s", parts[2])
	}

	// Повышаем версию в зависимости от переданного аргумента
	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("неизвестный тип повышения версии: %s", bumpType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func GetStableProtocVersion() (string, error) {
	platform := runtime.GOOS
	switch platform {
	case "darwin":
		resp, err := http.Get("https://formulae.brew.sh/api/formula/protobuf.json")
		if err != nil {
			return "", fmt.Errorf("failed to get stable protoc version: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return "", err
		}
		var brewProtocVersion BrewProtocVersion
		if err := json.Unmarshal(body, &brewProtocVersion); err != nil {
			return "", fmt.Errorf("failed to unmarshal protoc version: %w", err)
		}

		return brewProtocVersion.Version.Stable, nil
	case "linux":
		protocVersion, err := checkGitHubProtobufVersion()
		if err != nil {
			return "", fmt.Errorf("failed to check github protobuf version: %w", err)
		}

		return protocVersion, nil
	default:
		return "", fmt.Errorf("не поддерживается платформа: %s", platform)
	}
}

func checkGitHubProtobufVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/protocolbuffers/protobuf/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest protobuf release: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var release struct {
		TagName string `json:"tag_name"` // nolint:tagliatelle
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	version := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	return version, nil
}

func InstallProtoBufLinuxGithub(version string) error {
	var architecture string
	url := `https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-linux-%s.zip`
	switch runtime.GOARCH {
	case "amd64":
		architecture = "x86_64"
	case "arm64":
		architecture = "aarch_64"
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	_, err := RunCommandWithOutput("curl", "-L", "-o", "protoc.zip", fmt.Sprintf(url, version, version, architecture))
	if err != nil {
		return fmt.Errorf("failed to download protoc: %w", err)
	}
	defer os.Remove("protoc.zip")

	if err := RunCommand("sudo", "unzip", "-o", "protoc.zip", "-d", "/usr/local"); err != nil {
		return fmt.Errorf("failed to unzip protoc: %w", err)
	}

	if err := RunCommand("sudo", "chmod", "+x", "/usr/local/bin/protoc"); err != nil {
		return fmt.Errorf("failed to chmod protoc: %w", err)
	}

	return nil
}

func DetectLinuxDistribution() (string, string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", "", fmt.Errorf("failed to open /etc/os-release: %w", err)
	}
	defer file.Close()

	var id, versionID, idLike string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		} else if strings.HasPrefix(line, "ID_LIKE=") {
			idLike = strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), `"`)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading /etc/os-release: %w", err)
	}
	if id == "" {
		return "", "", fmt.Errorf("could not determine Linux distribution")
	}

	// Handle SUSE variants
	if strings.Contains(strings.ToLower(id), "suse") || strings.Contains(strings.ToLower(idLike), "suse") {
		return "suse", versionID, nil
	}

	return strings.ToLower(id), versionID, nil
}

func RemovePackageManagerProtobuf(distro string) error {
	switch distro {
	case "ubuntu", "debian":
		// Check if installed
		if _, err := RunCommandWithOutput("dpkg", "-l", "protobuf-compiler"); err == nil {
			cmd := []string{"sudo", "apt-get", "remove", "-y", "protobuf-compiler"}
			if err := RunCommand(cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("failed to remove protobuf-compiler: %w", err)
			}
		}

	case "centos", "fedora", "rhel":
		// Check if installed
		if _, err := RunCommandWithOutput("rpm", "-q", "protobuf-compiler"); err == nil {
			cmd := []string{"sudo", "dnf", "remove", "-y", "protobuf-compiler"}
			if err := RunCommand(cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("failed to remove protobuf-compiler: %w", err)
			}
		}

	case "suse":
		// Check if installed
		if _, err := RunCommandWithOutput("rpm", "-q", "protobuf"); err == nil {
			cmd := []string{"sudo", "zypper", "--non-interactive", "remove", "protobuf"}
			if err := RunCommand(cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("failed to remove protobuf: %w", err)
			}
		}

	case "alpine":
		// Check if installed
		if _, err := RunCommandWithOutput("apk", "info", "protobuf"); err == nil {
			cmd := []string{"sudo", "apk", "del", "protobuf"}
			if err := RunCommand(cmd[0], cmd[1:]...); err != nil {
				return fmt.Errorf("failed to remove protobuf: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}

	return nil
}
