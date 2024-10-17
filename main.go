package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func isPrivateIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	firstOctet, _ := strconv.Atoi(parts[0])
	secondOctet, _ := strconv.Atoi(parts[1])

	if firstOctet == 10 || (firstOctet == 172 && secondOctet >= 16 && secondOctet <= 31) || (firstOctet == 192 && secondOctet == 168) {
		return true
	}
	return false
}

func main() {
	logDir := "catch-server-list"
	logFileName := "laster-catch.txt"
	logFilePath := fmt.Sprintf("%s/%s", logDir, logFileName)

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err = os.Mkdir(logDir, 0755)
		if err != nil {
			fmt.Println("log error:", err)
			return
		}
	}

	timestamp := time.Now().Format("20060102150405")
	oldLogFileName := fmt.Sprintf("%s/catched_%s.txt", logDir, timestamp)

	if _, err := os.Stat(logFilePath); err == nil {
		err = os.Rename(logFilePath, oldLogFileName)
		if err != nil {
			fmt.Println("log error:", err)
			return
		}
	}

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("log error:", err)
		return
	}
	defer logFile.Close()

	ipChannel := make(chan string, 4)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)

	go func() {
		for result := range ipChannel {
			_, err := logFile.WriteString(result)
			if err != nil {
				fmt.Println("log error:", err)
			}
		}
	}()

	for i := 110; i < 256; i++ {
		for j := 42; j < 256; j++ {
			for k := 1; k < 256; k++ {
				for l := 1; l < 256; l++ {
					ip := fmt.Sprintf("%d.%d.%d.%d", i, j, k, l)
					if isPrivateIP(ip) {
						continue
					}

					wg.Add(1)
					sem <- struct{}{}

					time.Sleep(1000 * time.Millisecond)

					go func(ip string) {
						defer wg.Done()
						defer func() { <-sem }()

						output, err := exec.Command("go", "run", "mcping/mcping.go", ip).Output()
						if err != nil {
							fmt.Println(ip, err, "No data.")
							return
						}
						ipChannel <- fmt.Sprintf("Server address: %s, Back: %s\n", ip, strings.TrimSpace(string(output)))
					}(ip)
				}
			}
		}
	}

	wg.Wait()
	close(ipChannel)
}
