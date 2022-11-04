package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	hostName string
	cpuRatio int64
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	hostName, _ = os.Hostname()
}

func work(n int) int {
	if n <= 2 {
		return 1
	}

	return work(n-2) + work(n-1)
}

func compute(w http.ResponseWriter, req *http.Request) {
	log.Printf("request path: %s", req.URL.Path)
	w.Header().Set("X-Fc-Instance-Id", hostName)
	w.Header().Set("My-Cpu-Usage", fmt.Sprintf("%v", atomic.LoadInt64(&cpuRatio)))

	start := time.Now()
	minStr := req.URL.Query().Get("min")
	min, _ := strconv.Atoi(minStr)
	maxStr := req.URL.Query().Get("max")
	max, _ := strconv.Atoi(maxStr)

	msStr := req.URL.Query().Get("ms")
	ms, _ := strconv.Atoi(msStr)

	if max <= min {
		max = min + 1
	}
	n := min + rand.Intn(max-min)
	work(n)
	costMs := time.Now().Sub(start).Nanoseconds() / 1e6
	time.Sleep(time.Duration(ms) * time.Millisecond)
	w.Header().Set("My-Invocation-Duration", fmt.Sprintf("%v", int(costMs)+ms))
	fmt.Fprintf(w, "Hello FC! n=%v, ms=%v, cost=%v ms\n", n, ms, costMs)
}

func getCpuUsage() (int64, error) {
	content, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/cpuacct.usage")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.Trim(string(content), "\n"), 10, 64)
}

func cpuStat() {
	last, err := getCpuUsage()
	if err != nil {
		log.Printf("GET CPU ERROR: %v", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			now, err := getCpuUsage()
			if err != nil {
				log.Printf("GET CPU ERROR: %v", err)
			}
			ratio := (now - last) * 100 / (5 * 1e9)
			log.Printf("cpu usage: %v", ratio)
			atomic.StoreInt64(&cpuRatio, ratio)
			last = now
		}
	}
}

func main() {
	http.HandleFunc("/", compute)
	http.HandleFunc("/compute", compute)

	go cpuStat()

	startTimeStr := os.Getenv("START_TIME_SEC")
	startTimeSec, _ := strconv.Atoi(startTimeStr)
	fmt.Printf("sleep %v sec before start\n", startTimeSec)
	time.Sleep(time.Duration(startTimeSec) * time.Second)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
