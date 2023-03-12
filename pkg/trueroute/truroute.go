package trueroute

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

type TrueRouteStep struct {
	*core.Step
}

const (
	TrurouteCodeInitial  = 1
	TrurouteCodeResponse = 2
	TrurouteCodeRelease  = 3
)

// TruRouteRequest XML struct for the request from a truroute services
//
// See: https://github.com/saulchelewani/truroute-ussd-adapter/
type TruRouteRequest struct {
	XMLName xml.Name `xml:"ussd"`
	Type    int      `xml:"type"`
	Message string   `xml:"msg"`
	Session string   `xml:"sessionid"`
	Msisdn  string   `xml:"msisdn"`
}

// TruRouteResponse XML struct for the response from a truroute services
//
// See: https://github.com/saulchelewani/truroute-ussd-adapter/blob/master/src/UssdServiceProvider.php
type TruRouteResponse struct {
	XMLName xml.Name `xml:"ussd"`
	Type    int      `xml:"type"`
	Message string   `xml:"msg"`
	Premium struct {
		Cost int    `xml:"cost"`
		Ref  string `xml:"ref"`
	} `xml:"premium"`
	Msisdn string `xml:"msisdn"`
}

// isResponse checks if the response is truly a success
func (t *TruRouteResponse) IsResponse() bool {
	return t.Type == TrurouteCodeResponse
}

func (t *TruRouteResponse) IsRelease() bool {
	return t.Type == TrurouteCodeRelease
}

// GetText returns the response text
func (t *TruRouteResponse) GetText() string {
	return t.Message
}

// ExecuteAsTruRouteRequest executes a step and returns the result of the request
// May return an empty string ("") upon failure
func (s *TrueRouteStep) ExecuteAsTruRouteRequest(session *core.Session) (string, error) {
	text := s.Text
	if text == "" {
		return "", errors.New("input Text cannot be nil")
	}

	req := &TruRouteRequest{}

	req.Type = TrurouteCodeResponse
	req.Message = text
	req.Session = session.ID
	req.Msisdn = session.PhoneNumber

	if s.IsDial {
		req.Type = TrurouteCodeInitial
		req.Message = "0"
	}

	marshalledXml, err := xml.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal XML request %v", req)
		return "", err
	}

	res, err := session.Client.Post(session.Url, "text/xml", bytes.NewReader(marshalledXml))
	if err != nil || res.StatusCode != 200 {
		log.Printf("Failed to send request %s to %s", marshalledXml, session.Url)
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	trurouteResponse := &TruRouteResponse{}

	if err = xml.Unmarshal(b, &trurouteResponse); err != nil {
		return "", err
	}

	s.IsLast = trurouteResponse.IsRelease()

	return trurouteResponse.GetText(), nil
}
