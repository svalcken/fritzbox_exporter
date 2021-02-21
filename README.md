# Fritz!Box Upnp statistics exporter for prometheus

This exporter exports some variables from an
[AVM Fritzbox](http://avm.de/produkte/fritzbox/) to prometheus.

This exporter is tested with a Fritzbox 7590 software version 07.12 and
07.20.

The goal of the fork is:
- [x] allow passing of username / password using environment variable
- [x] use https instead of http for communication with fritz.box
- [x] move config of metrics to be exported to config file rather then
      code
- [x] add a configuration for additional metrics to collect (especially
      from TR-064 API)
- [x] create a grafana dashboard consuming the additional metrics
- [x] add a docker build
- [x] exposes health check endpoints

Other changes:
- replaced digest authentication code with own implementation
- improved error messages
- **New:** test mode prints details about all SOAP Actions and their
  parameters
- **New:** collect option to directly test collection of results
- **New:** additional metrics to collect details about connected hosts
  and DECT devices
- **New:** support to use results like hostname or MAC address as labels
  to metrics

[TOC]: # "## Table of Contents"

## Table of Contents
- [Building](#building)
- [Running](#running)
  - [Running with docker](#running-with-docker)
- [Exported metrics](#exported-metrics)
- [Output of `-test`](#output-of--test)
- [Customizing metrics](#customizing-metrics)
- [Grafana Dashboard](#grafana-dashboard)

## Building

```shell script
git clone https://gitlab.com/dekarl/fritzbox_exporter.git
cd fritzbox_exporter
go mod download
go build
```

Alternatively there is a [`Dockerfile`](Dockerfile) to build a docker
image.

```shell script
docker build -t fritzbox-exporter .
```

## Running

In the configuration of the Fritzbox the option `Statusinformationen
über UPnP übertragen` in the dialog `Heimnetz > Heimnetzübersicht >
Netzwerkeinstellungen` has to be enabled.

Usage:

```
$GOPATH/bin/fritzbox_exporter -h
Usage of /fritzbox-exporter/fritzbox-exporter:
  -collect=false: 
    print configured metrics to stdout and exit
  -gateway-url="http://fritz.box:49000": 
    The URL of the FRITZ!Box
  -json-out="": 
    store metrics also to JSON file when running test
  -listen-address="127.0.0.1:9042": 
    The address to listen on for HTTP requests.
  -metrics-file="metrics.json": 
    The JSON file with the metric definitions.
  -password="": 
    The password for the FRITZ!Box UPnP service
  -test=false: 
    print all available metrics to stdout
  -username="": 
    The user for the FRITZ!Box UPnP service
  -verifyTls=false: 
    Verify the tls connection when connecting to the FRITZ!Box
```

The password can be passed over environment variables to test in shell:

```shell script
read -rs PASSWORD && export PASSWORD && ./fritzbox_exporter -username <user> -test; unset PASSWORD
```

### Running with docker

The fritzbox-exporter will be built by the Gitlab Infrastructure
which can be used with:

```shell script
docker run -p 8080:8080 registry.gitlab.com/dekarl/fritzbox_exporter
```

It supports all commandline arguments like the original one:

```shell script
docker run registry.gitlab.com/dekarl/fritzbox_exporter -h
```

See also <https://gitlab.com/dekarl/fritzbox_exporter/container_registry>.

## Exported metrics

Start the exporter and run:

```shell script
curl -s http://127.0.0.1:9042/metrics 
```

## Output of `-test`

The exporter prints all available Variables to `stdout` when called with
the `-test` option. It retrieves these values by parsing all services
from <http://fritz.box:49000/igddesc.xml> and
<http://fritzbox:49000/tr64desc.xml>. To access TR64 the exporter needs
username and password.

## Customizing metrics

The metrics to collect are no longer hard coded, but have been moved to
the [`metrics.json`](metrics.json) file, so just adjust to your needs.
For a list of all the available metrics just execute the exporter with
`-test` (username and password are needed for the TR-064 API!)

For a list of all available metrics, see the dumps below (the format is
the same as in the metrics.json file, so it can be used to easily add
further metrics to retrieve):
- [FritzBox 7590 v7.12](all_available_metrics_7590_7.12.json)
- [FritzBox 7590 v7.20](all_available_metrics_7590_7.20.json)

## Grafana Dashboard

The dashboard is now also published on
[Grafana](https://grafana.com/grafana/dashboards/12579).
