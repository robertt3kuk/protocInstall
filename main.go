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
		var stableProtocVersion string
		var localProtocVersion string

		stableProtocVersion, err := utils.GetStableProtocVersion()
		if err != nil {
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
