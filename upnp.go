package main

import (
    "bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const TEXT_XML = `text/xml; charset="utf-8"`

var (
    ErrResultNotFound = errors.New("result not found")
    ErrResultWithoutChardata = errors.New("result without chardata")
)

// curl "http://fritz.box:49000/igdupnp/control/WANIPConn1"
//   -H "Content-Type: text/xml; charset="utf-8""
//   -H "SoapAction:urn:schemas-upnp-org:service:WANIPConnection:1#GetExternalIPAddress"
//   -d "<?xml version='1.0' encoding='utf-8'?>
//      <s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'>
//      <s:Body>
//          <u:GetExternalIPAddress xmlns:u='urn:schemas-upnp-org:service:WANIPConnection:1' />
//      </s:Body> </s:Envelope>"

type UpnpValue struct {
	path    string
	service string
	method  string
	ret_tag string
}

func (v *UpnpValue) Query(device string) (string, error) {
	url := fmt.Sprintf("http://%s:49000%s", device, v.path)

	bodystr := fmt.Sprintf(`
        <?xml version='1.0' encoding='utf-8'?> 
        <s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'> 
            <s:Body> 
                <u:%s xmlns:u='urn:schemas-upnp-org:service:%s' /> 
            </s:Body>
        </s:Envelope>
    `, v.method, v.service)

	body := strings.NewReader(bodystr)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	action := fmt.Sprintf("urn:schemas-upnp-org:service:%s#%s", v.service, v.method)

	req.Header["Content-Type"] = []string{TEXT_XML}
	req.Header["SoapAction"] = []string{action}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

    data := new(bytes.Buffer)
    data.ReadFrom(resp.Body)

    // fmt.Printf(data.String())

	dec := xml.NewDecoder(data)

	for {
		t, err := dec.Token()
		if err == io.EOF {
			return "", ErrResultNotFound
		}

		if err != nil {
			return "", err
		}

		if se, ok := t.(xml.StartElement); ok {
			if se.Name.Local == v.ret_tag {
				t2, err := dec.Token()
				if err != nil {
					return "", err
				}

				data, ok := t2.(xml.CharData)
				if !ok {
					return "", ErrResultWithoutChardata
				}
				return string(data), nil
			}
		}

	}
}
