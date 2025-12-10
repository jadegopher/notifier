package internal

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkAggregator_Handle(b *testing.B) {
	const (
		inputChanSize     = 100
		outputChanSize    = 100
		maxBatchSizeBytes = 1024 * 10 // 10KB batch
		flushInterval     = 500 * time.Millisecond
	)

	// Create channels
	inputChan := make(chan string, inputChanSize)

	// Initialize Aggregator
	// Note: Assuming NewAggregator sets up the internal batch and other fields correctly
	agg := NewAggregator(inputChan, outputChanSize, maxBatchSizeBytes, flushInterval)

	// Sample message payload
	msg := "benchmark_payload_data_string_normal_sized_string_less_than_120_symbols_but_pretty_average_readable_string"
	msgLen := len(msg)

	var wg sync.WaitGroup

	go func() {
		for range agg.OutputChan() {
			// Intentionally discard output for benchmark
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agg.Handle()
	}()

	b.ReportAllocs()
	b.SetBytes(int64(msgLen))
	b.ResetTimer() // Start timing only after setup is complete

	for i := 0; i < b.N; i++ {
		inputChan <- msg
	}

	b.StopTimer() // Stop timing before teardown
	close(inputChan)
	wg.Wait()
}

func BenchmarkAggregator_Handle_Parallel(b *testing.B) {
	inputChan := make(chan string, 1000)
	agg := NewAggregator(inputChan, 100, 1024*10, 1*time.Minute)

	// Drain output
	go func() {
		for range agg.OutputChan() {
		}
	}()

	// Start Handle
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agg.Handle()
	}()

	msg := "benchmark_payload_data_string_normal_sized_string_less_than_120_symbols_but_pretty_average_readable_string"
	msgLen := len(msg)

	b.ReportAllocs()
	b.SetBytes(int64(msgLen))
	b.ResetTimer()

	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				inputChan <- msg
			}
		},
	)
	b.StopTimer()

	close(inputChan)
	wg.Wait()
}

func TestAggregator_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		data       []string
		output     [][]string
		sleepTime  time.Duration
		aggregator Aggregator
	}{
		{
			name: "flush_by_overflow",
			data: []string{"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3"},
			aggregator: Aggregator{
				outputChan:    make(chan []string, 10),
				flushInterval: time.Second,
				batch:         newBatch(5),
			},
			sleepTime: 100 * time.Millisecond,
			output:    [][]string{{"1", "1", "1", "1", "1"}, {"2", "2", "2", "2", "2"}, {"3", "3", "3", "3", "3"}},
		},
		{
			name: "flush_by_overflow_timer_reset",
			data: []string{
				"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3",
				"3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3",
				"3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3", "3",
			},
			aggregator: Aggregator{
				outputChan:    make(chan []string, 10),
				flushInterval: 100 * time.Microsecond,
				batch:         newBatch(5),
			},
			sleepTime: 100 * time.Millisecond,
			output: [][]string{
				{"1", "1", "1", "1", "1"}, {"2", "2", "2", "2", "2"}, {"3", "3", "3", "3", "3"},
				{"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"},
				{"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"},
				{"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"}, {"3", "3", "3", "3", "3"},
			},
		},
		{
			name: "flush_by_timer",
			data: []string{"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3"},
			aggregator: Aggregator{
				outputChan:    make(chan []string, 10),
				flushInterval: 100 * time.Millisecond,
				batch:         newBatch(500),
			},
			sleepTime: 200 * time.Millisecond,
			output:    [][]string{{"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3"}},
		},
		{
			name: "flush_by_exit_flush",
			data: []string{"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3"},
			aggregator: Aggregator{
				outputChan:    make(chan []string, 10),
				flushInterval: 1000 * time.Millisecond,
				batch:         newBatch(500),
			},
			sleepTime: 0,
			output:    [][]string{{"1", "1", "1", "1", "1", "2", "2", "2", "2", "2", "3", "3", "3", "3", "3"}},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				inputChan := make(chan string, 10)
				tt.aggregator.inputChan = inputChan

				wg := &sync.WaitGroup{}
				result := make([][]string, 0, len(tt.output))

				wg.Add(2)
				go func() {
					defer wg.Done()
					tt.aggregator.Handle()
				}()

				go func() {
					defer wg.Done()

					ch := tt.aggregator.OutputChan()
					for data := range ch {
						result = append(result, data)
					}
				}()

				for _, data := range tt.data {
					inputChan <- data
				}

				time.Sleep(tt.sleepTime)

				close(inputChan)

				wg.Wait()

				if diff := cmp.Diff(result, tt.output); diff != "" {
					t.Errorf("Aggregate() mismatch (-want +got):\n%s %s", diff, result)
					return
				}
			},
		)
	}
}
