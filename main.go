package main

import (
	"PGCloudDisk/db"
	"PGCloudDisk/routers"
	"PGCloudDisk/utils/lg"
)

func init() {
	lg.Init()
	db.Init()
}

func main() {
	err := routers.Init().Run(":8080")
	if err != nil {
		lg.Logger.Fatalf("%#v", err)
	}
}
