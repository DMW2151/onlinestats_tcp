package main

import (
	"bufio"
	"log"
	"strconv"
	"sync"
)

// checkBytesNumeric - if all bytes in array are numeric...
// e.g. Dec value of byte in (46, 58) exclusive of 47
func checkBytesNumeric(byteArr []byte) bool {

	if len(byteArr) == 0 { // Fail whitespaces...
		return false
	}

	// Check byte decimal maps to . or [0-9]
	// TODO: use map of acceptable bytes...
	for i := range byteArr {
		if (byteArr[i] < 46) || (byteArr[i] > 58) {
			return false
		}
	}
	return true
}

// stageValues - Scans the input and pushes valid entries to intermediate channel
func stageValues(sc *bufio.Scanner, topic string, wg *sync.WaitGroup, jobs chan<- float64) {
	defer wg.Done()

	var valBytes []byte
	var scannedCt int
	var sendCt int
	var sizeBytes int

	for sc.Scan() {

		valBytes = sc.Bytes()

		// Filters to numeric only...
		if checkBytesNumeric(valBytes) {
			fl64, err := strconv.ParseFloat(string(valBytes), 64)
			if err != nil {
				continue
			} else {
				sendCt++
				jobs <- fl64
			}
		}

		// Increment size and count
		sizeBytes += len(valBytes)
		scannedCt++
	}
	log.Printf("Scanned %v obs (%v bytes); send %v values to %v", scannedCt, sizeBytes, sendCt, topic)

	close(jobs)
}

// pushToStatAgg - reads a message; filters; pushes data to correct topic...
func pushToStatAgg(msg message, vals map[string]chan float64) bool {

	// Init Intermediate Channel
	jobs := make(chan float64)

	// Filter to numeric values through intermediate channel, `jobs`
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go stageValues(msg.dataScanner, msg.topic, wg, jobs)

	go func() {
		wg.Wait()
	}()

	// Send to src specific channel
	for stgVal := range jobs {
		vals[msg.topic] <- stgVal
	}

	return true
}
