package worker

import (
	"redis-key-dashboard/pkg/types"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

func Scanner() {
	for {
		if types.ScanStatus == types.StatusWorker {
			time.Sleep(time.Duration(1) * time.Second)

			types.ScanStatus = types.StatusProcess
			types.RedisInfo.StartTime = time.Now()

			scan()

			types.RedisInfo.EndTime = time.Now()
			types.ScanStatus = types.StatusReady
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func scan() {
	client := redis.NewClient(&redis.Options{
		Addr:        types.ScanConfReq.ServerAddress,
		Password:    types.ScanConfReq.Password,
		DB:          0,
		ReadTimeout: 1 * time.Minute,
	})
	defer client.Close()

	redisI, _ := client.Do("MEMORY", "STATS").Result()
	if stats, ok := redisI.([]interface{}); ok {
		for i := 0; i < len(stats); i += 2 {
			switch stats[i] {
			case "total.allocated":
				if i+1 < len(stats) {
					types.RedisInfo.TotalMemory = stats[i+1].(int64)
				}
			case "keys.count":
				if i+1 < len(stats) {
					types.RedisInfo.TotalKeyCount = stats[i+1].(int64)
				}

			}
		}
	}

	mr := types.KeyReports{}
	delimiters := strings.Split(types.ScanConfReq.Delimiters, ",")
	cursor := uint64(0)
	groupKey := ""

	isGroupKey := types.ScanConfReq.GroupKey
	isMemoryUsage := types.ScanConfReq.MemoryUsage

	for {
		keys, cursor, err := client.Scan(cursor, types.ScanConfReq.Pattern, 1000).Result()
		if err != nil {
			types.ScanStatus = types.StatusFail
			types.ScanErrMsg = "Redis not connect !! => " + types.ScanConfReq.ServerAddress
			break
		}

		for _, key := range keys {
			scanKey(client, isGroupKey, isMemoryUsage, key, delimiters, groupKey, mr)
		}

		if cursor == 0 {
			break
		}
	}

	if isMemoryUsage {
		for _, report := range mr {
			types.SortedReportListBySize = append(types.SortedReportListBySize, report)
		}
		sort.Sort(types.SortedReportListBySize)
	} else {
		for _, report := range mr {
			types.SortedReportListByCount = append(types.SortedReportListByCount, report)
		}
		sort.Sort(types.SortedReportListByCount)
	}
}

func scanKey(client *redis.Client, isGroupKey, isMemoryUsage bool, key string, delimiters []string, groupKey string, mr types.KeyReports) {
	var memoryUsage int64
	if isMemoryUsage {
		memoryUsage, _ = client.MemoryUsage(key).Result()
	}

	if !isGroupKey {
		mr[key] = types.Report{Key: key, Count: 1, Size: memoryUsage}
		return
	}

	if len(delimiters) <= 1 {
		groupKey = key
	} else {
		for _, delimiter := range delimiters {
			tmp := strings.Split(key, delimiter)
			if len(tmp) > 1 {
				groupKey = strings.Join(tmp[0:len(tmp)-1], delimiter) + delimiter + "*"
				break
			}

			groupKey = key
		}
	}

	r := types.Report{}
	if _, ok := mr[groupKey]; ok {
		r = mr[groupKey]
	} else {
		r = types.Report{Key: groupKey}
	}

	r.Size += memoryUsage
	r.Count++
	mr[groupKey] = r
}
