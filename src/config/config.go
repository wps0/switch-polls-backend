package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var Cfg *Configuration

type CORSConfiguration struct {
	AccessControlAllowOrigin  string
	AccessControlAllowHeaders string
}

type EmailConfiguration struct {
	OrganizationDomain string
	// The contents of 'FROM' email header
	SenderEmail           string
	SenderMailboxUsername string
	SmtpHost              string
	SmtpPort              int
	SenderEmailPasswd     string
	EmailSubject          string
	EmailTemplatePath     string
	// Internal-use only - contents of the file specified in EmailTemplatePath is loaded there upon startup
	EmailTemplate string `json:"-"`
}

type WebConfiguration struct {
	CORS                              CORSConfiguration
	ListeningAddress                  string
	Domain                            string
	Port                              uint
	Protocol                          string
	ApiPrefix                         string
	RecaptchaVerifyEndpoint           string
	RecaptchaSecret                   string
	TokenVerificationRedirectLocation string
}

type Configuration struct {
	EmailConfig EmailConfiguration
	WebConfig   WebConfiguration
	DbString    string
}

var defaultConfig = Configuration{
	EmailConfig: EmailConfiguration{
		OrganizationDomain:    "zsi.kielce.pl",
		SenderEmail:           "switch@zsi.kielce.pl",
		SenderMailboxUsername: "switch",
		SmtpHost:              "smtp.zsi.kielce.pl",
		SmtpPort:              587,
		SenderEmailPasswd:     "",
		EmailSubject:          "[SWITCH POLLS] Potwierdź swój głos",
		EmailTemplatePath:     "./EmailTemplate.html",
	},
	WebConfig: WebConfiguration{
		CORS: CORSConfiguration{
			AccessControlAllowOrigin:  "*",
			AccessControlAllowHeaders: "g-recaptcha-response",
		},
		ApiPrefix:               "/api",
		RecaptchaVerifyEndpoint: "https://www.google.com/recaptcha/api/siteverify",
	},
	DbString: "username:passwd@tcp(localhost:3306)/mydatabase",
}

// flags
var configPath string
var DevMode bool

func InitConfig() {
	log.Println("Initialising config...")
	var err error
	flag.StringVar(&configPath, "cfg", "./config.json", "The path to the config file.")
	flag.BoolVar(&DevMode, "dev", false, "Enable development mode.")
	flag.Parse()

	f, err := os.Open(configPath)
	if os.IsNotExist(err) {
		log.Printf("No config file found in %s (the path is customisable via -cfg <path> argument)", configPath)
		createConfig(configPath)
	} else {
		log.Printf("Found a file (possibly config) in '%s'", configPath)
		f.Close()
	}

	Cfg, err = loadConfiguration(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration! Error: %s\n", err)
		return
	}
	log.Printf("Config loaded!")
	//
	//log.Printf("Validating the sender email address %s...", Cfg.EmailConfig.SenderEmail)
	//err = checkmail.ValidateHostAndUser(Cfg.EmailConfig.SmtpHost, Cfg.EmailConfig.SenderEmail, Cfg.EmailConfig.SenderEmail)
	//if err != nil {
	//	log.Fatalf("Failed to validate the sender's email address %v", err)
	//	return
	//}
	log.Printf("Config initialised!")
}

func createConfig(path string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0740)
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
}

func loadConfiguration(path string) (*Configuration, error) {
	conf := &Configuration{}
	cfgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()
	d := json.NewDecoder(cfgFile)
	err = d.Decode(conf)
	if err != nil {
		return nil, err
	}

	loadEmailTemplate(conf, conf.EmailConfig.EmailTemplatePath)
	return conf, nil
}

func loadEmailTemplate(cfg *Configuration, path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicf("template file load error: %v\n", err)
	}
	cfg.EmailConfig.EmailTemplate = string(data)
}
