package main

import "bufio"

// message - parsed from command line;
// e.g.: echo UPDATE PASSENGERS 134 120 211
type message struct {
	method      string // One of GET / UPDATE
	topic       string // topic to post values to or get values from
	dataScanner *bufio.Scanner
}
