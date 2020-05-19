package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// readFromJSON - Reads JSON on Disk into StatAggregaator on stat server startup
func readFromJSON(fi *os.File) (initVals StatAggregator) {
	byteValue, _ := ioutil.ReadAll(fi)
	json.Unmarshal(byteValue, &initVals)

	return
}

// Get ENV VARS w. Default...
func getEnvDefault(key string, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultVal
}
