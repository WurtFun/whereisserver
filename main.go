package main

import (
	"flag"
	"fmt"
	"net"
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

func parseIP(ipStr string) ([]int, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("not an IPv4 address: %s", ipStr)
	}
	return []int{int(ip[0]), int(ip[1]), int(ip[2]), int(ip[3])}, nil
}

func ipToString(ip []int) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func iterateIPRange(startIP, endIP []int, ipChannel chan<- string, wg *sync.WaitGroup, sem chan struct{}, logFile *os.File) {
	for i := startIP[0]; i <= endIP[0]; i++ {
		for j := startIP[1]; j <= endIP[1]; j++ {
			for k := startIP[2]; k <= endIP[2]; k++ {
				for l := startIP[3]; l <= endIP[3]; l++ {
					ip := []int{i, j, k, l}
					ipStr := ipToString(ip)
					if isPrivateIP(ipStr) {
						continue
					}

					wg.Add(1)
					sem <- struct{}{}

					time.Sleep(1000 * time.Millisecond)

					go func(ipStr string) {
						defer wg.Done()
						defer func() { <-sem }()

						output, err := exec.Command("go", "run", "mcping/mcping.go", ipStr).Output()
						if err != nil {
							fmt.Println(ipStr, err, "No data.")
							return
						}
						ipChannel <- fmt.Sprintf("Server address: %s, Back: %s\n", ipStr, strings.TrimSpace(string(output)))
					}(ipStr)
				}
			}
		}
	}
}

func main() {
	startIPStr := flag.String("start", "0.0.0.0")
	endIPStr := flag.String("end", "255.255.255.255")
	flag.Parse()

	startIP, err := parseIP(*startIPStr)
	if err != nil {
		fmt.Println("Error parsing start IP:", err)
		return
	}
	endIP, err := parseIP(*endIPStr)
	if err != nil {
		fmt.Println("Error parsing end IP:", err)
		return
	}

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

	iterateIPRange(startIP, endIP, ipChannel, &wg, sem, logFile)

	wg.Wait()
	close(ipChannel)
}
