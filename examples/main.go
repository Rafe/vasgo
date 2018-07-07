package main

import (
	"fmt"

	"github.com/rafe/vasgo"
)

const VASGO_URL = "localhost:6379"

func main() {
	service := vasgo.NewService(VASGO_URL, "")

	app := vasgo.NewEndpoint("app", "0.0.1", "app.beansauce.io", false)
	app2 := vasgo.NewEndpoint("app", "0.0.1", "beta.app.beansauce.io", false)
	web := vasgo.NewEndpoint("web", "0.1.2", "web.beansauce.io", false)
	db := vasgo.NewEndpoint("db", "0.1.3", "db.beansauce.io", false)

	service.Register(app)
	service.Register(app2)
	service.Register(web)
	service.Register(db)

	dependencies := map[string]string{
		"app": "0.0.1",
		"web": "0.1.2",
		"db":  "0.1.3",
	}

	result, _ := service.FindDependencies(dependencies)

	for _, r := range result {
		fmt.Println(r)
	}
}
