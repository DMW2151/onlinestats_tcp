package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	/* Test
	-- Get Data (15,857,625 obs) and pre-cut to a single sample column...
	wget http://sdm.lbl.gov/fastbit/data/star2002-full.csv.gz && gunzip star2002-full.csv.gz
	cat ./data/star2002-full.csv | cut -d',' -f3 > ./data/star2002-no-cut.csv

	-- Sample run on local...
	go run ./src/
	time echo UPDATE SPACE | cat - ./data/star2002-no-cut.csv | nc 192.168.1.4 27001
	*/

	statServerAddress := getEnvDefault("STATSERVER_ADDR", "192.168.1.5:27001")
	statServerDump := getEnvDefault("STATSERVER_DUMP", "./dumps/test.json")
	statServerLogFile := getEnvDefault("STATSERVER_LOG", "./logs/test.log")

	// Ensure logging directory...
	absPath, err := filepath.Abs(statServerLogFile)
	if err != nil {
		log.Panicf("Absolute path %v is not valid - %v", absPath, err)
	}

	// Create if DNE...
	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Panicf("Failed to open logfile %v, %v", absPath, err)
	}
	defer file.Close()

	log.SetOutput(file)

	// Start Server...
	s := NewServer(statServerAddress, statServerDump)
	go s.Serve()

	// Allow for "graceful" shutdown w. Ctrl+C; write file before quit...
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	for range c {
		s.gracefulExit(c)
	}

}
