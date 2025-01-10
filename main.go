package main

import (
	"fmt"
	"log"
	"regexp"
	"runtime"

	"github.com/robertt3kuk/protocInstall/utils"
)

func main() {
	if err := devToolsInstall(); err != nil {
		log.Fatal(err)
	}
}
func devToolsInstall() error {
	log.Printf("Starting devTools installation")
	platform := runtime.GOOS
	log.Printf("Detected platform: %s", platform)

	switch platform {
	case "darwin":
		log.Printf("Processing installation for Darwin/MacOS")
		var localProtocVersion string
		var stableProtocVersion string
		output, err := utils.RunCommandWithOutput("protoc", "--version")
		if err != nil {
			log.Printf("Protoc not found, attempting installation via brew")
			if err = utils.RunCommand("brew", "install", "protobuf"); err != nil {
				return fmt.Errorf("failed to install protobuf on darwin: %w", err)
			}
			log.Printf("Protobuf installed successfully, checking version")
			output, err = utils.RunCommandWithOutput("protoc", "--version")
			if err != nil {
				return fmt.Errorf("failed to get protoc version after installation: %w", err)
			}
		}
		// regex float number and get
		localProtocVersion = regexp.MustCompile(`\d+\.\d+`).FindString(string(output))

		log.Printf("Local protoc version: '%s'", (localProtocVersion))
		stableProtocVersion, err = utils.GetStableProtocVersion()
		if err != nil {
			return fmt.Errorf("failed to get stable protoc version: %w", err)
		}
		log.Printf("Stable protoc version: '%s'", (stableProtocVersion))

		if localProtocVersion != stableProtocVersion {
			log.Printf("Version mismatch detected, updating protobuf")
			if err = utils.RunCommand("brew", "install", "protobuf"); err != nil {
				return fmt.Errorf("failed to install protobuf on darwin: %w", err)
			}
			log.Printf("Protobuf updated successfully, verifying installation")
			_, err = utils.RunCommandWithOutput("protoc", "--version")
			if err != nil {
				return fmt.Errorf("failed to get protoc version after installation: %w", err)
			}
		}
	case "linux":
		log.Printf("Processing installation for Linux")
		log.Printf("Initializing version variables")
		var stableProtocVersion string
		var localProtocVersion string

		log.Printf("Detecting Linux distribution")
		distro, _, err := utils.DetectLinuxDistribution()
		if err != nil {
			log.Printf("Failed to detect Linux distribution: %v", err)
			return fmt.Errorf("failed to detect linux distribution: %w", err)
		}
		log.Printf("Detected Linux distribution: %s", distro)

		log.Printf("Removing existing protobuf from package manager")
		err = utils.RemovePackageManagerProtobuf(distro)
		if err != nil {
			log.Printf("Failed to remove existing protobuf: %v", err)
			return fmt.Errorf("failed to remove package manager protobuf: %w", err)
		}
		log.Printf("Successfully removed existing protobuf installation")

		log.Printf("Fetching stable protoc version")
		stableProtocVersion, err = utils.GetStableProtocVersion()
		if err != nil {
			log.Printf("Failed to get stable protoc version: %v", err)
			return fmt.Errorf("failed to get stable protoc version: %w", err)
		}
		log.Printf("Retrieved stable version: %s", stableProtocVersion)

		// Check if protoc is installed
		output, err := utils.RunCommandWithOutput("protoc", "--version")
		if err != nil {
			// If not installed, install it
			log.Printf("Protoc not found, attempting installation")
			if err = utils.InstallProtoBufLinuxGithub(stableProtocVersion); err != nil {
				return fmt.Errorf("failed to install protobuf on linux: %w", err)
			}
			err = utils.RunCommand("export", "PATH=$PATH:/home/user/protoc/bin")
			if err != nil {
				return fmt.Errorf("failed to add protoc to path: %w", err)
			}
			log.Printf("Protobuf installed successfully, checking version")
			_, err = utils.RunCommandWithOutput("protoc", "--version")
			if err != nil {
				return fmt.Errorf("failed to get protoc version after installation: %w", err)
			}
		}

		localProtocVersion = regexp.MustCompile(`\d+\.\d+`).FindString(string(output))
		if localProtocVersion == "" {
			return fmt.Errorf("failed to parse local protoc version from: %s", output)
		}

		log.Printf("Local protoc version: %s", localProtocVersion)
		log.Printf("Stable protoc version: %s", stableProtocVersion)

		if localProtocVersion != stableProtocVersion {
			log.Printf("Version mismatch detected, updating protobuf")
			if err = utils.InstallProtoBufLinuxGithub(stableProtocVersion); err != nil {
				return fmt.Errorf("failed to update protobuf on linux: %w", err)
			}
			log.Printf("Protobuf updated successfully")
		}

	default:
		log.Printf("Unsupported platform detected: %s", platform)
		return fmt.Errorf("unsupported platform: %s. Try to install protobuf manually", platform)
	}
	log.Printf("DevTools installation completed successfully")
	return nil
}
