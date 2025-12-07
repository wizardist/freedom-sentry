package config

import (
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"
)

const EnvAccessToken = "ACCESS_TOKEN"
const EnvApiEndpoint = "API_ENDPOINT"
const envSuppressionListName = "LIST_NAME"
const EnvWikiDomain = "WIKI_DOMAIN"
const EnvEventStreamsURL = "EVENTSTREAMS_URL"
const EnvLogLevel = "LOG_LEVEL"
const EnvBatchingSuppressorPeriod = "BATCHING_SUPPRESSOR_PERIOD"

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

func GetLogLevel() slog.Level {
	level := strings.ToUpper(os.Getenv(EnvLogLevel))
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func GetBatchingSuppressorPeriod() time.Duration {
	period := os.Getenv(EnvBatchingSuppressorPeriod)
	if period == "" {
		return 5 * time.Second
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		slog.Warn("invalid batching suppressor period, using default", "error", err, "default", "5s")
		return 5 * time.Second
	}

	return duration
}
