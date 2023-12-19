package main

import (
	"encoding/json"
	metarParser "github.com/eugecm/gometar/metar/parser"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
)

type MetarData []struct {
	MetarID     int     `json:"metar_id"`
	IcaoID      string  `json:"icaoId"`
	ReceiptTime string  `json:"receiptTime"`
	ObsTime     int     `json:"obsTime"`
	ReportTime  string  `json:"reportTime"`
	Temp        int     `json:"temp"`
	Dewp        int     `json:"dewp"`
	Wdir        int     `json:"wdir"`
	Wspd        int     `json:"wspd"`
	Wgst        any     `json:"wgst"`
	Visib       string  `json:"visib"`
	Altim       int     `json:"altim"`
	Slp         any     `json:"slp"`
	QcField     int     `json:"qcField"`
	WxString    string  `json:"wxString"`
	PresTend    any     `json:"presTend"`
	MaxT        any     `json:"maxT"`
	MinT        any     `json:"minT"`
	MaxT24      any     `json:"maxT24"`
	MinT24      any     `json:"minT24"`
	Precip      any     `json:"precip"`
	Pcp3Hr      any     `json:"pcp3hr"`
	Pcp6Hr      any     `json:"pcp6hr"`
	Pcp24Hr     any     `json:"pcp24hr"`
	Snow        any     `json:"snow"`
	VertVis     any     `json:"vertVis"`
	MetarType   string  `json:"metarType"`
	RawOb       string  `json:"rawOb"`
	MostRecent  int     `json:"mostRecent"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Elev        int     `json:"elev"`
	Prior       int     `json:"prior"`
	Name        string  `json:"name"`
	Clouds      []struct {
		Cover string `json:"cover"`
		Base  int    `json:"base"`
	} `json:"clouds"`
}
type metarCollector struct {
	metarTemp  *prometheus.Desc
	metarDewp  *prometheus.Desc
	metarWdir  *prometheus.Desc
	metarWspd  *prometheus.Desc
	metarAltim *prometheus.Desc
	metarVisib *prometheus.Desc
	metarWx    *prometheus.Desc
	execTime   *prometheus.Desc
	stations   []string
}

func newMetarCollector(stations []string) *metarCollector {
	return &metarCollector{
		metarTemp:  prometheus.NewDesc("metar_temperature", "Temperature in Celsius", []string{"station"}, nil),
		metarDewp:  prometheus.NewDesc("metar_dewPoint", "Dew Point in Celsius", []string{"station"}, nil),
		metarWdir:  prometheus.NewDesc("metar_wind_direction", "Wind Direction in degrees", []string{"station"}, nil),
		metarWspd:  prometheus.NewDesc("metar_wind_speed", "Wind speed in Kt", []string{"station"}, nil),
		metarAltim: prometheus.NewDesc("metar_altim", "Altimeter in Hg", []string{"station"}, nil),
		metarVisib: prometheus.NewDesc("metar_visibilityStatute", "Sichtweite in Miles", []string{"station"}, nil),
		metarWx:    prometheus.NewDesc("metar_wx", "Aktelles Wetter", []string{"station"}, nil),
		execTime:   prometheus.NewDesc("metar_execTime", "func runtime", nil, nil),
		stations:   stations,
	}
}

func (collector *metarCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.metarTemp
	ch <- collector.metarDewp
	ch <- collector.metarWdir
	ch <- collector.metarWspd
	ch <- collector.metarAltim
	ch <- collector.metarVisib
	ch <- collector.metarWx
	ch <- collector.execTime

}

func (c *metarCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	//var metricValue float64
	//https://aviationweather.gov/cgi-bin/data/metar.php?ids=EDDH,EDDM&format=json&taf=false

	req, err := http.NewRequest("GET", "https://aviationweather.gov/cgi-bin/data/metar.php?ids="+StationsArrayToIDs(c.stations)+"&format=json&taf=false", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0")
	q := req.URL.Query()
	//q.Add("includeTimeseries", "true")
	//q.Add("includeCurrentMeasurement", "true")
	req.URL.RawQuery = q.Encode()

	res1, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res1.StatusCode != 200 {
		log.WithFields(log.Fields{
			"statusCode": res1.StatusCode,
			"status":     res1.Status,
		}).Warn("Bad StatusCode von aviationweather.gov")
		return
	}
	resBody, _ := io.ReadAll(res1.Body)
	data := MetarData{}
	json.Unmarshal(resBody, &data)
	if len(c.stations) != len(data) {
		log.WithFields(log.Fields{
			"angefragt": len(c.stations),
			"empfangen": len(data),
		}).Warn("Es wurden nicht alle Stationen abgerufen")
	}
	for _, station := range data {
		log.WithFields(log.Fields{
			"station":    station.Name,
			"IACO":       station.IcaoID,
			"temp":       station.Temp,
			"dewp":       station.Dewp,
			"wind.speed": station.Wspd,
			"wind.dir":   station.Wdir,
			"altim":      station.Altim,
			"visibility": station.Visib,
			"wx":         station.WxString,
		}).Debug("Daten empfangen")
		metarP := metarParser.New()
		metarraw, _ := metarP.Parse(station.RawOb)
		vis, _ := strconv.Atoi(metarraw.Visibility.Distance)

		m1 := prometheus.MustNewConstMetric(c.metarTemp, prometheus.GaugeValue, float64(station.Temp), station.Name)
		m2 := prometheus.MustNewConstMetric(c.metarDewp, prometheus.GaugeValue, float64(station.Dewp), station.Name)
		m3 := prometheus.MustNewConstMetric(c.metarWspd, prometheus.GaugeValue, float64(station.Wspd), station.Name)
		m4 := prometheus.MustNewConstMetric(c.metarWdir, prometheus.GaugeValue, float64(station.Wdir), station.Name)
		m5 := prometheus.MustNewConstMetric(c.metarAltim, prometheus.GaugeValue, float64(station.Altim), station.Name)
		m6 := prometheus.MustNewConstMetric(c.metarVisib, prometheus.GaugeValue, float64(vis), station.Name)
		//m3 := prometheus.MustNewConstMetric(c.metarWx, prometheus.GaugeValue, float64(metarraw.Weather.), station.Name)
		ch <- m1
		ch <- m2
		ch <- m3
		ch <- m4
		ch <- m5
		ch <- m6
	}
	elapsed := time.Since(start)
	m7 := prometheus.MustNewConstMetric(c.execTime, prometheus.GaugeValue, float64(elapsed))
	ch <- m7
	log.WithField("exec.elapsed", elapsed).Debug("Func Runtime")
}
