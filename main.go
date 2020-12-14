package main

import (
	"flag"
	"fmt"
	"html/template"
	"math"
	"redis-key-dashboard/pkg/api"
	"redis-key-dashboard/pkg/types"
	"redis-key-dashboard/pkg/worker"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	basicAuth  = flag.String("auth", "admin:admin", "Admin User")
	serverPort = flag.Int("port", 8080, "Server Port")
)

func main() {
	flag.Parse()

	router := gin.Default()
	router.Static("/assets", "./assets")
	router.SetFuncMap(template.FuncMap{
		"indexView": func(s int) string {
			return fmt.Sprintf("%d.", s+1)
		},
		"formatMib": func(s int64) string {
			return fmt.Sprintf("%.5f %s", float64(s)/1024/1024, "MiB")
		},
		"formatMibRaw": func(s int64) float64 {
			return math.Round(float64(s)/1024/1024*10000) / 10000
		},
	})
	router.LoadHTMLFiles("./template/index.html")

	if *basicAuth != "" {
		parts := strings.Split(*basicAuth, ":")
		router.Use(gin.BasicAuth(gin.Accounts{parts[0]: parts[1]}))
	}

	router.GET("/", api.MainHandler)
	router.POST("/api/worker", api.WorkerHandler)
	router.POST("/api/reset-worker", api.ResetWorkerHandler)
	router.POST("/api/check-status", api.CheckStatusHandler)
	router.GET("/api/csv-export", api.CsvExportHandler)

	types.ScanStatus = types.StatusIdle
	go func() {
		worker.Scanner()
	}()

	router.Run(fmt.Sprintf(":%d", *serverPort))
}
