package flashback

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func TestStatsAnalyzer(t *testing.T) {
	TestingT(t)
}

type TestStatsAnalyzerSuite struct{}

var _ = Suite(&TestStatsAnalyzerSuite{})

func (s *TestStatsAnalyzerSuite) TestBasics(c *C) {
	opsExecuted := int64(0)
	latencyChan := make(chan Latency)

	analyser := NewStatsAnalyzer(
		[]*StatsCollector{}, &opsExecuted, latencyChan, 1000,
	)

	for _, latencyList := range analyser.latencies {
		c.Assert(latencyList, HasLen, 0)
	}
	for i := 0; i < 10; i += 1 {
		for _, opType := range AllOpTypes {
			latencyChan <- Latency{opType, time.Duration(i)}
		}
	}
	close(latencyChan)
	<-analyser.finished
	for _, latencyList := range analyser.latencies {
		c.Assert(latencyList, HasLen, 10)
	}
}

func (s *TestStatsAnalyzerSuite) TestLatencies(c *C) {
	opsExecuted := int64(0)
	latencyChan := make(chan Latency)

	analyser := NewStatsAnalyzer(
		[]*StatsCollector{}, &opsExecuted, latencyChan, 1000,
	)

	start := 1000
	for _, opType := range AllOpTypes {
		for i := 100; i >= 0; i-- {
			latencyChan <- Latency{opType, time.Duration(start + i)}
		}
		start += 2000
	}

	// ugly hack because GetStatus races with latencyChan being consumed
	time.Sleep(10)
	status := analyser.GetStatus()

	// Check results
	start = 1000
	for _, opType := range AllOpTypes {
		sinceLast := status.SinceLastLatencies[opType]
		allTime := status.AllTimeLatencies[opType]
		for i, perc := range latencyPercentiles {
			c.Assert(sinceLast[i], Equals, int64(perc+start))
			c.Assert(allTime[i], Equals, int64(perc+start))
		}
		start += 2000
	}

	// -- second round
	start = 2000
	for _, opType := range AllOpTypes {
		for i := 100; i >= 0; i-- {
			latencyChan <- Latency{opType, time.Duration(start + i)}
		}
		start += 2000
	}

	close(latencyChan)
	<-analyser.finished
	status = analyser.GetStatus()

	start = 2000
	for _, opType := range AllOpTypes {
		sinceLast := status.SinceLastLatencies[opType]
		allTime := status.AllTimeLatencies[opType]
		for i, perc := range latencyPercentiles {
			c.Assert(sinceLast[i], Equals, int64(perc+start))
		}
		c.Assert(allTime[len(allTime)-1], Equals, int64(start+100))
		c.Assert(allTime[0], Equals, int64(start-1000+100))
		start += 2000
	}
}
