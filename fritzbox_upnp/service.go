package fritzbox_upnp

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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// curl http://fritz.box:49000/igddesc.xml
// curl http://fritz.box:49000/any.xml
// curl http://fritz.box:49000/igdconnSCPD.xml
// curl http://fritz.box:49000/igdicfgSCPD.xml
// curl http://fritz.box:49000/igddslSCPD.xml
// curl http://fritz.box:49000/igd2ipv6fwcSCPD.xml

const text_xml = `text/xml; charset="utf-8"`

var ErrInvalidSOAPResponse = errors.New("invalid SOAP response")

type Root struct {
	BaseUrl  string
	Device   Device `xml:"device"`
	Services map[string]*Service
}

type Device struct {
	root *Root

	DeviceType       string `xml:"deviceType"`
	FriendlyName     string `xml:"friendlyName"`
	Manufacturer     string `xml:"manufacturer"`
	ManufacturerUrl  string `xml:"manufacturerURL"`
	ModelDescription string `xml:"modelDescription"`
	ModelName        string `xml:"modelName"`
	ModelNumber      string `xml:"modelNumber"`
	ModelUrl         string `xml:"modelURL"`
	UDN              string `xml:"UDN"`

	Services []*Service `xml:"serviceList>service"`
	Devices  []*Device  `xml:"deviceList>device"`

	PresentationUrl string `xml:"presentationURL"`
}

type Service struct {
	Device *Device

	ServiceType string `xml:"serviceType"`
	ServiceId   string `xml:"serviceId"`
	ControlUrl  string `xml:"controlURL"`
	EventSubUrl string `xml:"eventSubURL"`
	SCPDUrl     string `xml:"SCPDURL"`

	Actions        map[string]*Action
	StateVariables []*StateVariable
}

type scpdRoot struct {
	Actions        []*Action        `xml:"actionList>action"`
	StateVariables []*StateVariable `xml:"serviceStateTable>stateVariable"`
}

type Action struct {
	service *Service

	Name        string      `xml:"name"`
	Arguments   []*Argument `xml:"argumentList>argument"`
	ArgumentMap map[string]*Argument
}

func (a *Action) IsGetOnly() bool {
	for _, a := range a.Arguments {
		if a.Direction == "in" {
			return false
		}
	}
	return len(a.Arguments) > 0
}

type Argument struct {
	Name                 string `xml:"name"`
	Direction            string `xml:"direction"`
	RelatedStateVariable string `xml:"relatedStateVariable"`
	StateVariable        *StateVariable
}

type StateVariable struct {
	Name         string `xml:"name"`
	DataType     string `xml:"dataType"`
	DefaultValue string `xml:"defaultValue"`
}

type Result map[string]interface{}

func (r *Root) load() error {
	igddesc, err := http.Get(
		fmt.Sprintf("%s/igddesc.xml", r.BaseUrl),
	)

	if err != nil {
		return err
	}

	dec := xml.NewDecoder(igddesc.Body)

	err = dec.Decode(r)
	if err != nil {
		return err
	}

	r.Services = make(map[string]*Service)
	return r.Device.fillServices(r)
}

func (d *Device) fillServices(r *Root) error {
	d.root = r

	for _, s := range d.Services {
		s.Device = d

		response, err := http.Get(r.BaseUrl + s.SCPDUrl)
		if err != nil {
			return err
		}

		var scpd scpdRoot

		dec := xml.NewDecoder(response.Body)
		err = dec.Decode(&scpd)
		if err != nil {
			return err
		}

		s.Actions = make(map[string]*Action)
		for _, a := range scpd.Actions {
			s.Actions[a.Name] = a
		}
		s.StateVariables = scpd.StateVariables

		for _, a := range s.Actions {
			a.service = s
			a.ArgumentMap = make(map[string]*Argument)

			for _, arg := range a.Arguments {
				for _, svar := range s.StateVariables {
					if arg.RelatedStateVariable == svar.Name {
						arg.StateVariable = svar
					}
				}

				a.ArgumentMap[arg.Name] = arg
			}
		}

		r.Services[s.ServiceType] = s
	}
	for _, d2 := range d.Devices {
		err := d2.fillServices(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Action) Call() (Result, error) {
	bodystr := fmt.Sprintf(`
        <?xml version='1.0' encoding='utf-8'?> 
        <s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'> 
            <s:Body> 
                <u:%s xmlns:u='%s' /> 
            </s:Body>
        </s:Envelope>
    `, a.Name, a.service.ServiceType)

	url := a.service.Device.root.BaseUrl + a.service.ControlUrl
	body := strings.NewReader(bodystr)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	action := fmt.Sprintf("%s#%s", a.service.ServiceType, a.Name)

	req.Header["Content-Type"] = []string{text_xml}
	req.Header["SoapAction"] = []string{action}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := new(bytes.Buffer)
	data.ReadFrom(resp.Body)

	// fmt.Printf(data.String())
	return a.parseSoapResponse(data)

}

func (a *Action) parseSoapResponse(r io.Reader) (Result, error) {
	res := make(Result)
	dec := xml.NewDecoder(r)

	for {
		t, err := dec.Token()
		if err == io.EOF {
			return res, nil
		}

		if err != nil {
			return nil, err
		}

		if se, ok := t.(xml.StartElement); ok {
			arg, ok := a.ArgumentMap[se.Name.Local]

			if ok {
				t2, err := dec.Token()
				if err != nil {
					return nil, err
				}

				var val string
				switch element := t2.(type) {
				case xml.EndElement:
					val = ""
				case xml.CharData:
					val = string(element)
				default:
					return nil, ErrInvalidSOAPResponse
				}

				converted, err := convertResult(val, arg)
				if err != nil {
					return nil, err
				}
				res[arg.StateVariable.Name] = converted
			}
		}

	}
}

func convertResult(val string, arg *Argument) (interface{}, error) {
	switch arg.StateVariable.DataType {
	case "string":
		return val, nil
	case "boolean":
		return bool(val == "1"), nil

	case "ui1", "ui2", "ui4":
        // type ui4 can contain values greater than 2^32!
		res, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint64(res), nil
	default:
		return nil, fmt.Errorf("unknown datatype: %s", arg.StateVariable.DataType)

	}
}

func LoadServices(device string, port uint16) (*Root, error) {
	var root = &Root{
		BaseUrl: fmt.Sprintf("http://%s:%d", device, port),
	}

	err := root.load()
	if err != nil {
		return nil, err
	}

	return root, nil
}
