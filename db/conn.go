package db

import (
	"PGCloudDisk/config"
	"PGCloudDisk/utils/lg"
	"fmt"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {
	var (
		user    = config.Cfg.MySQL.User
		pwd     = config.Cfg.MySQL.Password
		host    = config.Cfg.MySQL.Host
		port    = config.Cfg.MySQL.Port
		dbname  = config.Cfg.MySQL.Dbname
		charset = config.Cfg.MySQL.Charset
	)

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		user, pwd, host, port, dbname, charset)

	var err error
	db, err = gorm.Open(gormMysql.Open(dsn), &gorm.Config{})
	if err != nil {
		lg.Logger.Fatalln("Load mysql Failed")
	}
	lg.Logger.Printf("Connect MySQL %s:%d/%s Successfully\n", host, port, dbname)
}
