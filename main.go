package main

// Copyright 2016 Nils Decker
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
)

var (
	flag_test = flag.Bool("test", false, "print all available metrics to stdout")
	flag_addr = flag.String("listen-address", ":9133", "The address to listen on for HTTP requests.")

	flag_gateway_address = flag.String("gateway-address", "fritz.box", "The URL of the upnp service")
	flag_gateway_port    = flag.Int("gateway-port", 49000, "The URL of the upnp service")
)

var (
	collect_errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fritzbox_exporter_collect_errors",
		Help: "Number of collection errors.",
	})
)

type Metric struct {
	Service string
	Action  string
	Result  string
	OkValue string

	Desc       *prometheus.Desc
	MetricType prometheus.ValueType
}

var metrics = []*Metric{
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetTotalPacketsReceived",
		Result:  "TotalPacketsReceived",
		Desc: prometheus.NewDesc(
			"gateway_wan_packets_received",
			"packets received on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetTotalPacketsSent",
		Result:  "TotalPacketsSent",
		Desc: prometheus.NewDesc(
			"gateway_wan_packets_sent",
			"packets sent on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "TotalBytesReceived",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_received",
			"bytes received on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "TotalBytesSent",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_sent",
			"bytes sent on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "Layer1UpstreamMaxBitRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_upstream_max_bitrate",
			"Layer1 upstream max bitrate",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "Layer1DownstreamMaxBitRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_downstream_max_bitrate",
			"Layer1 downstream max bitrate",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "PhysicalLinkStatus",
		OkValue: "Up",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_link_status",
			"Status of physical link (Up = 1)",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANIPConnection:1",
		Action:  "GetStatusInfo",
		Result:  "ConnectionStatus",
		OkValue: "Connected",
		Desc: prometheus.NewDesc(
			"gateway_wan_connection_status",
			"WAN connection status (Connected = 1)",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANIPConnection:1",
		Action:  "GetStatusInfo",
		Result:  "Uptime",
		Desc: prometheus.NewDesc(
			"gateway_wan_connection_uptime_seconds",
			"WAN connection uptime",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
}

type FritzboxCollector struct {
	Root    *upnp.Root
	Gateway string
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range metrics {
		ch <- m.Desc
	}
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
	var err error
	var last_service string
	var last_method string
	var last_result upnp.Result

	for _, m := range metrics {
		if m.Service != last_service || m.Action != last_method {
			service, ok := fc.Root.Services[m.Service]
			if !ok {
				// TODO
				fmt.Println("cannot find service", m.Service)
				fmt.Println(fc.Root.Services)
				continue
			}
			action, ok := service.Actions[m.Action]
			if !ok {
				// TODO
				fmt.Println("cannot find action", m.Action)
				continue
			}

			last_result, err = action.Call()
			if err != nil {
				fmt.Println(err)
				collect_errors.Inc()
				continue
			}
		}

		val, ok := last_result[m.Result]
		if !ok {
			fmt.Println("result not found", m.Result)
			collect_errors.Inc()
			continue
		}

		var floatval float64
		switch tval := val.(type) {
		case uint64:
			floatval = float64(tval)
		case bool:
			if tval {
				floatval = 1
			} else {
				floatval = 0
			}
		case string:
			if tval == m.OkValue {
				floatval = 1
			} else {
				floatval = 0
			}
		default:
			fmt.Println("unknown", val)
			collect_errors.Inc()
			continue

		}

		ch <- prometheus.MustNewConstMetric(
			m.Desc,
			m.MetricType,
			floatval,
			fc.Gateway,
		)
	}
}

func main() {
	flag.Parse()

	root, err := upnp.LoadServices(*flag_gateway_address, uint16(*flag_gateway_port))
	if err != nil {
		panic(err)
	}

	if *flag_test {
		for _, s := range root.Services {
			fmt.Printf("%s: %s\n", s.Device.FriendlyName, s.ServiceType)
			for _, a := range s.Actions {
				if !a.IsGetOnly() {
					continue
				}

				res, err := a.Call()
				if err != nil {
					panic(err)
				}

				fmt.Printf("  %s\n", a.Name)
				for _, arg := range a.Arguments {
					fmt.Printf("    %s: %v\n", arg.RelatedStateVariable, res[arg.StateVariable.Name])
				}
			}
		}

		return
	}

	prometheus.MustRegister(&FritzboxCollector{root, *flag_gateway_address})
	prometheus.MustRegister(collect_errors)

	// Since we are dealing with custom Collector implementations, it might
	// be a good idea to enable the collect checks in the registry.
	prometheus.EnableCollectChecks(true)

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(*flag_addr, nil)
}
