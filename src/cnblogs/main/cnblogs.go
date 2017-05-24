package main

import (
	"cnblogs/client"
	"cnblogs/conf"
	"cnblogs/site"
)

func main() {
	conf := conf.ReadConf()
	if conf.AutoStartTask {
		client.Main(conf)
	}
	if conf.EnableSite {
		site.Main(conf)
	}
}
