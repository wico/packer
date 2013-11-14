// Simple wrapper for Apache CloudStack API.

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Template struct {
	Id           uint
	Name         string
	Distribution string
}

type TemplatesResp struct {
	Templates []Template
}

type CloudStackClient struct {
	// The http client for communicating
	client *http.Client

	// The base URL of the API
	BaseURL string

	// Credentials
	APIKey string
	Secret string
}

// Creates a new client for communicating with CloudStack
func (cloudstack CloudStackClient) New(apiurl string, apikey string, secret string) *CloudStackClient {
	c := &CloudStackClient{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		BaseURL: apiurl,
		APIKey:  apikey,
		Secret:  secret,
	}
	return c
}

// Create a SSH key
func (c CloudStackClient) CreateKey(name string, pub string) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "createSSHKeyPair", params)
	return 0, err
}

// Deletes an SSH key
func (c CloudStackClient) DeletesKey(id uint) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "deleteSSHKeyPair", params)
	return 0, err
}

// Deploys a Virtual Machine and returns it's id
func (c CloudStackClient) DeployVirtualMachine(serviceofferingid string, templateid string, zoneid string) (uint, error) {
	params := url.Values{}
	params.Set("serviceofferingid", serviceofferingid)
	params.Set("templateid", templateid)
	params.Set("zoneid", zoneid)

	_, err :=NewRequest(c, "deployVirtualMachine", params)
	return 0, err
}

// Destroys a Virtual Machine
func (c CloudStackClient) DestroyVirtualMachine(id string) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "destroyVirtualMachine", params)
	return 0, err
}

// Powers off a Virtual Machine
func (c CloudStackClient) StopVirtualMachine(id uint) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "stopVirtualMachine", params)
	return 0, err
}

// Shutdown a Virtual Machine
func (c CloudStackClient) ShutdownVM(id uint) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "stopVirtualMachine", params)
	return 0, err
}

// Creates a snaphot of a Virtual Machine by it's ID
func (c CloudStackClient) CreateSnapshot(id uint, name string) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "createSnapshot", params)
	return 0, err
}

// Returns all available templates
func (c CloudStackClient) Templates() ([]Template, error) {
	params := url.Values{}
	_, err := NewRequest(c, "listTemplates", params)
	return nil, err
}

// Deletes an template by its ID.
func (c CloudStackClient) DeleteTemplate(id uint) (uint, error) {
	params := url.Values{}
	_, err := NewRequest(c, "deleteTemplate", params)
	return 0, err
}

// Returns CloudStack string representation of status "off" "new" "active" etc.
func (c CloudStackClient) VMStatus(id uint) (string, string, error) {
	params := url.Values{}
	_, err := NewRequest(c, "poweroff", params)
	return "", "", err
}

func NewRequest(c CloudStackClient, request string, params url.Values) (map[string]interface{}, error) {
	client := c.client

	params.Set("apikey", c.APIKey)
	params.Set("command", request)
	params.Set("response", "json")

	// Generate signature for API call
	// * Serialize parameters and sort them by key, done by Encode
	// * Convert the entire argument string to lowercase
	// * Calculate HMAC SHA1 of argument string with CloudStack secret
	// * URL encode the string and convert to base64
	s := params.Encode()
	s2 := strings.ToLower(s)
	mac := hmac.New(sha1.New, []byte(c.Secret))
	mac.Write([]byte(s2))
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	signature = url.QueryEscape(signature)
	// Apparently we need to manually(?) escape the underscore
	signature = strings.Replace(signature, "_", "%2F", -1)

	// Create the final URL before we issue the request
	url := c.BaseURL + "?" + s + "&signature=" + signature

	log.Printf("Calling %s ", url)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	log.Printf("response from cloudstack: %s", body)

	var decodedResponse map[string]interface{}
	err = json.Unmarshal(body, &decodedResponse)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to decode JSON response (HTTP %v) from CloudStack: %s",
			resp.StatusCode, body))
		return decodedResponse, err
	}

	return decodedResponse, nil
}
