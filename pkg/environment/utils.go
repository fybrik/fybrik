// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"os"
	"strconv"
)

// Returns the integer value of an environment variable.
// If the environment variable is not set or cannot be parsed the default value is returned.
func GetEnvAsInt(key string, defaultValue int) int {
	if env, isSet := os.LookupEnv(key); isSet {
		i, err := strconv.Atoi(env)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

// Returns the float32 value of an environment variable.
// If the environment variable is not set or cannot be parsed the default value is returned.
func GetEnvAsFloat32(key string, defaultValue float32) float32 {
	if env, isSet := os.LookupEnv(key); isSet {
		f, err := strconv.ParseFloat(env, 32)
		if err == nil {
			return float32(f)
		}
	}
	return defaultValue
}

func MustGetEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}
