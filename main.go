package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
)

var (
	flag_test = flag.Bool("test", false, "print all available metrics to stdout")
	flag_addr = flag.String("listen-address", ":9111", "The address to listen on for HTTP requests.")

	flag_gateway_address = flag.String("gateway-address", "fritz.box", "The URL of the upnp service")
	flag_gateway_port    = flag.Int("gateway-port", 49000, "The URL of the upnp service")
)

var (
	collect_errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fritzbox_exporter_collect_errors",
		Help: "Number of collection errors.",
	})
)

type UpnpMetric struct {
	upnp.UpnpValueUint
	*prometheus.Desc
}

func (m UpnpMetric) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.Desc
}

func (m UpnpMetric) Collect(gateway string, port uint16, ch chan<- prometheus.Metric) error {
	val, err := m.Query(gateway, port)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		m.Desc,
		prometheus.CounterValue,
		float64(val),
		gateway,
	)
	return nil
}

func NewUpnpMetric(v upnp.UpnpValueUint) UpnpMetric {
	return UpnpMetric{
		v,
		prometheus.NewDesc(
			prometheus.BuildFQName("gateway", "wan", v.ShortName),
			v.Help,
			[]string{"gateway"},
			nil,
		),
	}
}

type FritzboxCollector struct {
	gateway string
	port    uint16
	metrics []UpnpMetric
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range fc.metrics {
		m.Describe(ch)
	}
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range fc.metrics {
		err := m.Collect(fc.gateway, fc.port, ch)
		if err != nil {
            collect_errors.Inc()
		}
	}
}

func main() {
	flag.Parse()

	if *flag_test {
		for _, v := range upnp.Values {
			res, err := v.Query(*flag_gateway_address, uint16(*flag_gateway_port))
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s: %d\n", v.ShortName, res)
		}
		return
	}

	metrics := make([]UpnpMetric, len(upnp.Values))
	for _, v := range upnp.Values {
		metrics = append(metrics, NewUpnpMetric(v))
	}

	prometheus.MustRegister(&FritzboxCollector{
		*flag_gateway_address,
		uint16(*flag_gateway_port),
		metrics,
	})
	// Since we are dealing with custom Collector implementations, it might
	// be a good idea to enable the collect checks in the registry.
	prometheus.EnableCollectChecks(true)

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(*flag_addr, nil)
}
