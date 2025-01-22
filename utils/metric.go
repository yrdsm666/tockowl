package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Metric struct {
	TotalTransaction    int
	BatchSize           int
	StartTime           time.Time
	TotalLatency        time.Duration
	MinLatency          time.Duration
	MaxLatency          time.Duration
	LatencyNum          int
	TotalEpoch          int
	TotalTimeWithSecond float64
}

func NewMetric(batchSize int) *Metric {
	return &Metric{
		TotalTransaction:    0,
		BatchSize:           batchSize,
		StartTime:           time.Now(),
		TotalLatency:        0,
		MinLatency:          100 * time.Second,
		MaxLatency:          0,
		LatencyNum:          0,
		TotalEpoch:          0,
		TotalTimeWithSecond: 0,
	}
}

func (m *Metric) StartMetric() {
	m.StartTime = time.Now()
}

func (m *Metric) ThroughputMeasurement(txs []byte) {
	if txs == nil {
		return
	}

	allTx := [][]byte{}
	err := json.Unmarshal(txs, &allTx)
	if err != nil {
		log.Printf("Unmarshal txs error: %v", err)
		return
	}
	m.TotalTransaction += len(allTx)
}

func (m *Metric) AddTotalTransaction(txNum int) {
	m.TotalTransaction += txNum
}

func (m *Metric) AddTotalEpoch() {
	m.TotalEpoch += 1
	m.TotalTimeWithSecond = float64(time.Since(m.StartTime)) / 1000 / 1000 / 1000
}

func (m *Metric) LatencyMeasurement(latency time.Duration) {
	if latency < 0 {
		return
	}

	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}

	if latency < m.MinLatency {
		m.MinLatency = latency
	}

	m.TotalLatency += latency
	m.LatencyNum++
}

func (m *Metric) PrintMetric() {
	// TotalTimeWithSecond := float64(time.Since(m.StartTime)) / 1000 / 1000 / 1000
	// TotalTimeWithSecond := float64(m.TotalLatency) / 1000 / 1000 / 1000

	// MinLatencyWithSecond := float64(m.MinLatency/time.Millisecond) / 1000
	// if m.TotalEpoch != 0 {
	// 	fmt.Printf("Throughput: %f\nAvgLatency: %d ms\nMinLatency: %d ms\nMaxLatency: %d ms\nPeakThroughput %f\nEpoch: %d\n",
	// 		// float64(m.TotalTransaction)/m.TotalTimeWithSecond,
	// 		float64(m.TotalTransaction)/m.TotalTimeWithSecond,
	// 		m.TotalLatency/time.Duration(m.LatencyNum)/time.Millisecond,
	// 		m.MinLatency/time.Millisecond,
	// 		m.MaxLatency/time.Millisecond,
	// 		float64(m.BatchSize)/MinLatencyWithSecond,
	// 		m.TotalEpoch)
	// } else {
	// 	fmt.Printf("Throughput: %f\nAvgLatency: %d ms\nMinLatency: %d ms\nMaxLatency: %d ms\nPeakThroughput %f\nEpoch: %d\n",
	// 		float64(m.TotalTransaction)/m.TotalTimeWithSecond,
	// 		0,
	// 		m.MinLatency/time.Millisecond,
	// 		m.MaxLatency/time.Millisecond,
	// 		float64(m.BatchSize)/MinLatencyWithSecond,
	// 		m.TotalEpoch)
	// }
	fmt.Println("\n=========== RESULTS ===========")
	if m.TotalEpoch != 0 {
		fmt.Printf("Throughput: %f\nAvgLatency: %d ms\nEpoch: %d\n",
			// float64(m.TotalTransaction)/m.TotalTimeWithSecond,
			float64(m.TotalTransaction)/m.TotalTimeWithSecond,
			m.TotalLatency/time.Duration(m.LatencyNum)/time.Millisecond,
			m.TotalEpoch)
	} else {
		fmt.Printf("Throughput: %f\nAvgLatency: %d ms\nEpoch: %d\n",
			float64(m.TotalTransaction)/m.TotalTimeWithSecond,
			0,
			m.TotalEpoch)
	}
	fmt.Println("=========== RESULTS ===========\n")
}
