package config

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type cfg struct {
	Log    logCfg   `yaml:"log"`
	MySQL  mysqlCfg `yaml:"mysql"`
	JwtCfg JwtCfg   `yaml:"jwt"`
}

type logCfg struct {
	Filename string `yaml:"filename"`
}

type JwtCfg struct {
	JwtSecret string `yaml:"jwtSecret"`
}

type mysqlCfg struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	Dbname   string `yaml:"dbname"`
	Charset  string `yaml:"charset"`
}

var Cfg *cfg

func init() {
	bytes, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalln("Load Config Failed")
	}

	Cfg = new(cfg)
	err = yaml.Unmarshal(bytes, Cfg)
	if err != nil {
		log.Fatalln("Load Config Failed")
	}

	if Cfg.MySQL.Host == "" {
		Cfg.MySQL.Host = "127.0.0.1"
	}

	if Cfg.MySQL.Port == 0 {
		Cfg.MySQL.Port = 3306
	}

	if Cfg.MySQL.Charset == "" {
		Cfg.MySQL.Charset = "utf8mb4"
	}

	if Cfg.Log.Filename == "" {
		Cfg.Log.Filename = "./PGCloudDisk_Log_" + time.Now().Format("2006_01_02") + ".log"
	}

	bytes, _ = yaml.Marshal(&Cfg)
	log.Println("Configure : \n", string(bytes))
}

func Init() {
	// do nothing
}
