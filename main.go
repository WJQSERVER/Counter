package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port string `yaml:"port"`
	File string `yaml:"file"`
}

type Stats struct {
	Total int64            `json:"total"`
	Daily map[string]int64 `json:"daily"`
}

var (
	config  Config
	stats   Stats
	statsMu sync.Mutex
	buffer  = make(chan int, 100)
)

func main() {
	// 加载配置
	loadConfig()

	http.HandleFunc("/add", handleRequest)
	http.HandleFunc("/api/counter", handleRead)
	http.HandleFunc("/api/counter/daily", handleDaily)
	http.HandleFunc("/api/counter/total", handleTotal)

	// 从文件中加载统计数据
	loadStats()

	// 每10秒将缓冲区数据写入文件
	go flushStats()

	// 每天0点更新每日统计数据
	go updateDailyStats()

	http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 异步统计请求
	if r.Method == http.MethodGet && r.URL.Path == "/add" {
		go countRequest()
	}
	fmt.Fprintf(w, "Request received!")
}

func handleRead(w http.ResponseWriter, r *http.Request) {
	// 读取统计数据并返回
	statsMu.Lock()
	data, _ := json.Marshal(stats)
	statsMu.Unlock()
	fmt.Fprintf(w, string(data))
}

// 读取今日统计数据
func handleDaily(w http.ResponseWriter, r *http.Request) {
	// 读取今日统计数据并返回
	statsMu.Lock()
	today := time.Now().Format("2006-01-02")
	count := stats.Daily[today]
	statsMu.Unlock()

	// 增加适当的错误处理机制
	if _, err := fmt.Fprintf(w, "%d", count); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func handleTotal(w http.ResponseWriter, r *http.Request) {
	// 读取总计数并返回
	statsMu.Lock()
	total := stats.Total
	statsMu.Unlock()

	// 增加适当的错误处理机制
	if _, err := fmt.Fprintf(w, "%d", total); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func countRequest() {
	buffer <- 1

	statsMu.Lock()
	defer statsMu.Unlock()

	today := time.Now().Format("2006-01-02") // 在锁内计算today
	stats.Total++
	stats.Daily[today]++
}

func loadConfig() {
	file, err := os.Open("/data/counter/config/config.yaml")
	if err != nil {
		fmt.Println("Failed to open config file:", err)
		fmt.Println("Using default configuration...")
		config = Config{
			Port: "8080",
			File: "/data/counter/count/count.json",
		}
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Failed to decode config file:", err)
		fmt.Println("Using default configuration...")
		config = Config{
			Port: "8080",
			File: "stats.json",
		}
	}
}

func loadStats() {
	file, err := os.Open(config.File)
	if err != nil {
		stats = Stats{
			Daily: make(map[string]int64),
		}
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&stats)
	if err != nil {
		stats = Stats{
			Daily: make(map[string]int64),
		}
	}
}

func saveStats() {
	statsMu.Lock()
	defer statsMu.Unlock()

	file, err := os.Create(config.File)
	if err != nil {
		fmt.Println("Failed to create stats file:", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(stats)
	if err != nil {
		fmt.Println("Failed to marshal stats:", err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Failed to write stats to file:", err)
	}
}

func flushStats() {
	for {
		time.Sleep(10 * time.Second)

		statsMu.Lock()
		bufferLen := len(buffer)
		for i := 0; i < bufferLen; i++ {
			<-buffer
		}
		statsMu.Unlock()

		saveStats()
	}
}

func updateDailyStats() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(next.Sub(now))

		statsMu.Lock()
		today := time.Now().Format("2006-01-02")
		stats.Daily[today] = 0
		statsMu.Unlock()

		saveStats()
	}
}
