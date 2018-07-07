package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rafe/vasgo"
)

const VASGO_URL = "localhost:6379"

func setup(service *vasgo.Service) {
	app := vasgo.NewEndpoint("app", "0.0.1", "beta.app.beansauce.io", false)
	web := vasgo.NewEndpoint("web", "0.1.2", "web.beansauce.io", false)
	db := vasgo.NewEndpoint("db", "0.1.3", "db.beansauce.io", false)
	service.Register(app)
	service.Register(web)
	service.Register(db)
}

func main() {
	service := vasgo.NewService(VASGO_URL, "")
	setup(service)

	app := vasgo.NewEndpoint("app", "0.0.1", "app.beansauce.io", false)

	dependencies := map[string]string{
		"app": "0.0.1",
		"web": "0.1.2",
		"db":  "0.1.3",
	}

	go service.Register(app)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		endpoints, err := service.FindDependencies(dependencies)
		if err != nil {
			c.JSON(200, gin.H{
				"status": "error",
			})
		}

		c.JSON(200, endpoints)
	})
	r.POST("/register", func(c *gin.Context) {
		name := c.PostForm("name")
		version := c.PostForm("version")
		url := c.PostForm("url")

		app := vasgo.NewEndpoint(name, version, url, true)
		_, err := service.Register(app)

		if err != nil {
			c.JSON(200, gin.H{
				"status": "error",
			})
		}

		c.JSON(200, gin.H{
			"status": "success",
		})
	})
	r.Run()
}
