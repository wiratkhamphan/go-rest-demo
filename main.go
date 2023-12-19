package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", homePage)

	router.Run(":8080")
}

func homePage(c *gin.Context) {
	c.String(http.StatusOK, "This is my home page")

}
