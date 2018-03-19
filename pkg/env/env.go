package env

import (
	"fmt"
	"os"
	"strconv"
)

func Get(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		fmt.Printf("Environment variable '%s' not set\n", key)
	}

	return value
}

func GetBool(key string) bool {
	if Get(key) == "yes" {
		return true
	}
	return false
}

func GetInt(key string) int {
	val, err := strconv.Atoi(Get(key))
	if err != nil {
		fmt.Printf("Error while setting env '%s': ", key)
		fmt.Println(err)
	}
	return val
}
