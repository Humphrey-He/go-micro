//go:build ignore

// R1 验收压测：1000 并发扣 100 库存 → 零超卖、零少卖、对账零差异。
//
// 运行（需先启动 inventory-service 与依赖）：
//
//	go run scripts/bench/prestock_load.go -u http://localhost:8082 -sku BENCH-001 -total 100 -c 1000
//
// 前置：DB 中存在 inventory 行（BENCH-001, stock>=100），stock_journal 表已建。
// 输出的 CHECK 三行即验收结论。
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	baseURL = flag.String("u", "http://localhost:8082", "inventory-service 地址")
	sku     = flag.String("sku", "BENCH-001", "压测 SKU")
	total   = flag.Int("total", 100, "总库存")
	conc    = flag.Int("c", 1000, "并发数")
)

func post(path string, body any) (map[string]any, error) {
	data, _ := json.Marshal(body)
	resp, err := http.Post(*baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	return out, json.Unmarshal(raw, &out)
}

func get(path string) (map[string]any, error) {
	resp, err := http.Get(*baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	return out, json.Unmarshal(raw, &out)
}

func main() {
	flag.Parse()
	client := &http.Client{Timeout: 5 * time.Second}
	_ = client

	// 1. 初始化额度 = total（演示口径：Redis 额度对齐 DB 真账）
	if _, err := post("/inventory/prestock/init", map[string]any{
		"sku_id": *sku, "quantity": *total, "request_id": "init",
	}); err != nil {
		fmt.Println("FATAL init:", err)
		return
	}

	// 2. 并发抢购
	var granted, insufficient, duplicate, failed int64
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			out, err := post("/inventory/prestock/reserve", map[string]any{
				"sku_id": *sku, "quantity": 1, "request_id": fmt.Sprintf("bench-%d", n),
			})
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}
			data, _ := out["data"].(map[string]any)
			switch data["result"] {
			case "GRANTED", "DUPLICATE":
				atomic.AddInt64(&granted, 1)
			case "INSUFFICIENT":
				atomic.AddInt64(&insufficient, 1)
			default:
				atomic.AddInt64(&failed, 1)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)

	// 3. 触发落库 + 对账
	time.Sleep(200 * time.Millisecond)
	syncOut, _ := post("/inventory/prestock/sync", map[string]any{})
	recOut, _ := post("/inventory/prestock/reconcile", map[string]any{})
	availOut, _ := get("/inventory/prestock/avail/" + *sku)

	fmt.Printf("并发=%d 库存=%d 耗时=%v\n", *conc, *total, elapsed.Round(time.Millisecond))
	fmt.Printf("granted=%d insufficient=%d duplicate=%d failed=%d\n",
		granted, insufficient, duplicate, failed)
	fmt.Printf("sync=%v reconcile=%v\n", syncOut["data"], recOut["data"])

	data, _ := availOut["data"].(map[string]any)
	fmt.Printf("CHECK 零超卖(成功数=库存):    %v\n", granted == int64(*total))
	fmt.Printf("CHECK 零少卖(拒绝数=并发-库存): %v\n", insufficient == int64(*conc-*total))
	fmt.Printf("CHECK Redis 剩余额度=0:       %v (avail=%v)\n",
		data != nil && data["avail"] == float64(0), data["avail"])
	if rd, ok := recOut["data"].(map[string]any); ok {
		fmt.Printf("CHECK 对账零差异:             %v (diffs=%v)\n", rd["Diffs"] == float64(0), rd["Diffs"])
	}
}
