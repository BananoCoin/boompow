package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bananocoin/boompow/libs/models"
	"github.com/golang/glog"
)

type RPCClient struct {
	Url string
}

type SendResponse struct {
	Block string `json:"block"`
}

// Base request
func (client RPCClient) makeRequest(request interface{}) ([]byte, error) {
	requestBody, _ := json.Marshal(request)
	// HTTP post
	resp, err := http.Post(client.Url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		glog.Errorf("Error making RPC request %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	// Try to decode+deserialize
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Error decoding response body %s", err)
		return nil, err
	}
	return body, nil
}

// send
func (client RPCClient) MakeSendRequest(request models.SendRequest) (*SendResponse, error) {
	response, err := client.makeRequest(request)
	if err != nil {
		glog.Errorf("Error making request %s", err)
		return nil, err
	}
	// Try to decode+deserialize
	var sendResponse SendResponse
	err = json.Unmarshal(response, &sendResponse)
	if err != nil {
		glog.Errorf("Error unmarshaling response %s, %s", string(response), err)
		return nil, errors.New("Error")
	}
	return &sendResponse, nil
}
