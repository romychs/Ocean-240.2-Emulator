package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const defaultMonitorFile = "okean240/MON_r6.BIN"
const defaultCPMFile = "okean240/CPM_r7.BIN"

type OkEmuConfig struct {
	LogFile     string `yaml:"logFile"`
	LogLevel    string `yaml:"logLevel"`
	MonitorFile string `yaml:"monitorFile"`
	CPMFile     string `yaml:"cpmFile"`
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
}
