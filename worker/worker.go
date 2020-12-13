package worker

import (
	"redis-key-dashboard/types"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

func Scanner() {

	var err error
	for {
		if types.ScanStatus == types.StatusWorker {

			time.Sleep(time.Duration(1) * time.Second)

			types.ScanStatus = types.StatusProcess
			types.RedisInfo.StartTime = time.Now()

			var client *redis.Client
			client = redis.NewClient(&redis.Options{
				Addr:        types.ScanConfReq.ServerAddress,
				Password:    types.ScanConfReq.Password,
				DB:          0,
				ReadTimeout: 1 * time.Minute,
			})

			redisI, _ := client.Do("MEMORY", "STATS").Result()
			if stats, ok := redisI.([]interface{}); ok {
				for i := 0; i < len(stats); i += 2 {
					if stats[i] == "total.allocated" && i+1 < len(stats) {
						types.RedisInfo.TotalMemory = stats[i+1].(int64)
					}

					if stats[i] == "keys.count" && i+1 < len(stats) {
						types.RedisInfo.TotalKeyCount = stats[i+1].(int64)
					}
				}
			}

			mr := types.KeyReports{}
			delimiters := strings.Split(types.ScanConfReq.Delimiters, ",")
			cursor := uint64(0)
			groupKey := ""

			for {
				var keys []string
				if keys, cursor, err = client.Scan(cursor, types.ScanConfReq.Pattern, 1000).Result(); err != nil {
					types.ScanStatus = types.StatusFail
					types.ScanErrMsg = "Redis not connect !! => " + types.ScanConfReq.ServerAddress
					break
				}

				for _, key := range keys {

					var memoryUsage int64
					if types.ScanConfReq.MemoryUsage {
						_, _ = client.Type(key).Result()
						memoryUsage, _ = client.MemoryUsage(key).Result()
					}

					if types.ScanConfReq.GroupKey {
						if len(delimiters) > 1 {
							for _, delimiter := range delimiters {
								tmp := strings.Split(key, delimiter)
								if len(tmp) > 1 {
									groupKey = strings.Join(tmp[0:len(tmp)-1], delimiter) + delimiter + "*"
									break
								} else {
									groupKey = key
								}
							}
						} else {
							groupKey = key
						}

						r := types.Report{}
						if _, ok := mr[groupKey]; ok {
							r = mr[groupKey]
						} else {
							r = types.Report{groupKey, 0, 0}
						}

						r.Size += memoryUsage
						r.Count++
						mr[groupKey] = r
					} else {
						mr[key] = types.Report{key, 1, memoryUsage}
					}
				}

				if cursor == 0 {
					break
				}
			}

			if (types.ScanConfReq.GroupKey && !types.ScanConfReq.MemoryUsage) || (!types.ScanConfReq.GroupKey && !types.ScanConfReq.MemoryUsage) {
				for _, report := range mr {
					types.SortedReportListByCount = append(types.SortedReportListByCount, report)
				}
				sort.Sort(types.SortedReportListByCount)
			}

			if types.ScanConfReq.MemoryUsage {
				for _, report := range mr {
					types.SortedReportListBySize = append(types.SortedReportListBySize, report)
				}
				sort.Sort(types.SortedReportListBySize)
			}

			types.RedisInfo.EndTime = time.Now()
			types.ScanStatus = types.StatusReady
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}
