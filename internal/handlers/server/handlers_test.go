package server

import (
	"devops_analytics/internal/storage"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHomePage(t *testing.T) {
	type want struct {
		response    string
		contentType string
		statusCode  int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "expected result",
			want: want{
				response:    "Hello World!",
				contentType: "text/plain",
				statusCode:  http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			HomePage(w, r)

			res := w.Result()
			assert.Equal(t, res.Header.Get("Content-Type"), test.want.contentType)
			assert.Equal(t, res.StatusCode, test.want.statusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}

func TestUpdateMetricHandler(t *testing.T) {
	type args struct {
		metricType, metricName string
		metricValue            string
		want                   string
	}
	tests := []struct {
		name string
		ms   *storage.MemStorage
		args []args
	}{
		{
			name: "gauge expected behavior",
			ms:   storage.NewMemStorageHandler(),
			args: []args{
				{
					metricType:  "gauge",
					metricName:  "Alloc",
					metricValue: "12.53523",
					want:        "12.53523",
				}, {
					metricType:  "gauge",
					metricName:  "Alloc",
					metricValue: "12535322353",
					want:        "12535322353",
				},
				{
					metricType:  "gauge",
					metricName:  "Alloc",
					metricValue: "-0.000045",
					want:        "-0.000045",
				},
			},
		},
		{
			name: "counter expected behavior",
			ms:   storage.NewMemStorageHandler(),
			args: []args{
				{
					metricType:  "counter",
					metricName:  "TotalAlloc",
					metricValue: "124",
					want:        "124",
				}, {
					metricType:  "counter",
					metricName:  "TotalAlloc",
					metricValue: "1355",
					want:        "1479",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, arg := range test.args {
				nestedHandler := UpdateMetricHandler(test.ms)
				r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%s", arg.metricType, arg.metricName, arg.metricValue), nil)
				w := httptest.NewRecorder()
				nestedHandler(w, r)

				res := w.Result()
				res.Body.Close()
				assert.Equal(t, http.StatusOK, res.StatusCode)
				assert.Equal(t, res.Header.Get("Content-Type"), "text/plain")

				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, "", string(resBody))

				switch arg.metricType {
				case "gauge":
					assert.Equal(t, arg.want, strconv.FormatFloat(test.ms.GetMetrics().Gauges[arg.metricName], 'f', -1, 64))
				case "counter":
					assert.Equal(t, arg.want, strconv.Itoa(int(test.ms.GetMetrics().Counters[arg.metricName])))
				}
			}
		})
	}
}
