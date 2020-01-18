package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
)

const (
	TRUROUTE_INITIAL  = 1
	TRUROUTE_RESPONSE = 2
	TRUROUTE_RELEASE  = 3
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

func (t *TruRouteResponse) isResponse() bool {
	return t.Type == TRUROUTE_RESPONSE
}

func (t *TruRouteResponse) isRelease() bool {
	return t.Type == TRUROUTE_RELEASE
}

func (t *TruRouteResponse) GetText() string {
	return t.Message
}

/// Executes a step and returns the result of the request
/// May return an empty string ("") upon failure
func (s *Step) ExecuteAsTruRouteRequest(session *Session) (string, error) {
	var text = s.Text
	if &text == nil {
		return "", errors.New("Input Text cannot be nil")
	}

	req := &TruRouteRequest{}

	req.Type = TRUROUTE_RESPONSE
	req.Message = text
	req.Session = session.ID
	req.Msisdn = session.PhoneNumber

	if s.isDial {
		req.Type = TRUROUTE_INITIAL
		req.Message = "0"
	}

	marshalledXml, err := xml.Marshal(req)

	res, err := session.client.Post(session.url, "text/xml", bytes.NewReader(marshalledXml))
	if err != nil || res.StatusCode != 200 {
		log.Printf("Failed to send request %s to %s", marshalledXml, session.url)
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

	s.isLast = trurouteResponse.isRelease()

	return trurouteResponse.GetText(), nil
}
