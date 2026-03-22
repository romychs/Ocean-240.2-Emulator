package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const defaultMonitorFile = "rom/MON_v5.bin"
const defaultCPMFile = "rom/CPM_v5.bin"
const DefaultDebufPort = 10000
const confFile = "okemu.yml"

type OkEmuConfig struct {
	LogFile     string         `yaml:"logFile"`
	LogLevel    string         `yaml:"logLevel"`
	MonitorFile string         `yaml:"monitorFile"`
	CPMFile     string         `yaml:"cpmFile"`
	FDC         []FDCConfig    `yaml:"fdc"`
	Debugger    DebuggerConfig `yaml:"debugger"`
}

type FDCConfig struct {
	AutoLoad   bool   `yaml:"autoLoad"`
	AutoSave   bool   `yaml:"autoSave"`
	FloppyFile string `yaml:"floppyFile"`
}

type DebuggerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

var config *OkEmuConfig

//func usage() {
//	fmt.Printf("Use: %s --config <file_path>.yml\n", filepath.Base(os.Args[0]))
//	os.Exit(2)
//}

func GetConfig() *OkEmuConfig {
	return config
}

func LoadConfig() {
	//args := os.Args
	//if len(args) != 3 {
	//	usage()
	//}
	//
	//if args[1] != "--config" {
	//	usage()
	//}
	// confFile := args[2]

	conf := OkEmuConfig{}
	data, err := os.ReadFile(confFile)
	if err == nil {
		err := yaml.Unmarshal(data, &conf)
		if err != nil {
			log.Panicf("Config file error: %v", err)
		}
		setDefaultConf(&conf)
		checkConfig(&conf)
	} else {
		log.Panicf("Can not open config file: %v", err)
	}

	config = &conf
}

func checkConfig(conf *OkEmuConfig) {
	if conf.Debugger.Host == "" {
		conf.Debugger.Host = "localhost"
	}
}

func setDefaultConf(conf *OkEmuConfig) {
	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}
	if conf.LogFile == "" {
		conf.LogFile = "okemu.log"
	}
	if conf.MonitorFile == "" {
		conf.MonitorFile = defaultMonitorFile
	}
	if conf.CPMFile == "" {
		conf.CPMFile = defaultCPMFile
	}
	if conf.Debugger.Port < 80 || conf.Debugger.Port > 65535 {
		log.Infof("Port %d incorrect, using default: %d", conf.Debugger.Port, DefaultDebufPort)
		conf.Debugger.Port = DefaultDebufPort
	}
}

func (c *OkEmuConfig) Clone() *OkEmuConfig {

	fds := make([]FDCConfig, 2)
	for n, fd := range c.FDC {
		fds[n].FloppyFile = fd.FloppyFile
		fds[n].AutoLoad = fd.AutoLoad
		fds[n].AutoSave = fd.AutoSave
	}

	return &OkEmuConfig{
		LogFile:     c.LogFile,
		LogLevel:    c.LogLevel,
		MonitorFile: c.MonitorFile,
		CPMFile:     c.CPMFile,
		FDC:         fds,
		Debugger: DebuggerConfig{
			Enabled: c.Debugger.Enabled,
			Host:    c.Debugger.Host,
			Port:    c.Debugger.Port,
		},
	}
}

func (c *OkEmuConfig) Save() {
	data, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("config error: %v", err)
	}
	err = os.WriteFile(confFile, data, 0600)
	if err != nil {
		log.Errorf("save config file: %v", err)
	}
}
