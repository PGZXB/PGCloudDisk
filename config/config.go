package config

import (
	"PGCloudDisk/utils/fileutils"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type cfg struct {
	Log          logCfg       `yaml:"log"`
	MySQL        mysqlCfg     `yaml:"mysql"`
	JwtCfg       jwtCfg       `yaml:"jwt"`
	LocalSaveCfg localSaveCfg `yaml:"localSave"`
}

type logCfg struct {
	Filename string `yaml:"filename"`
}

type jwtCfg struct {
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

type localSaveCfg struct {
	Root string `yaml:"root"`
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

	// 验证localSave.Root的合法性, 不合法则产生默认目录
	if !fileutils.IsDir(Cfg.LocalSaveCfg.Root) {

		log.Println("Use Default Local-Save Root Path")

		// 默认在 运行路径/CloudDiskFiles
		path, err := os.Getwd()
		if err != nil {
			log.Fatalln("Get Current Path Failed")
		}
		path = filepath.Join(path, "CloudDiskFiles")

		// 目录存在则以, 不存在则要创建
		if !fileutils.IsDir(path) {
			err = os.Mkdir(path, 0755)
			if err != nil {
				log.Fatalln("Create LocalSave Path Failed")
			}
		}

		Cfg.LocalSaveCfg.Root = path
	}

	bytes, _ = yaml.Marshal(&Cfg)
	log.Println("Configure : \n", string(bytes))
}

func Init() {
	// do nothing
}
