package main

import (
	"log"
	"sync"
)

// stat - record for each topic
type stat struct {
	Count    float64 `json:"count"`    // N-Obs in Stream
	Average  float64 `json:"average"`  // Average over Stream - From Start
	Variance float64 `json:"variance"` // Variance over Stream - From Start
}

// StatAggregator - Contains statistics `stat` for each "topic"
type StatAggregator struct {
	Data map[string]*stat `json:"data"`
}

// updateState - Updates statistic state for given topic; Calculation Reference:
// https://nestedsoftware.com/2018/03/20/calculating-a-moving-average-on-streaming-data-5a7k.22879.html
func (sa *StatAggregator) updateState(topic string, new float64) {

	var prevAvg float64 = sa.Data[topic].Average

	// Count - Increment
	sa.Data[topic].Count++

	// Average - Recurrance relationship b/t Avg_n-1 and Avg
	// prevents division on huge #s or floating point issues w. large sums & counts...
	if sa.Data[topic].Count > 1 {
		sa.Data[topic].Average += (new - sa.Data[topic].Average) / sa.Data[topic].Count
	} else {
		sa.Data[topic].Average = new
	}

	// Variance - Recurrance relationship b/t S_n-1 and S; depends on dn2 (variance * (n-1))
	// Prevents division on huge #s or floating point issues
	if sa.Data[topic].Count > 1 {
		var dn2 float64 = sa.Data[topic].Variance * sa.Data[topic].Count
		sa.Data[topic].Variance = (dn2 + ((new - sa.Data[topic].Average) * (new - prevAvg))) / sa.Data[topic].Count
	} else {
		sa.Data[topic].Variance = 0.0
	}

}

// listenTopic - receives from channel and updates topic statistics...
func (sa *StatAggregator) listenTopic(wg *sync.WaitGroup, topic string, topicStream chan float64) {
	defer wg.Done()

	// Recv value from topicStream and call updateState...
	for val := range topicStream {
		sa.updateState(topic, val)
	}
}

// startValueManager - Initializes listenTopic on
func (sa *StatAggregator) startValueManager(topics []string) (values map[string]chan float64) {

	// Make a Topic for Each Value Manager
	values = make(map[string]chan float64)

	if topics == nil {

		// If no topics passed; initialize from Data
		var topicNames []string
		for t := range sa.Data {
			topicNames = append(topicNames, t)
		}

		log.Printf("Initialized value Manager - %v", topicNames) //NOTE: FIX...
		for topic := range sa.Data {
			values[topic] = make(chan float64)
		}

	} else {
		// Otherwise; from topics
		log.Printf("Initialized value Manager - %v", topics)
		for _, topic := range topics {
			values[topic] = make(chan float64)
		}
	}

	wg := new(sync.WaitGroup)

	for topic, topicStream := range values {
		wg.Add(1)
		go sa.listenTopic(wg, topic, topicStream)
	}

	return values
}
