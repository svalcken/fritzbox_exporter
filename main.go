package main

import (
	"flag"
	"net/http"
    "strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	flag_addr        = flag.String("listen-address", ":9111", "The address to listen on for HTTP requests.")
	flag_dev_address = flag.String("device-address", "fritz.box", "The URL of the upnp service")
)

var (
	WAN_IP = UpnpValue{
		path:    "/igdupnp/control/WANIPConn1",
		service: "WANIPConnection:1",
		method:  "GetExternalIPAddress",
		ret_tag: "NewExternalIPAddress",
	}

	WAN_Packets_Received = UpnpValue{
		path:    "/igdupnp/control/WANCommonIFC1",
		service: "WANCommonInterfaceConfig:1",
		method:  "GetTotalPacketsReceived",
		ret_tag: "NewTotalPacketsReceived",
	}

	WAN_Packets_Sent = UpnpValue{
		path:    "/igdupnp/control/WANCommonIFC1",
		service: "WANCommonInterfaceConfig:1",
		method:  "GetTotalPacketsSent",
		ret_tag: "NewTotalPacketsSent",
	}

	WAN_Bytes_Received = UpnpValue{
		path:    "/igdupnp/control/WANCommonIFC1",
		service: "WANCommonInterfaceConfig:1",
		method:  "GetAddonInfos",
		ret_tag: "NewTotalBytesReceived",
	}

	WAN_Bytes_Sent = UpnpValue{
		path:    "/igdupnp/control/WANCommonIFC1",
		service: "WANCommonInterfaceConfig:1",
		method:  "GetAddonInfos",
		ret_tag: "NewTotalBytesSent",
	}
)

type Metric struct {
    UpnpValue
    *prometheus.Desc
}

func (m Metric) Value() (uint64, error) {
    strval, err := m.Query(*flag_dev_address)
    if err != nil {
        return 0, err
    }

    return strconv.ParseUint(strval, 10, 64)
}

func (m Metric) Describe(ch chan<- *prometheus.Desc) {
    ch <- m.Desc
}

func (m Metric) Collect(ch chan<- prometheus.Metric) error {
    val, err := m.Value()
    if err != nil {
        return err
    }

    ch <- prometheus.MustNewConstMetric(
        m.Desc,
        prometheus.CounterValue,
        float64(val),
    )
    return nil
}

var (
    packets_sent = Metric{
        WAN_Packets_Sent,
        prometheus.NewDesc(
            prometheus.BuildFQName("gateway", "wan", "packets_sent"),
            "packets sent on gateway wan interface",
            nil, 
            prometheus.Labels{"gateway": *flag_dev_address},
        ),
    }
    packets_received = Metric{
        WAN_Packets_Received,
        prometheus.NewDesc(
            prometheus.BuildFQName("gateway", "wan", "packets_received"),
            "packets received on gateway wan interface",
            nil, 
            prometheus.Labels{"gateway": *flag_dev_address},
        ),
    }
    bytes_sent = Metric{
        WAN_Bytes_Sent,
        prometheus.NewDesc(
            prometheus.BuildFQName("gateway", "wan", "bytes_sent"),
            "bytes sent on gateway wan interface",
            nil, 
            prometheus.Labels{"gateway": *flag_dev_address},
        ),
    }
    bytes_received = Metric{
        WAN_Bytes_Received,
        prometheus.NewDesc(
            prometheus.BuildFQName("gateway", "wan", "bytes_received"),
            "bytes received on gateway wan interface",
            nil, 
            prometheus.Labels{"gateway": *flag_dev_address},
        ),
    }
)


type FritzboxCollector struct {
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
    packets_sent.Describe(ch)
    packets_received.Describe(ch)
    bytes_sent.Describe(ch)
    bytes_received.Describe(ch)
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
    packets_sent.Collect(ch)
    packets_received.Collect(ch)
    bytes_sent.Collect(ch)
    bytes_received.Collect(ch)
}


func main() {
	flag.Parse()

	prometheus.MustRegister(&FritzboxCollector{})
	// Since we are dealing with custom Collector implementations, it might
	// be a good idea to enable the collect checks in the registry.
	prometheus.EnableCollectChecks(true)

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(*flag_addr, nil)
}
