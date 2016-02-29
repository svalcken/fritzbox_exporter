package fritzbox_upnp

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const text_xml = `text/xml; charset="utf-8"`

var (
	ErrResultNotFound        = errors.New("result not found")
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
	Path    string
	Service string
	Method  string
	RetTag  string

	ShortName string
	Help      string
}

func (v *UpnpValue) query(device string, port uint16) (string, error) {
	url := fmt.Sprintf("http://%s:%d%s", device, port, v.Path)

	bodystr := fmt.Sprintf(`
        <?xml version='1.0' encoding='utf-8'?> 
        <s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'> 
            <s:Body> 
                <u:%s xmlns:u='urn:schemas-upnp-org:service:%s' /> 
            </s:Body>
        </s:Envelope>
    `, v.Method, v.Service)

	body := strings.NewReader(bodystr)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	action := fmt.Sprintf("urn:schemas-upnp-org:service:%s#%s", v.Service, v.Method)

	req.Header["Content-Type"] = []string{text_xml}
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
			if se.Name.Local == v.RetTag {
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

type UpnpValueString struct{ UpnpValue }

func (v *UpnpValueString) Query(device string, port uint16) (string, error) {
	return v.query(device, port)
}

type UpnpValueUint struct{ UpnpValue }

func (v *UpnpValueUint) Query(device string, port uint16) (uint64, error) {
	strval, err := v.query(device, port)
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseUint(strval, 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
