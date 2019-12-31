package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
)

const (
	TRUROUTE_REQUEST = 1 // this is a guess
	TRUROUTE_RESPONSE = 2
	TRUROUTE_RELEASE = 3
)

// TruRouteRequest XML struct for the request from a truroute services
//
// See: https://github.com/saulchelewani/truroute-ussd-adapter/
type TruRouteRequest struct {
	Ussd struct {
		Type int `xml:"type"`
		Message string `xml:"message"`
		Session string `xml:"session"`
		Msisdn string `xml:"msisdn`
	} `xml:"ussd"`
}

// TruRouteResponse XML struct for the response from a truroute services
// 
// See: https://github.com/saulchelewani/truroute-ussd-adapter/blob/master/src/UssdServiceProvider.php
type TruRouteResponse struct {
	Ussd struct {
		Type int `xml:"type"`
		Message string `xml:"msg"`
		Premium struct {
			Cost int `xml:"cost"`
			Ref string `xml:"ref"`
		} `xml:"premium"`
		Msisdn string `xml:"msisdn`
	} `xml:"ussd"`
}

func (t *TruRouteResponse) isResponse() bool {
	return t.Ussd.Type == TRUROUTE_RESPONSE
}

func (t *TruRouteResponse) isRelease() bool {
	return t.Ussd.Type == TRUROUTE_RELEASE
}

func (t *TruRouteResponse) GetText() string {
	return t.Ussd.Message
}

/// Executes a step and returns the result of the request
/// May return an empty string ("") upon failure
func (s *Step) executeAsTruRouteRequest(session *Session) (string, error) {
	var text = s.Text
	if &text == nil {
		return "", errors.New("Input Text cannot be nil")
	}

	req := &TruRouteRequest {}

	req.Ussd.Type = TRUROUTE_RESPONSE
	if s.isDial {
		req.Ussd.Type = TRUROUTE_REQUEST
	}
	req.Ussd.Message = text
	req.Ussd.Session = session.ID
	req.Ussd.Msisdn = session.PhoneNumber

	marshalledXml, err := xml.Marshal(req)
	res, err := session.client.Post(session.url, "text/xml", bytes.NewReader(marshalledXml))
	if err != nil {
		log.Printf("Failed to send request to %s", session.url)
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
