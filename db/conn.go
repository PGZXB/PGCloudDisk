package db

import (
	"PGCloudDisk/config"
	"PGCloudDisk/utils/lg"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"unsafe"
)

var conn *gorm.DB

type defaultErr struct {
}

func (d defaultErr) Error() string {
	return "default-error"
}

var dftErr defaultErr

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
	conn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		lg.Logger.Fatalln("Load mysql Failed")
	}
	lg.Logger.Printf("Connect MySQL %s:%d/%s Successfully\n", host, port, dbname)
}

func Transaction(fn func() bool) {
	err := conn.Transaction(func(tx *gorm.DB) error {
		var m sync.Mutex
		m.Lock() // 加锁 // FIXME 也许有更高效的方法
		defer m.Unlock()

		old := conn // 保存
		conn = tx
		if config.Cfg.RunMode.IsDebug {
			lg.Logger.Printf("conn @%v is reset to @%v\n", unsafe.Pointer(old), unsafe.Pointer(conn))
		}

		ok := fn()
		conn = old // 恢复
		if config.Cfg.RunMode.IsDebug {
			lg.Logger.Printf("conn @%v is recovered to @%v\n", unsafe.Pointer(tx), unsafe.Pointer(conn))
		}

		if ok { // 成功才提交
			return nil
		}
		return dftErr
	})
	if err != nil && config.Cfg.RunMode.IsDebug {
		lg.Logger.Println("Transaction Error. Already Rollback")
	}
}
