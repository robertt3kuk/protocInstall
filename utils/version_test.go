package utils

import (
	"testing"
)

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		bumpType    string
		expected    string
		expectError bool
	}{
		{"Bump major version", "1.2.3", "major", "2.0.0", false},
		{"Bump minor version", "1.2.3", "minor", "1.3.0", false},
		{"Bump patch version", "1.2.3", "patch", "1.2.4", false},
		{"Bump major version with pre-release", "1.2.3-alpha", "major", "2.0.0", false},
		{"Bump minor version with pre-release", "1.2.3-beta.1", "minor", "1.3.0", false},
		{"Bump patch version with pre-release", "1.2.3-rc.1+build.123", "patch", "1.2.4", false},
		{"Invalid version format", "1.2", "patch", "", true},
		{"Invalid major version", "a.2.3", "patch", "", true},
		{"Invalid minor version", "1.b.3", "patch", "", true},
		{"Invalid patch version", "1.2.c", "patch", "", true},
		{"Invalid bump type", "1.2.3", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BumpVersion(tt.version, tt.bumpType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestBumpVersionEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		bumpType    string
		expected    string
		expectError bool
	}{
		{"Bump major version from 0", "0.1.2", "major", "1.0.0", false},
		{"Bump minor version from 0", "1.0.2", "minor", "1.1.0", false},
		{"Bump patch version from 0", "1.2.0", "patch", "1.2.1", false},
		{"Very large version numbers", "999999.999999.999999", "patch", "999999.999999.1000000", false},
		{"Version with leading zeros", "01.02.03", "minor", "1.3.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BumpVersion(tt.version, tt.bumpType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, but got %s", tt.expected, result)
				}
			}
		})
	}
}
