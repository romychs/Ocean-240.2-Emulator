package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const defaultMonitorFile = "rom/MON_v5.bin"
const defaultCPMFile = "rom/CPM_v5.bin"
const DefaultDebufPort = 10000

type OkEmuConfig struct {
	LogFile     string `yaml:"logFile"`
	LogLevel    string `yaml:"logLevel"`
	MonitorFile string `yaml:"monitorFile"`
	CPMFile     string `yaml:"cpmFile"`
	FloppyB     string `yaml:"floppyB"`
	FloppyC     string `yaml:"floppyC"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
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
	confFile := "okemu.yml"

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
	if conf.Host == "" {
		conf.Host = "localhost"
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
	if conf.Port < 80 || conf.Port > 65535 {
		log.Infof("Port %d incorrect, using default: %d", conf.Port, DefaultDebufPort)
		conf.Port = DefaultDebufPort
	}
}
