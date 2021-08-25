package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	tpState "healthcare-system-sawtooth/tp/state"
)

// GetStateData returns the data of the address in byte slice.
func GetStateData(addr string) ([]byte, error) {
	apiSuffix := fmt.Sprintf("%s/%s", StateAPI, addr)
	resp, err := sendRequestByAPISuffix(apiSuffix, nil, "")
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(resp["data"].(string))
}

// List returns the list of data that address started with the address prefix.
func list(address, start string, limit uint) (result []interface{}, err error) {
	apiSuffix := fmt.Sprintf("%s?address=%s", StateAPI, address)
	if start != "" {
		apiSuffix = fmt.Sprintf("%s&start=%s", apiSuffix, start)
	}
	if limit > 0 {
		apiSuffix = fmt.Sprintf("%s&limit=%v", apiSuffix, limit)
	}
	response, err := sendRequestByAPISuffix(apiSuffix, nil, "")
	if err != nil {
		return
	}
	return response["data"].([]interface{}), nil
}

// ListUsers returns the list of data that address started with the UserNamespace.
func ListUsers(start string, limit uint) ([]interface{}, error) {
	return list(tpState.Namespace+tpState.UserNamespace, start, limit)
}

// sendRequest send the request to the Hyperledger Sawtooth rest api by giving url.
func sendRequest(url string, data []byte, contentType string) (map[string]interface{}, error) {
	// SendUploadQuery request to validator rest api
	var response *http.Response
	var err error
	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes.NewBuffer(data))
	} else {
		response, err = http.Get(url)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to REST API: %v", err)
	}
	if response.StatusCode == 404 {
		return nil, fmt.Errorf("no such endpoint: %s", url)
	} else if response.StatusCode >= 400 {
		return nil, fmt.Errorf("error %d: %s", response.StatusCode, response.Status)
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(responseBody, &responseMap)
	if err != nil {
		return nil, err
	}
	return responseMap, nil
}

// sendRequest send the request to the Hyperledger Sawtooth rest api by giving api suffix.
func sendRequestByAPISuffix(apiSuffix string, data []byte, contentType string) (map[string]interface{}, error) {
	var url string
	// Construct url
	if strings.HasPrefix(TPURL, "http://") {
		url = fmt.Sprintf("%s/%s", TPURL, apiSuffix)
	} else {
		url = fmt.Sprintf("http://%s/%s", TPURL, apiSuffix)
	}

	return sendRequest(url, data, contentType)
}
