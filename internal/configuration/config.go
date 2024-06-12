package configuration

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"path"
	"time"
)

type RedshiftConfig struct {
	TransitionDuration time.Duration `json:"transitionDuration"`
}

type Configuration struct {
	Redshift RedshiftConfig `json:"redshift"`
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
		viper.AddConfigPath("/etc/system-control/")
	}

	viper.AutomaticEnv() // read in environment variables that match
}

func ensureConfigDirExists() error {
	return os.MkdirAll(BaseDir, 0755)
}

func setDefaultValues() {
	viper.SetDefault("Redshift.TransitionDuration", 60*time.Minute)
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
