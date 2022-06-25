package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
)

var (
	reUssdCon = regexp.MustCompile(`^CON\s?`)
	reUssdEnd = regexp.MustCompile(`^END\s?`)
)

/// Executes a step as an AfricasTalking API request
/// May return an empty string ("") upon failure
func (s *Step) ExecuteAsAfricasTalking(session *Session) (string, error) {
	data := url.Values{}
	data.Set("sessionId", session.ID)
	data.Set("phoneNumber", session.PhoneNumber)
	data.Set("serviceCode", session.serviceCode)
	var text = s.Text
	if &text == nil {
		return "", errors.New("Input Text cannot be nil")
	}
	data.Set("text", text)  // TODO(zikani): concat the input
	data.Set("channel", "") // TODO: Get the channel

	res, err := session.client.PostForm(session.url, data)
	if err != nil {
		log.Printf("Failed to send request to %s", session.url)
		return "", err
	}

	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	responseText := string(b)
	if reUssdCon.MatchString(responseText) {
		responseText = strings.Replace(responseText, "CON ", "", 1)
		s.isLast = false
	} else if reUssdEnd.MatchString(responseText) {
		responseText = strings.Replace(responseText, "END ", "", 1)
		s.isLast = true
	}

	return responseText, nil
}
