package main

import (
	"fmt"

	"github.com/beka652/rss_feed_aggregator/internal/config"
)


func main() {
	cfg, err  := config.Read()
	if err != nil {
		fmt.Println(err)
		return 
	}
	fmt.Println(cfg)
	
	err = cfg.SetUser("lane")
	if err != nil {
		fmt.Println(err)
		return 
	}
	
	cfg, err = config.Read()
	if err != nil {
		fmt.Println(err)
		return 
	}
	fmt.Println(cfg)
	
}