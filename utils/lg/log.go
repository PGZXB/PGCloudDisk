package lg

import (
	"PGCloudDisk/config"
	"log"
	"os"
	"strings"
)

var Logger *log.Logger

func Init() {
	logFilename := config.Cfg.Log.Filename

	if strings.ToLower(logFilename) == "stdout" { // 输出到标准输出
		Logger = log.New(os.Stdout, "[PGCloudDisk] ", log.LstdFlags|log.Lshortfile)
	} else { // 输出到文件
		logFile, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			log.Fatalln("Load log Failed")
		}
		Logger = log.New(logFile, "[PGCloudDisk] ", log.LstdFlags|log.Lshortfile)
	}

}
