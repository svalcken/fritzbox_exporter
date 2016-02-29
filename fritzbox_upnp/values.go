package fritzbox_upnp

// curl http://fritz.box:49000/igddesc.xml
// curl http://fritz.box:49000/any.xml
// curl http://fritz.box:49000/igdconnSCPD.xml
// curl http://fritz.box:49000/igdicfgSCPD.xml
// curl http://fritz.box:49000/igddslSCPD.xml
// curl http://fritz.box:49000/igd2ipv6fwcSCPD.xml

var (
	WAN_IP = UpnpValueString{UpnpValue{
		Path:    "/igdupnp/control/WANIPConn1",
		Service: "WANIPConnection:1",
		Method:  "GetExternalIPAddress",
		RetTag:  "NewExternalIPAddress",

		ShortName: "wan_ip",
		Help:      "WAN IP Adress",
	}}

	WAN_Packets_Received = UpnpValueUint{UpnpValue{
		Path:    "/igdupnp/control/WANCommonIFC1",
		Service: "WANCommonInterfaceConfig:1",
		Method:  "GetTotalPacketsReceived",
		RetTag:  "NewTotalPacketsReceived",

		ShortName: "packets_received",
		Help:      "packets received on gateway WAN interface",
	}}

	WAN_Packets_Sent = UpnpValueUint{UpnpValue{
		Path:    "/igdupnp/control/WANCommonIFC1",
		Service: "WANCommonInterfaceConfig:1",
		Method:  "GetTotalPacketsSent",
		RetTag:  "NewTotalPacketsSent",

		ShortName: "packets_sent",
		Help:      "packets sent on gateway WAN interface",
	}}

	WAN_Bytes_Received = UpnpValueUint{UpnpValue{
		Path:    "/igdupnp/control/WANCommonIFC1",
		Service: "WANCommonInterfaceConfig:1",
		Method:  "GetAddonInfos",
		RetTag:  "NewTotalBytesReceived",

		ShortName: "bytes_received",
		Help:      "bytes received on gateway WAN interface",
	}}

	WAN_Bytes_Sent = UpnpValueUint{UpnpValue{
		Path:    "/igdupnp/control/WANCommonIFC1",
		Service: "WANCommonInterfaceConfig:1",
		Method:  "GetAddonInfos",
		RetTag:  "NewTotalBytesSent",

		ShortName: "bytes_sent",
		Help:      "bytes sent on gateway WAN interface",
	}}
)

var Values = []UpnpValueUint{
	WAN_Packets_Received,
	WAN_Packets_Sent,
	WAN_Bytes_Received,
	WAN_Bytes_Sent,
}
