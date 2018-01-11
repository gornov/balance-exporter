package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"flag"
	"time"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"fmt"
	"io"
	"log"
	"os"
)

type ApiResponse []Wallet

type Wallet struct {
	AssetId  string
	Balance  float64
	Reserved float64
}

type Exporter struct {
	URI            string
	balanceMetrics map[string]*prometheus.Desc
}

func fetchHTTP(uri string, timeout time.Duration) func() (io.ReadCloser, error) {
	http.DefaultClient.Timeout = timeout

	return func() (io.ReadCloser, error) {
		resp, err := http.DefaultClient.Get(uri)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

func newBalanceMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(*metricsNamespace, "", metricName),
		docString, labels, nil,
	)
}

func NewExporter(uri string) *Exporter {
	return &Exporter{
		URI: uri,
		balanceMetrics: map[string]*prometheus.Desc{
			"balance":  newBalanceMetric("balance", "balance by asset", []string{"asset"}),
			"reserved": newBalanceMetric("reserved", "reserved by asset", []string{"asset"}),
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.balanceMetrics {
		ch <- m
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	body, err := fetchHTTP(e.URI, time.Duration(*scrapeTimeout)*time.Second)()
	if err != nil {
		log.Println("fetchHTTP failed", err)
		return
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("ioutil.ReadAll failed", err)
		return
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(data, &apiResponse)
	if err != nil {
		log.Println("json.Unmarshal failed", err)
		return
	}

	for _, wallet := range apiResponse {
		ch <- prometheus.MustNewConstMetric(e.balanceMetrics["balance"], prometheus.GaugeValue, wallet.Balance, wallet.AssetId)
		ch <- prometheus.MustNewConstMetric(e.balanceMetrics["reserved"], prometheus.GaugeValue, wallet.Reserved, wallet.AssetId)
	}
}

var (
	listenAddress    = flag.String("address", ":9913", "Address on which to expose metrics.")
	metricsEndpoint  = flag.String("metrics_endpoint", "/metrics", "Path under which to expose metrics.")
	metricsNamespace = flag.String("metrics_namespace", "wallet", "Prometheus metrics namespace")
	scrapeURI        = flag.String("scrape_uri", "http://localhost/api/WalletsClientBalances/0", "URI to api page")
	scrapeTimeout    = flag.Int("scrape_timeout", 60, "The number of seconds to wait for an HTTP response from the scrape_timeout")
)

func main() {
	flag.Parse()

	exporter := NewExporter(*scrapeURI)
	prometheus.MustRegister(exporter)
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle(*metricsEndpoint, promhttp.Handler())

	log.Printf("Starting Server at : %s", *listenAddress)
	log.Printf("Metrics endpoint: %s", *metricsEndpoint)
	log.Printf("Metrics namespace: %s", *metricsNamespace)
	log.Printf("Scraping information from : %s", *scrapeURI)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
