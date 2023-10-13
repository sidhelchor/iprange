package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func showBanner() {
	fmt.Println("Mr. X's Shark IP Ranger & Live IP Collector V10.2023")
}

func getPassword() string {
	fmt.Print("Enter the password: ")
	password := ""
	_, _ = fmt.Scanln(&password)
	return password
}

func checkIPList(inputFile, outputFile string, concurrentChecks int, timeoutSeconds int) {
	var mu sync.Mutex
	var totalIPs, liveIPs int

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		return
	}
	defer file.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer output.Close()

	scanner := bufio.NewScanner(file)
	// Create a buffered channel for concurrent checks
	semaphore := make(chan struct{}, concurrentChecks)

	for scanner.Scan() {
		ip := scanner.Text()
		ipParts := strings.Split(ip, ".")
		if len(ipParts) != 4 {
			fmt.Printf("Invalid IP address format: %s\n", ip)
			continue
		}

		baseIP := strings.Join(ipParts[:3], ".")
		totalIPs += 255

		for i := 1; i <= 255; i++ {
			semaphore <- struct{}{} // Acquire a semaphore
			go func(baseIP string, i int) {
				defer func() { <-semaphore }() // Release the semaphore
				ipToCheck := fmt.Sprintf("%s.%d", baseIP, i)
				if checkLiveIPWithTimeout(ipToCheck, timeoutSeconds) {
					mu.Lock()
					liveIPs++
					fmt.Printf("Live IP (%d/%d): %s\n", liveIPs, totalIPs, ipToCheck)
					_, _ = output.WriteString(ipToCheck + "\n")
					mu.Unlock()
				} else {
					fmt.Printf("Checking IP (%d/%d): %s\n", liveIPs, totalIPs, ipToCheck)
				}
			}(baseIP, i)
		}
	}

	fmt.Printf("Total IPs Checked: %d\n", totalIPs)
	fmt.Printf("Total Live IPs Collected: %d\n", liveIPs)
	fmt.Printf("Unique live IPs saved to %s\n", outputFile)

	showBanner() // Display the banner again
}

func checkLiveIPWithTimeout(ip string, timeoutSeconds int) bool {
	conn, err := net.DialTimeout("tcp", ip+":80", time.Duration(timeoutSeconds)*time.Second)
	if err == nil {
		defer conn.Close()
		return true
	}
	return false
}

func main() {
	showBanner() // Display the banner at the beginning

	var password, userPassword string

	password = "404"

	for userPassword != password {
		userPassword = getPassword()
		if userPassword != password {
			fmt.Println("Incorrect password. Try again.")
		}
	}

	var inputFile, outputFile string
	var concurrentChecks, timeoutSeconds int

	fmt.Print("Enter the name of the input file with IP addresses: ")
	_, _ = fmt.Scanln(&inputFile)

	fmt.Print("Enter the name of the output file to save unique live IPs: ")
	_, _ = fmt.Scanln(&outputFile)

	fmt.Print("Enter the number of concurrent checks (e.g., 100): ")
	_, _ = fmt.Scanln(&concurrentChecks)

	fmt.Print("Enter the timeout in seconds for live IP checks (e.g., 5): ")
	_, _ = fmt.Scanln(&timeoutSeconds)

	checkIPList(inputFile, outputFile, concurrentChecks, timeoutSeconds)
}
