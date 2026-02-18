package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type RedshiftConfigBlock struct {
	DayColorTemperature   int64  `toml:"temp-day"`
	NightColorTemperature int64  `toml:"temp-night"`
	LocationProvider      string `toml:"location-provider"`
}

type RedshiftManualConfigBlock struct {
	Lat float64 `toml:"lat"`
	Lon float64 `toml:"lon"`
}

type RedshiftConfig struct {
	Redshift RedshiftConfigBlock       `toml:"redshift"`
	Manual   RedshiftManualConfigBlock `toml:"manual"`
}

func ReadRedshiftConfig() (RedshiftConfig, error) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/redshift.conf")
	doc, err := os.ReadFile(configPath)

	lines := strings.Split(string(doc), "\n")
	var linesWithoutComments []string
	for _, line := range lines {
		// work around for toml parser not supporting comments starting with ;
		if strings.HasPrefix(line, ";") {
			continue
		}

		// work around for toml parser not supporting unquoted values
		if strings.Contains(line, "=manual") {
			line = strings.ReplaceAll(line, "=manual", "='manual'")
		}
		if strings.Contains(line, "=randr") {
			line = strings.ReplaceAll(line, "=randr", "='randr'")
		}

		linesWithoutComments = append(linesWithoutComments, line)
	}

	configWithoutComments := strings.Join(linesWithoutComments, "\n")

	var cfg RedshiftConfig
	err = toml.Unmarshal([]byte(configWithoutComments), &cfg)
	return cfg, err
}
