package types

import "time"

const StatusIdle = "idle"
const StatusWorker = "worker"
const StatusProcess = "process"
const StatusFail = "fail"
const StatusReady = "ready"

var ScanStatus string
var ScanErrMsg string
var RedisInfo RedisInfoStruct
var ScanConfReq ScanConfReqStruct
var SortedReportListByCount SortByCount
var SortedReportListBySize SortBySize

type SortByCount []Report

func (a SortByCount) Len() int           { return len(a) }
func (a SortByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }
func (a SortByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type SortBySize []Report

func (a SortBySize) Len() int           { return len(a) }
func (a SortBySize) Less(i, j int) bool { return a[i].Size > a[j].Size }
func (a SortBySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Report struct {
	Key   string
	Count int64
	Size  int64
}

type KeyReports map[string]Report

type ScanConfReqStruct struct {
	ServerAddress string `form:"serveraddress" binding:"required"`
	Password      string `form:"password"`
	Pattern       string `form:"pattern" binding:"required"`
	GroupKey      bool   `form:"groupkey"`
	Delimiters    string `form:"delimiters"`
	MemoryUsage   bool   `form:"memoryusage"`
}

type RedisInfoStruct struct {
	TotalMemory   int64
	TotalKeyCount int64
	StartTime     time.Time
	EndTime       time.Time
}
