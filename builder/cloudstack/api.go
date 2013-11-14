// Simple wrapper for Apache CloudStack API.

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	NewRequest(c, "createSSHKeyPair", params)
	return 0, nil
}

// Deletes an SSH key
func (c CloudStackClient) DeletesKey(id uint) error {
	params := url.Values{}
	NewRequest(c, "deleteSSHKeyPair", params)
	return nil
}

// Deploys a Virtual Machine and returns it's id
func (c CloudStackClient) DeployVirtualMachine(serviceofferingid string, templateid string, zoneid string) (uint, error) {
	params := url.Values{}
	params.Set("serviceofferingid", serviceofferingid)
	params.Set("templateid", templateid)
	params.Set("zoneid", zoneid)

	NewRequest(c, "deployVirtualMachine", params)
	return 0, nil
}

// Destroys a Virtual Machine
func (c CloudStackClient) DestroyVM(id uint) error {
	params := url.Values{}
	NewRequest(c, "destroyVirtualMachine", params)
	return nil
}

// Powers off a Virtual Machine
func (c CloudStackClient) StopVirtualMachine(id uint) error {
	params := url.Values{}
	NewRequest(c, "stopVirtualMachine", params)
	return nil
}

// Shutdown a Virtual Machine
func (c CloudStackClient) ShutdownVM(id uint) error {
	params := url.Values{}
	NewRequest(c, "stopVirtualMachine", params)
	return nil
}

// Creates a snaphot of a Virtual Machine by it's ID
func (c CloudStackClient) CreateSnapshot(id uint, name string) error {
	params := url.Values{}
	NewRequest(c, "createSnapshot", params)
	return nil
}

// Returns all available templates
func (c CloudStackClient) Templates() ([]Template, error) {
	params := url.Values{}
	NewRequest(c, "listTemplates", params)
	return nil, nil
}

// Deletes an template by its ID.
func (c CloudStackClient) DeleteTemplate(id uint) error {
	params := url.Values{}
	NewRequest(c, "deleteTemplate", params)
	return nil
}

// Returns CloudStack string representation of status "off" "new" "active" etc.
func (c CloudStackClient) VMStatus(id uint) (string, string, error) {
	params := url.Values{}
	NewRequest(c, "poweroff", params)
	return "", "", nil
}

func NewRequest(c CloudStackClient, request string, params url.Values) {
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
	api_call := c.BaseURL + "?" + s + "&signature=" + signature

	fmt.Println("Calling: " + api_call)

	// Print the results if we recieve a 200 response.
	resp, err := c.client.Get(api_call)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", string(contents))
	}
}
