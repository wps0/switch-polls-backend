package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

var Cfg *Configuration

type EmailConfiguration struct {
	SenderEmail       string
	SenderEmailPasswd string
	EmailTemplatePath string
}

type WebConfiguration struct {
	Domain    string
	Port      uint
	Protocol  string
	ApiPrefix string
}

type Configuration struct {
	EmailConfig EmailConfiguration
	WebConfig   WebConfiguration
	DbString string
}

// flags
var configPath string
var createDefaultConfig bool

func InitConfig() {
	var err error
	flag.StringVar(&configPath, "cfg", "./config.json", "The path to a config file.")
	flag.BoolVar(&createDefaultConfig, "c", false, "Indicates whether the default config file should "+
		"be created. If so, the application terminates after having created it..")
	flag.Parse()

	if createDefaultConfig {
		f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE, 0740)
		defer f.Close()
		if err != nil {
			log.Fatalf("Failed to create config file! Error: %s\n", err)
			return
		}
		var data []byte
		data, err = json.MarshalIndent(Configuration{}, "", "  ")
		if err != nil {
			log.Fatalf("Failed to create config file! Error: %s\n", err)
			return
		}
		_, err = f.Write(data)
		if err != nil {
			log.Fatalf("Failed to create config file! Error: %s\n", err)
			return
		}
		f.Close()
	}

	Cfg, err = loadConfiguration(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration! Error: %s\n", err)
		return
	}
	log.Printf("Config loaded!")
}

func loadConfiguration(conf_path string) (*Configuration, error) {
	conf := &Configuration{}
	cfg_file, err := os.Open(conf_path)
	if err != nil {
		return nil, err
	}
	defer cfg_file.Close()
	d := json.NewDecoder(cfg_file)
	err = d.Decode(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
