package main

import (
	"flag"
	"fmt"
	"html/template"
	"math"
	"redis-key-dashboard/api"
	"redis-key-dashboard/types"
	"redis-key-dashboard/worker"

	"github.com/gin-gonic/gin"
)

var (
	basicAuthActive = flag.Bool("basicAuthActive", false, "Basic Auth Active ?")
	adminUser       = flag.String("adminUser", "admin", "Admin User")
	adminPass       = flag.String("adminPass", "pass", "Admin Password")
	serverPort      = flag.String("serverPort", "8080", "Server Port")
)

func indexView(s int) string {
	return fmt.Sprintf("%d.", s+1)
}

func formatMb(s int64) string {
	return fmt.Sprintf("%.5f %s", float64(s)/1024/1024, "MB")
}

func formatMbRaw(s int64) float64 {
	return math.Round(float64(s)/1024/1024*10000) / 10000
}

func main() {

	flag.Parse()

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.SetFuncMap(template.FuncMap{
		"indexView":   indexView,
		"formatMb":    formatMb,
		"formatMbRaw": formatMbRaw,
	})
	router.LoadHTMLFiles("./template/index.html")

	if *basicAuthActive {
		authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
			*adminUser: *adminPass,
		}))

		authorized.GET("/", api.MainHandler)
		authorized.POST("/api/worker", api.WorkerHandler)
		authorized.POST("/api/reset-worker", api.ResetWorkerHandler)
		authorized.POST("/api/check-status", api.CheckStatusHandler)
		authorized.GET("/api/csv-export", api.CsvExportHandler)

	} else {
		router.GET("/", api.MainHandler)
		router.POST("/api/worker", api.WorkerHandler)
		router.POST("/api/reset-worker", api.ResetWorkerHandler)
		router.POST("/api/check-status", api.CheckStatusHandler)
		router.GET("/api/csv-export", api.CsvExportHandler)
	}

	types.ScanStatus = types.StatusIdle
	go func() {
		worker.Scanner()
	}()

	router.Run(":" + *serverPort)
}
