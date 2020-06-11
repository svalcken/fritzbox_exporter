# Fritz!Box Upnp statistics exporter for prometheus

This exporter exports some variables from an 
[AVM Fritzbox](http://avm.de/produkte/fritzbox/)
to prometheus.

This exporter is tested with a Fritzbox 7590 software version 07.12.

This is a fork from:
https://github.com/123Haynes/fritzbox_exporter
which is forked from:
https://github.com/ndecker/fritzbox_exporter

The goal of the fork is:
  - allow passing of username / password using evironment variable - done
  - use https instead of http for communitcation with fritz.box - done
  - move config of metrics to be exported to config file rather then code
  - add config for additional metrics to collect (especially from TR-064 API)
  - create a grafana dashboard consing the additional metrics


## Building

    go get github.com/sberk42/fritzbox_exporter/
    cd $GOPATH/src/github.com/sberk42/fritzbox_exporter
    go install

## Running

In the configuration of the Fritzbox the option "Statusinformationen über UPnP übertragen" in the dialog "Heimnetz >
Heimnetzübersicht > Netzwerkeinstellungen" has to be enabled.

Usage:

    $GOPATH/bin/fritzbox_exporter -h
    Usage of ./fritzbox_exporter:
      -gateway-url string
            The URL of the FRITZ!Box (default "https://fritz.box:49443")
      -listen-address string
            The address to listen on for HTTP requests. (default "127.0.0.1:9042")
      -password string
            The password for the FRITZ!Box UPnP service
      -test
            print all available metrics to stdout
      -username string
            The user for the FRITZ!Box UPnP service

    The password (needed for metrics from TR-064 API) can be passed over environment variables to test in shell:
    read -rs PASSWORD && ./fritzbox_exporter -username <user> -test; unset PASSWORD

## Exported metrics

start exporter and run
curl -s http://127.0.0.1:9042/metrics 

## Output of -test

The exporter prints all available Variables to stdout when called with the -test option.
These values are determined by parsing all services from http://fritz.box:49000/igddesc.xml and http://fritzbox:49000/tr64desc.xml (for TR64 username and password is needed!!!)

