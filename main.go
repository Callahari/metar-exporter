package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	//metars := flag.String("stations", "", "")
	var logLevelParam string
	flag.StringVar(&logLevelParam, "loglevel", "info", "debug,info,warning,error")
	var listen string
	flag.StringVar(&listen, "listen", "localhost:9093", "<ip>:<port>")
	var stations string
	flag.StringVar(&stations, "stations", "EDDH", "ICAO-Code (Mehrere Station Komma getrennt)")
	flag.Parse()

	logLevel, err := log.ParseLevel(logLevelParam)
	if err != nil {
		log.Warning("Unknown LogLevel, use Info")
		logLevel, _ = log.ParseLevel("info")
	}
	log.SetLevel(logLevel)

	metar := newMetarCollector(StationsToArray(stations))
	prometheus.MustRegister(metar)

	http.Handle("/metrics", promhttp.Handler())
	log.WithFields(log.Fields{
		"listen":   listen,
		"stations": stations,
	}).Info("Prometheus METAR exporter")
	http.ListenAndServe(listen, nil)
}
