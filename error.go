package letsencrypt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// An HTTP error conforming to the following spec
// https://tools.ietf.org/html/draft-ietf-appsawg-http-problem-01
type Error struct {
	Typ    string `json:"type"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("acme error '%s': %s", err.Typ, err.Detail)
}

func checkHTTPError(resp *http.Response, expCode int) error {
	if resp.StatusCode == expCode {
		return nil
	}

	// errors are only defined for status codes 4XX and 5XXX
	// https://letsencrypt.github.io/acme-spec/#errors
	if !(resp.StatusCode >= 400 && resp.StatusCode < 600) {
		return fmt.Errorf("acme: expected Status %d %s, got %s", expCode, http.StatusText(expCode), resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %v", err)
	}
	var errData struct {
		Typ    string `json:"type"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &errData); err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}
	return &Error{
		Typ:    strings.TrimPrefix(errData.Typ, "urn:acme:error:"),
		Detail: errData.Detail,
		Status: resp.StatusCode,
	}
}
