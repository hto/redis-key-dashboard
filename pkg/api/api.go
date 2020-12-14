package api

import (
	"bytes"
	"encoding/csv"
	"log"
	"net/http"
	"redis-key-dashboard/pkg/types"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func MainHandler(c *gin.Context) {
	var workerTime float64
	if types.RedisInfo.EndTime.IsZero() {
		workerTime = time.Now().Sub(types.RedisInfo.StartTime).Seconds()
	} else {
		workerTime = types.RedisInfo.EndTime.Sub(types.RedisInfo.StartTime).Seconds()
	}

	report1Len, report2Len := 0, 0
	if len(types.SortedReportListByCount) < 25 {
		report1Len = len(types.SortedReportListByCount)
	} else {
		report1Len = 25
	}

	if len(types.SortedReportListBySize) < 25 {
		report2Len = len(types.SortedReportListBySize)
	} else {
		report2Len = 25
	}

	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"status":                  types.ScanStatus,
		"scanErrMsg":              types.ScanErrMsg,
		"scanConfReq":             types.ScanConfReq,
		"redisInfo":               types.RedisInfo,
		"workerTime":              workerTime,
		"sortedReportListByCount": types.SortedReportListByCount[0:report1Len],
		"sortedReportListBySize":  types.SortedReportListBySize[0:report2Len],
	})
}

func ResetWorkerHandler(c *gin.Context) {
	types.ScanStatus = types.StatusIdle
	types.ScanErrMsg = ""
	types.RedisInfo = types.RedisInfoStruct{}
	types.ScanConfReq = types.ScanConfReqStruct{}
	types.SortedReportListByCount = types.SortByCount{}
	types.SortedReportListBySize = types.SortBySize{}
}

func WorkerHandler(c *gin.Context) {
	if err := c.ShouldBindWith(&types.ScanConfReq, binding.Form); err != nil {
		c.JSON(401, gin.H{"message": "Invalid Form", "response": "err"})
		return
	}

	types.ScanStatus = types.StatusWorker
	c.JSON(200, gin.H{"response": "success"})
}

func CheckStatusHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": types.ScanStatus})
}

func CsvExportHandler(c *gin.Context) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=report:"+types.ScanConfReq.ServerAddress+".csv")

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	if err := w.Write([]string{"Key", "Count", "Size"}); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	isMemoryUsage := types.ScanConfReq.MemoryUsage
	if !isMemoryUsage {
		for _, d := range types.SortedReportListByCount {
			w.Write([]string{d.Key, strconv.FormatInt(d.Count, 10), "-"})
		}
	} else {
		for _, d := range types.SortedReportListBySize {
			w.Write([]string{d.Key, strconv.FormatInt(d.Count, 10), strconv.FormatInt(d.Size, 10)})
		}
	}
	w.Flush()

	c.Data(http.StatusOK, "text/csv", b.Bytes())
}
