package main

import (
	"cnblogs/client"
	"cnblogs/conf"
	"cnblogs/site"
)

func main() {
	conf := conf.ReadConf()
	client.Main(conf)
	site.Main(conf)
}
