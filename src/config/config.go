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
	Domain                  string
	Port                    uint
	Protocol                string
	ApiPrefix               string
	RecaptchaVerifyEndpoint string
	RecaptchaSecret         string
}

type Configuration struct {
	EmailConfig EmailConfiguration
	WebConfig   WebConfiguration
	DbString    string
}

var defaultConfig = Configuration{
	EmailConfig: EmailConfiguration{},
	WebConfig: WebConfiguration{
		ApiPrefix:               "/api",
		RecaptchaVerifyEndpoint: "https://www.google.com/recaptcha/api/siteverify",
	},
	DbString: "root:passwd@tcp(localhost:3306)/mydatabase",
}

// flags
var configPath string
var createDefaultConfig bool
var DevMode bool

func InitConfig() {
	var err error
	flag.StringVar(&configPath, "cfg", "./config.json", "The path to a config file.")
	flag.BoolVar(&createDefaultConfig, "d", false, "Indicates whether the default config file should "+
		"be created. If so, the application terminates after having created it..")
	flag.BoolVar(&DevMode, "dev", false, "Enable development mode")
	flag.Parse()

	if createDefaultConfig {
		f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE, 0740)
		defer f.Close()
		if err != nil {
			log.Fatalf("Failed to create config file! Error: %s\n", err)
			return
		}
		var data []byte
		data, err = json.MarshalIndent(defaultConfig, "", "  ")
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
