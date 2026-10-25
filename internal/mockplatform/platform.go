package mockplatform

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Platform 是「电商平台」的 mock：一份二进制以不同 env 起多个实例即多个平台。
// 三个运行期旋钮（/admin/config 可拨）正好对应分发调度要验证的三件事：
//   QPS_LIMIT —— 平台限流阈值（429 行为真实存在，验证调度器不超限）
//   LOSS_RATE —— 回执丢失率（验证轮次 diff 能抓漏发并补发）
//   重复投递幂等 —— 同 client_msg_id 再投返回原单（验证补发吸收）

type Task struct {
	ClientMsgID    string    `json:"client_msg_id"`
	PlatformTaskID string    `json:"platform_task_id"`
	SkuID          string    `json:"sku_id"`
	ShopID         string    `json:"shop_id"`
	ReceivedAt     time.Time `json:"received_at"`
	Lost           bool      `json:"-"` // 回执通道"丢失"：receipts 一直 404，直到重投唤醒
}

type Stats struct {
	Received     int64 `json:"received"`
	Duplicates   int64 `json:"duplicates"`
	Rejected429  int64 `json:"rejected_429"`
	Lost         int64 `json:"lost"`
	ReceiptReads int64 `json:"receipt_reads"`
}

type Platform struct {
	name     string
	mu       sync.Mutex
	tasks    map[string]*Task
	limiter  *rate.Limiter
	qps      float64
	lossRate float64
	stats    Stats
}

func New(name string, qps float64, lossRate float64) *Platform {
	return &Platform{
		name:     name,
		tasks:    make(map[string]*Task),
		limiter:  rate.NewLimiter(rate.Limit(qps), int(qps)+1),
		qps:      qps,
		lossRate: lossRate,
	}
}

func (p *Platform) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "platform": p.name}) })

	open := r.Group("/open/api")
	open.POST("/listing", p.listing)
	open.GET("/receipts/:client_msg_id", p.receipt)

	admin := r.Group("/admin")
	admin.POST("/config", p.setConfig)
	admin.GET("/config", p.getConfig)
	admin.GET("/stats", p.getStats)
	return r
}

// listing 受理商品分发。限流桶超了就 429——真实平台的限流行为。
func (p *Platform) listing(c *gin.Context) {
	var req struct {
		ClientMsgID string `json:"client_msg_id"`
		SkuID       string `json:"sku_id"`
		ShopID      string `json:"shop_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ClientMsgID == "" {
		c.JSON(400, gin.H{"code": 40001, "message": "invalid request"})
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.limiter.Allow() {
		p.stats.Rejected429++
		c.JSON(429, gin.H{"code": 42901, "message": "rate limited by platform"})
		return
	}

	if t, ok := p.tasks[req.ClientMsgID]; ok {
		// 重投唤醒丢失的回执通道（模拟"重推后平台补回执"）
		p.stats.Duplicates++
		t.Lost = false
		c.JSON(200, gin.H{"code": 0, "data": gin.H{
			"platform_task_id": t.PlatformTaskID, "duplicate": true}})
		return
	}

	lost := rand.Float64() < p.lossRate
	t := &Task{
		ClientMsgID:    req.ClientMsgID,
		PlatformTaskID: fmt.Sprintf("%s-%d", p.name, time.Now().UnixNano()),
		SkuID:          req.SkuID,
		ShopID:         req.ShopID,
		ReceivedAt:     time.Now(),
		Lost:           lost,
	}
	p.tasks[req.ClientMsgID] = t
	p.stats.Received++
	if lost {
		p.stats.Lost++
	}
	c.JSON(200, gin.H{"code": 0, "data": gin.H{
		"platform_task_id": t.PlatformTaskID, "duplicate": false}})
}

// receipts 回执查询。丢失中的任务永远 404——diff job 抓的就是它。
func (p *Platform) receipt(c *gin.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.ReceiptReads++

	t, ok := p.tasks[c.Param("client_msg_id")]
	if !ok || t.Lost {
		c.JSON(404, gin.H{"code": 40401, "message": "receipt not found"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": gin.H{
		"client_msg_id":    t.ClientMsgID,
		"platform_task_id": t.PlatformTaskID,
		"status":           "ACKED"}})
}

func (p *Platform) setConfig(c *gin.Context) {
	var req struct {
		QPS      *float64 `json:"qps"`
		LossRate *float64 `json:"loss_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 40001, "message": "invalid config"})
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if req.QPS != nil && *req.QPS > 0 {
		p.qps = *req.QPS
		p.limiter.SetLimit(rate.Limit(p.qps))
		p.limiter.SetBurst(int(p.qps) + 1)
	}
	if req.LossRate != nil {
		p.lossRate = *req.LossRate
	}
	c.JSON(200, gin.H{"code": 0, "data": p.snapshot()})
}

func (p *Platform) getConfig(c *gin.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	c.JSON(200, gin.H{"code": 0, "data": p.snapshot()})
}

func (p *Platform) getStats(c *gin.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	c.JSON(200, gin.H{"code": 0, "data": p.stats})
}

func (p *Platform) snapshot() gin.H {
	return gin.H{"platform": p.name, "qps": p.qps, "loss_rate": p.lossRate}
}

// StatsSnapshot / SetLost 供测试与运维脚本读取状态、指定某条任务丢回执。
func (p *Platform) StatsSnapshot() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

func (p *Platform) SetLost(clientMsgID string, lost bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if t, ok := p.tasks[clientMsgID]; ok {
		t.Lost = lost
	}
}
