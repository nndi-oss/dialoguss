package africastalking

import (
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/nndi-oss/dialoguss/pkg/core"
)

type AfricasTalkingRouteStep struct {
	*core.Step
}

var (
	reUssdCon = regexp.MustCompile(`^CON\s?`)
	reUssdEnd = regexp.MustCompile(`^END\s?`)
)

/// Executes a step as an AfricasTalking API request
/// May return an empty string ("") upon failure
func (s *AfricasTalkingRouteStep) ExecuteAsAfricasTalking(session *core.Session) (string, error) {
	data := url.Values{}
	data.Set("sessionId", session.ID)
	data.Set("phoneNumber", session.PhoneNumber)
	data.Set("serviceCode", session.ServiceCode)
	var text = s.Text
	if &text == nil {
		return "", errors.New("Input Text cannot be nil")
	}
	data.Set("text", text)  // TODO(zikani): concat the input
	data.Set("channel", "") // TODO: Get the channel

	res, err := session.Client.PostForm(session.Url, data)
	if err != nil {
		log.Printf("Failed to send request to %s", session.Url)
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
		s.IsLast = false
	} else if reUssdEnd.MatchString(responseText) {
		responseText = strings.Replace(responseText, "END ", "", 1)
		s.IsLast = true
	}

	return responseText, nil
}
