package p2p

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/utils"
)

func TestMain(m *testing.M) {
	err := os.Chdir("../")
	utils.PanicOnError(err)
	os.Exit(m.Run())
}

func TestLatency(t *testing.T) {
	mockP2p, _, _ := NewBaseNetwork(logrus.New().WithField("test", "p2ptest"), 1, 2, "")
	total := 0
	totalTimes := 1000
	singleTimes := 1000
	c0 := 1000
	c100 := 0
	for i := 0; i < totalTimes; i++ {
		m := 0
		for j := 0; j < singleTimes; j++ {
			c := mockP2p.timeDelay()
			if m < c {
				m = c
			}
		}
		if c0 > m {
			c0 = m
		}
		if c100 < m {
			c100 = m
		}
		total += m
	}
	avg := total / totalTimes
	t.Logf("avg: %d, min: %d, max: %d", avg, c0, c100)
}
