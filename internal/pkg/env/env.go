package env

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var isDebugBuild = false

type Env struct {
	IsDebugBuild bool
}

func New() Env {
	err := godotenv.Load()

	if err == nil {
		log.Println("loaded .env file...")
	}

	return Env{IsDebugBuild: isDebugBuild}
}

func (e Env) GetBool(key string, def bool) bool {
	value, ok := os.LookupEnv(key)

	if !ok {
		return def
	}

	if len(value) == 0 {
		return false
	}

	f := strings.ToLower(string(value[0]))

	if f == "t" || f == "1" {
		return true
	}

	return false
}

func (e Env) GetString(key string, def string) string {
	value, ok := os.LookupEnv(key)

	if !ok {
		return def
	}

	return value
}

func (e Env) GetInt(key string, def int) int {
	value, ok := os.LookupEnv(key)

	if !ok {
		return def
	}

	intValue, err := strconv.Atoi(value)

	if err != nil {
		log.Printf("unable to convert value `%v` to an int for env.GetInt", value)
		return def
	}

	return intValue
}
