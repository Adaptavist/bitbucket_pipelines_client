package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	retry "github.com/hashicorp/go-retryablehttp"
)

func (h *Client) getHttpClient() *retry.Client {
	if h.client == nil {
		h.client = retry.NewClient()
	}
	return h.client
}

func (h *Client) doRequest(r *retry.Request) (*http.Response, error) {
	r.SetBasicAuth(h.Config.Username, h.Config.Password)
	r.Header.Add("Content-Type", "application/json")
	return h.getHttpClient().Do(r)
}

func (h Client) get(url string) ([]byte, error) {
	req, requestErr := retry.NewRequest("GET", url, nil)

	if requestErr != nil {
		return nil, requestErr
	}

	res, httpErr := h.doRequest(req)

	if httpErr != nil {
		return nil, httpErr
	}

	resp, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return nil, readErr
	}

	errStatus := hasError(res.StatusCode)

	if errStatus != nil {
		return nil, fmt.Errorf("failed to GET (%s) - %s - %s", url, errStatus, resp)
	}

	return resp, nil
}

// hasError returns an error if 40x or 50x codes are given
func hasError(s int) (err error) {
	if s >= 400 && s < 600 {
		err = fmt.Errorf("received %s", strconv.Itoa(s))
	}
	return
}

// post an BasicAuth authenticated resource
func (h Client) post(url string, data interface{}) ([]byte, error) {
	reqData, dataErr := json.Marshal(data)

	if dataErr != nil {
		return nil, dataErr
	}

	reqBody := bytes.NewBuffer(reqData)
	req, reqErr := retry.NewRequest("POST", url, reqBody)

	if reqErr != nil {
		return nil, reqErr
	}

	res, httpErr := h.doRequest(req)

	if httpErr != nil {
		return nil, httpErr
	}

	statusErr := hasError(res.StatusCode)

	resp, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return nil, readErr
	}

	if statusErr != nil {
		return nil, fmt.Errorf("failed to POST (%s) %s - %s", url, statusErr, string(resp))
	}

	return resp, nil
}

// postUnmarshalled makes a POST HTTP request and unmarshalls the data
func (h Client) postUnmarshalled(url string, data interface{}, target interface{}) (err error) {
	resp, err := h.post(url, data)

	if err != nil {
		return
	}

	err = json.Unmarshal(resp, target)

	return
}

// getUnmarshalled makes a GET HTTP request and unmarshalls the data
func (h Client) getUnmarshalled(url string, targetPtr interface{}) (err error) {
	resp, err := h.get(url)

	if err != nil {
		err = fmt.Errorf("%s - %s", err.Error(), string(resp))
		return
	}

	err = json.Unmarshal(resp, targetPtr)

	return
}
