package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// Server ...
type Server struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	agg      *StatAggregator
	seedFile string
}

// NewServer ...
func NewServer(addr string, seedFile string) *Server {
	s := &Server{
		quit:     make(chan interface{}),
		seedFile: seedFile,
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to Start server on %v", err)
	}
	log.Printf("Server started on %v", addr)

	s.listener = l
	s.wg.Add(1)

	return s
}

// Serve --
func (s *Server) Serve() {
	defer s.wg.Done()

	var vals map[string]chan float64
	var topics []string
	s.initializeValues()

	vals = s.agg.startValueManager(topics) // NOTE: EVERYTHING IS NEW...

	// Initialize Conduit to Stat Agg

	s.acceptConnections(vals)
}

func (s *Server) initializeValues() {

	// Start Stat Manager
	if s.seedFile != "" {

		// Start Stat Manager from file...
		absPath, _ := filepath.Abs(s.seedFile)
		log.Printf("Initializing Values from %v", absPath)

		// File DNE - Create file && init values
		_, err := os.Stat(absPath)

		if os.IsNotExist(err) {
			log.Printf("Seed file %v DNE", absPath)
			log.Printf("Creating seed file %v", absPath)
			file, err := os.Create(absPath)
			if err != nil {
				log.Fatalf("Failed to create seed file %v, %v", absPath, err)
			}
			defer file.Close()

			// No Data in File; Agg from Scratch
			agg := &StatAggregator{
				Data: make(map[string]*stat),
			}
			s.agg = agg
			return
		}

		// File Exists - Read init values from file
		fi, err := os.Open(absPath)
		if err != nil {
			log.Fatalf("Failed to open seed file %v, %v", absPath, err)
		}
		defer fi.Close()
		// Read (small) JSON into map...
		agg := readFromJSON(fi)
		s.agg = &agg
		return
	}

	// Otherwise; Initialize new object from Scratch...
	agg := &StatAggregator{
		Data: make(map[string]*stat),
	}
	s.agg = agg
	return

}

func (s *Server) acceptConnections(vals map[string]chan float64) {

	// for New...
	var newVals map[string]chan float64

	// Accept Connections
	for {

		conn, err := s.listener.Accept()

		// Error...
		if err != nil {

			select {
			case <-s.quit:
				return
			default:
				log.Println("accept error", err)
			}
		} else {
			log.Printf("Accepted connection Values from %v", conn.RemoteAddr().String())
			// Working As Expected...
			go func() {
				defer conn.Close()

				// Push Values via channel
				msg := createMessage(conn)

				// SOL - Topic DNE... ??
				if _, ok := vals[msg.topic]; !ok {
					defer conn.Close()
					// INIT IN DATA IF DNE...
					if _, ok := s.agg.Data[msg.topic]; !ok {
						s.agg.Data[msg.topic] = &stat{}
					}

					// Create Chans...
					newVals = s.agg.startValueManager([]string{msg.topic})
					pushToStatAgg(msg, newVals)
				} else {
					// Push Values
					defer conn.Close()
					pushToStatAgg(msg, vals)
				}

			}()
		}
	}
}

//
func (s *Server) gracefulExit(c chan os.Signal) {
	log.Printf("Shutting down statServer on %v", s.listener.Addr())

	// Write to dump file...
	file, _ := json.MarshalIndent(s.agg, "", " ")
	_ = ioutil.WriteFile(s.seedFile, file, 0644)

	// Shutdown
	close(s.quit)
	close(c)
	s.listener.Close()
	s.wg.Wait()
}

func createMessage(conn net.Conn) message {
	// Init...
	var method string
	var topic string

	sc := bufio.NewScanner(conn)
	sc.Split(bufio.ScanWords)

	// Create MSG
	sc.Scan()
	method = sc.Text()
	sc.Scan()
	topic = sc.Text()

	msg := message{
		method:      method,
		topic:       topic,
		dataScanner: sc,
	}
	log.Printf("Created Message %v %v from %v", method, topic, conn.RemoteAddr().String())

	return msg
}
