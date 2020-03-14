package env

// Package env provides simple functions for defaulting non-existing environment variables

import (
	"os"
	"strconv"
)

// Int returns the environment variable for key, and defaults to value if it's not found.
//
// IMPORTANT NOTE - if the env var is not an integer, the value 0 will be returned
func Int(key string, value int) int {
	str, exists := os.LookupEnv(key)
	if !exists {
		return value
	}
	i, _ := strconv.Atoi(str)
	return i
}

// Float64 returns the environment variable for key, and defaults to value if it's not found.
//
// IMPORTANT NOTE - if the env var is not convertable to float64, the value 0.0 will be returned
func Float64(key string, value float64) float64 {
	str, exists := os.LookupEnv(key)
	if !exists {
		return value
	}
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

// String returns the environment variable for key, and defaults to value if it's not found.
func String(key string, value string) string {
	str, exists := os.LookupEnv(key)
	if !exists {
		return value
	}
	return str
}
