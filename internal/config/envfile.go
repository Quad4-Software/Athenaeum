package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv reads optional dotenv files into the process environment.
// Existing environment variables are never overwritten.
//
// Search order:
//  1. ATHENAEUM_ENV_FILE when set
//  2. .env in the current working directory
func LoadEnv() error {
	if path := strings.TrimSpace(os.Getenv("ATHENAEUM_ENV_FILE")); path != "" {
		if err := loadEnvFile(path); err != nil {
			return err
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dotenv := filepath.Join(cwd, ".env")
	if _, err := os.Stat(dotenv); err == nil {
		return loadEnvFile(dotenv)
	}
	return nil
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path) // #nosec G304 G703 -- operator-configured env file path
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read env file %s: %w", path, err)
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("env file %s:%d: invalid line", path, lineNo)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("env file %s:%d: empty key", path, lineNo)
		}
		value = strings.TrimSpace(unquoteEnvValue(value))
		if os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("env file %s:%d: %w", path, lineNo, err)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("read env file %s: %w", path, err)
	}
	return nil
}

func unquoteEnvValue(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// Env returns the environment variable key when set, otherwise def.
func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// EnvBool parses a boolean environment variable.
func EnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
