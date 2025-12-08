package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
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
const EnvBatchingSuppressorSize = "BATCHING_SUPPRESSOR_SIZE"

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
		return 1 * time.Second
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		slog.Warn("invalid batching suppressor period, using default", "error", err, "default", "1s")
		return 1 * time.Second
	}

	return duration
}

func GetBatchingSuppressorSize() int {
	sizeStr := os.Getenv(EnvBatchingSuppressorSize)
	if sizeStr == "" {
		return 50
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		slog.Warn("invalid batching suppressor size, using default", "error", err, "default", 50)
		return 50
	}

	if size < 1 {
		slog.Warn("batching suppressor size too small, using minimum", "size", size, "minimum", 1)
		return 1
	}

	return size
}
