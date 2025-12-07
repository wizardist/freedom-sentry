package config

import (
	"flag"
	"os"
)

const EnvAccessToken = "ACCESS_TOKEN"
const EnvApiEndpoint = "API_ENDPOINT"
const envSuppressionListName = "LIST_NAME"
const EnvWikiDomain = "WIKI_DOMAIN"
const EnvEventStreamsURL = "EVENTSTREAMS_URL"

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

func GetEventStreamsURL() string {
	url := os.Getenv(EnvEventStreamsURL)
	if url == "" {
		return "https://stream.wikimedia.org/v2/stream/recentchange"
	}
	return url
}

func GetWikiDomain() string {
	return os.Getenv(EnvWikiDomain)
}
