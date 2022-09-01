package config

import (
	"flag"
	"os"
)

const EnvAccessToken = "ACCESS_TOKEN"
const EnvApiEndpoint = "API_ENDPOINT"
const envSuppressionListName = "LIST_NAME"

var isInitFullscanSkipped bool

func InitFlags() {
	flag.BoolVar(&isInitFullscanSkipped, "skip-init-fullscan", false, "")

	flag.Parse()
}

func IsInitFullscanSkipped() bool {
	return isInitFullscanSkipped
}

func GetSuppressionListName() string {
	return os.Getenv(envSuppressionListName)
}
