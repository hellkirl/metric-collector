package main

import (
	"reflect"
	"runtime"
	"testing"
)

func TestMetrics_FillFromMemStats(t *testing.T) {
	type fields struct {
		MemStats    runtime.MemStats
		RandomValue uint64
		PollCount   uint64
	}
	type args struct {
		memStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				MemStats:    tt.fields.MemStats,
				RandomValue: tt.fields.RandomValue,
				PollCount:   tt.fields.PollCount,
			}
			m.FillFromMemStats(tt.args.memStats)
		})
	}
}

func TestMetrics_ToMap(t *testing.T) {
	type fields struct {
		MemStats    runtime.MemStats
		RandomValue uint64
		PollCount   uint64
	}
	type args struct {
		isGauge bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				MemStats:    tt.fields.MemStats,
				RandomValue: tt.fields.RandomValue,
				PollCount:   tt.fields.PollCount,
			}
			if got := m.ToMap(tt.args.isGauge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendMetrics(t *testing.T) {
	type args struct {
		endpoint string
		metrics  map[string]any
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendMetrics(tt.args.endpoint, tt.args.metrics)
		})
	}
}

func Test_startAgent(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startAgent()
		})
	}
}
