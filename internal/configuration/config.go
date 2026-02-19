package configuration

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Redshift RedshiftConfig `mapstructure:"redshift" yaml:"redshift"`
}

type RedshiftConfig struct {
	TransitionDuration time.Duration                  `mapstructure:"transitionDuration" yaml:"transitionDuration"`
	Brightness         RedshiftBrightnessConfig       `mapstructure:"brightness" yaml:"brightness"`
	ColorTemperature   RedshiftColorTemperatureConfig `mapstructure:"colorTemperature" yaml:"colorTemperature"`
}

type RedshiftBrightnessConfig struct {
	MinimumBrightness float64 `mapstructure:"minimumBrightness" yaml:"minimumBrightness"`
	MaximumBrightness float64 `mapstructure:"maximumBrightness" yaml:"maximumBrightness"`
}

type RedshiftColorTemperatureConfig struct {
	MinimumColorTemperature int64 `mapstructure:"minimumColorTemperature" yaml:"minimumColorTemperature"`
	MaximumColorTemperature int64 `mapstructure:"maximumColorTemperature" yaml:"maximumColorTemperature"`
}

var CurrentConfig Configuration

var currentUser, _ = user.Current()
var BaseDir = path.Join(currentUser.HomeDir, ".config", "system-control")

// InitConfig reads in config file and ENV variables if set.
func InitConfig(cfgFile string) {
	_ = ensureConfigDirExists()
	viper.SetConfigName("system-control")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			//ui.ErrorAndNotify("Path Error", "Couldn't detect home directory: %v", err)
			os.Exit(1)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath(home + "/.config/")
		viper.AddConfigPath(home + "/.config/" + "system-control/")
		viper.AddConfigPath("/etc/system-control/")
	}

	viper.AutomaticEnv() // read in environment variables that match

	setDefaultValues()
}

func ensureConfigDirExists() error {
	return os.MkdirAll(BaseDir, 0755)
}

func setDefaultValues() {
	viper.SetDefault("redshift.transitionDuration", 60*time.Minute)
	viper.SetDefault("redshift.brightness.minimumBrightness", 0.1)
	viper.SetDefault("redshift.brightness.maximumBrightness", 1.0)
	viper.SetDefault("redshift.colorTemperature.minimumColorTemperature", 1000)
	viper.SetDefault("redshift.colorTemperature.maximumColorTemperature", 25000)
}

// DetectAndReadConfigFile detects the path of the first existing config file
func DetectAndReadConfigFile() string {
	err := readInConfig()
	if err != nil {
		//ui.FatalWithoutStacktrace("Error reading config file, %s", err)
	}
	return GetFilePath()
}

// readInConfig reads and parses the config file
func readInConfig() error {
	return viper.ReadInConfig()
}

// GetFilePath this is only populated _after_ readInConfig()
func GetFilePath() string {
	return viper.ConfigFileUsed()
}

func LoadConfig() Configuration {
	// load default configuration values
	CurrentConfig = Configuration{}
	err := viper.Unmarshal(&CurrentConfig)
	if err != nil {
		//ui.Fatal("unable to decode into struct, %v", err)
	}
	return CurrentConfig
}

// PrintConfig prints the configuration to the console in YAML format
func PrintConfig() {
	configYaml, err := yaml.Marshal(CurrentConfig)
	if err != nil {
		log.Fatalf("Unable to marshal config to YAML: %v", err)
	}

	fmt.Println(string(configYaml))
}
