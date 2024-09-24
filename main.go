package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
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
)

func main() {
	loadConfig()
	loadStats()

	router := gin.Default()

	router.GET("/add", handleRequest)
	router.GET("/api/counter", handleRead)
	router.GET("/api/counter/daily", handleDaily)
	router.GET("/api/counter/total", handleTotal)

	go backgroundProcess()

	if err := router.Run(fmt.Sprintf(":%s", config.Port)); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleRequest(c *gin.Context) {
	statsMu.Lock()
	defer statsMu.Unlock()

	today := time.Now().Format("2006-01-02")
	stats.Total++
	stats.Daily[today]++

	c.String(200, "Request received!")
}

func handleRead(c *gin.Context) {
	statsMu.Lock()
	data, _ := json.Marshal(stats)
	statsMu.Unlock()

	c.Data(200, "application/json", data)
}

func handleDaily(c *gin.Context) {
	statsMu.Lock()
	today := time.Now().Format("2006-01-02")
	count := stats.Daily[today]
	statsMu.Unlock()

	c.String(200, fmt.Sprintf("%d", count))
}

func handleTotal(c *gin.Context) {
	statsMu.Lock()
	total := stats.Total
	statsMu.Unlock()

	c.String(200, fmt.Sprintf("%d", total))
}

func loadConfig() {
	configFilePath := os.Getenv("CONFIG_PATH")
	if configFilePath == "" {
		configFilePath = "/data/counter/config/config.yaml"
	}
	file, err := os.Open(configFilePath)
	if err != nil {
		log.Println("Failed to open config file:", err)
		log.Println("Using default configuration...")
		config = Config{
			Port: "8080",
			File: "/data/counter/count/count.json",
		}
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Println("Failed to decode config file:", err)
		log.Println("Using default configuration...")
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
		log.Println("Failed to create stats file:", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(stats)
	if err != nil {
		log.Println("Failed to marshal stats:", err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		log.Println("Failed to write stats to file:", err)
	}
}

func backgroundProcess() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			saveStats()

		case <-time.After(time.Until(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 0, 0, 0, 0, time.Now().Location()))):
			statsMu.Lock()
			today := time.Now().Format("2006-01-02")
			stats.Daily[today] = 0
			statsMu.Unlock()
			saveStats()
		}
	}
}
