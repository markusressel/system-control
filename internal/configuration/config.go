package configuration

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Daemon   DaemonConfig   `mapstructure:"daemon" yaml:"daemon"`
	Redshift RedshiftConfig `mapstructure:"redshift" yaml:"redshift"`
	Hotspot  HotspotConfig  `mapstructure:"hotspot" yaml:"hotspot"`
	Fan      FanConfig      `mapstructure:"fan" yaml:"fan"`
}

// FanConfig describes where the fan policy of the laptop is controlled.
type FanConfig struct {
	// PolicyFileGlob is a glob matching the sysfs file(s) holding the fan mode (0, 1, 2).
	// Defaults to the ASUS WMI policy file.
	PolicyFileGlob string `mapstructure:"policyFileGlob" yaml:"policyFileGlob"`
}

// HotspotConfig contains the defaults of `network wifi hotspot`. Empty values are detected or
// generated when the hotspot is turned on.
type HotspotConfig struct {
	Interface string `mapstructure:"interface" yaml:"interface"`
	SSID      string `mapstructure:"ssid" yaml:"ssid"`
	Password  string `mapstructure:"password" yaml:"password"`
}

type DaemonConfig struct {
	PollIntervals map[string]time.Duration `mapstructure:"pollIntervals" yaml:"pollIntervals"`
}

type RedshiftConfig struct {
	TransitionDuration time.Duration                  `mapstructure:"transitionDuration" yaml:"transitionDuration"`
	Brightness         RedshiftBrightnessConfig       `mapstructure:"brightness" yaml:"brightness"`
	ColorTemperature   RedshiftColorTemperatureConfig `mapstructure:"colorTemperature" yaml:"colorTemperature"`
	Gamma              RedshiftGammaConfig            `mapstructure:"gamma" yaml:"gamma"`
}

type RedshiftBrightnessConfig struct {
	MinimumBrightness float64 `mapstructure:"minimumBrightness" yaml:"minimumBrightness"`
	MaximumBrightness float64 `mapstructure:"maximumBrightness" yaml:"maximumBrightness"`
}

type RedshiftColorTemperatureConfig struct {
	MinimumColorTemperature int64 `mapstructure:"minimumColorTemperature" yaml:"minimumColorTemperature"`
	MaximumColorTemperature int64 `mapstructure:"maximumColorTemperature" yaml:"maximumColorTemperature"`
}

type RedshiftGammaConfig struct {
	MinimumGamma float64 `mapstructure:"minimumGamma" yaml:"minimumGamma"`
	MaximumGamma float64 `mapstructure:"maximumGamma" yaml:"maximumGamma"`
}

var CurrentConfig Configuration

// configMu guards CurrentConfig and viper. A long running daemon reloads the file while
// requests read the configuration.
var configMu sync.RWMutex

// fileCheckInterval limits how often Get looks at the modification time of the config file.
const fileCheckInterval = time.Second

var (
	// fileStamp identifies the version of the config file CurrentConfig was read from.
	fileStamp   fileVersion
	lastChecked time.Time
)

type fileVersion struct {
	modTime time.Time
	size    int64
}

func statConfigFile() (fileVersion, bool) {
	file := viper.ConfigFileUsed()
	if file == "" {
		return fileVersion{}, false
	}
	info, err := os.Stat(file)
	if err != nil {
		return fileVersion{}, false
	}
	return fileVersion{info.ModTime(), info.Size()}, true
}

// Get returns a snapshot of the current configuration.
//
// A long running process (the daemon) notices edits of the config file: it is read again when
// its modification time changed. A file that fails to parse (e.g. saved half-way) keeps the
// previous configuration.
func Get() Configuration {
	reloadIfChanged()

	configMu.RLock()
	defer configMu.RUnlock()
	return CurrentConfig
}

func reloadIfChanged() {
	configMu.Lock()
	defer configMu.Unlock()

	if time.Since(lastChecked) < fileCheckInterval {
		return
	}
	lastChecked = time.Now()

	stamp, ok := statConfigFile()
	if !ok || stamp == fileStamp {
		return
	}
	if err := reloadLocked(); err != nil {
		log.Printf("Could not reload the configuration, keeping the previous one: %v", err)
		// do not retry (and log) until the file changes again
		fileStamp = stamp
	}
}

// Reload reads the configuration file again, regardless of its modification time.
func Reload() error {
	configMu.Lock()
	defer configMu.Unlock()
	return reloadLocked()
}

func reloadLocked() error {
	if err := readInConfig(); err != nil {
		return err
	}
	next := Configuration{}
	if err := viper.Unmarshal(&next); err != nil {
		return err
	}
	CurrentConfig = next
	fileStamp, _ = statConfigFile()
	return nil
}

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
	viper.SetDefault("redshift.gamma.minimumGamma", 0.1)
	viper.SetDefault("redshift.gamma.maximumGamma", 2.0)
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
	configMu.Lock()
	defer configMu.Unlock()

	// load default configuration values
	CurrentConfig = Configuration{}
	err := viper.Unmarshal(&CurrentConfig)
	fileStamp, _ = statConfigFile()
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
