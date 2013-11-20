// Simple wrapper for Apache CloudStack API.

package cloudstack

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

type CreateSshKeyPairResponse struct {
	Createsshkeypairresponse struct {
		Keypair struct {
			Fingerprint string `json:"fingerprint"`
			Name        string `json:"name"`
			Privatekey  string `json:"privatekey"`
		} `json:"keypair"`
	} `json:"createsshkeypairresponse"`
}

type DeleteSshKeyPairResponse struct {
	Deletesshkeypairresponse struct {
		Success string `json:"success"`
	} `json:"deletesshkeypairresponse"`
}

type DeployVirtualMachineResponse struct {
	Deployvirtualmachineresponse struct {
		ID    string `json:"id"`
		Jobid string `json:"jobid"`
	} `json:"deployvirtualmachineresponse"`
}

type DestroyVirtualMachineResponse struct {
	Destroyvirtualmachineresponse struct {
		Jobid string `json:"jobid"`
	} `json:"destroyvirtualmachineresponse"`
}

type Template struct {
	Id           string
	Name         string
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

// Create a SSH key pair
func (c CloudStackClient) CreateSSHKeyPair(name string) (string, error) {
	params := url.Values{}
	params.Set("name", name)
	response, err := NewRequest(c, "createSSHKeyPair", params)
	if err != nil {
		return "", err
	}
	privatekey := response.(CreateSshKeyPairResponse).Createsshkeypairresponse.Keypair.Privatekey
	return privatekey, nil
}

// Deletes an SSH key
func (c CloudStackClient) DeleteSSHKeyPair(name string) (uint, error) {
	params := url.Values{}
	params.Set("name", name)
	_, err := NewRequest(c, "deleteSSHKeyPair", params)
	return 0, err
}

// Deploys a Virtual Machine and returns it's id
func (c CloudStackClient) DeployVirtualMachine(serviceofferingid string, templateid string, zoneid string, keypair string, displayname string, diskoffering string) (string, error) {
	params := url.Values{}
	params.Set("serviceofferingid", serviceofferingid)
	params.Set("templateid", templateid)
	params.Set("zoneid", zoneid)
	params.Set("keypair", keypair)
	params.Set("displayname", displayname)
	if diskoffering != "" {
		params.Set("diskoffering", diskoffering)
	}
	response, err := NewRequest(c, "deployVirtualMachine", params)
	if err != nil {
		return "", err
	}
	vmid := response.(DeployVirtualMachineResponse).Deployvirtualmachineresponse.ID
	return vmid, nil
}

// Destroys a Virtual Machine
func (c CloudStackClient) DestroyVirtualMachine(id string) (uint, error) {
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "destroyVirtualMachine", params)
	if err != nil {
		return "", err
	}
	jobid := response.(DestroyVirtualMachineResponse).Destroyvirtualmachineresponse.Jobid
	return jobid, nil
}

// Stops a Virtual Machine
func (c CloudStackClient) StopVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	_, err := NewRequest(c, "stopVirtualMachine", params)
	return "jobId", err
}

// Creates a Template of a Virtual Machine by it's ID
func (c CloudStackClient) CreateTemplate(displaytext string, name string, volumeid string, ostypeid string) (string, error) {
	params := url.Values{}
	params.Set("displaytext", displaytext)
	params.Set("name", name)
	params.Set("ostypeid", ostypeid)
	params.Set("volumeid", volumeid)
	_, err := NewRequest(c, "createTemplate", params)
	// return async job id
	return "jobId", err
}

// Returns all available templates
func (c CloudStackClient) Templates() ([]Template, error) {
	params := url.Values{}
	_, err := NewRequest(c, "listTemplates", params)
	// unmarshall json to a proper list
	return nil, err
}

// Deletes an template by its ID.
func (c CloudStackClient) DeleteTemplate(id string) (uint, error) {
	params := url.Values{}
	params.Set("id", id)
	_, err := NewRequest(c, "deleteTemplate", params)
	return 0, err
}

// Returns CloudStack string representation of status "off" "new" "active" etc.
func (c CloudStackClient) VirtualMachineState(id string) (string, string, error) {
	params := url.Values{}
	params.Set("id", id)
	_, err := NewRequest(c, "listVirtualMachines", params)
	// unpack state from json, return IP somehow as well
	return "1.2.3.4", "", err
}

// Query CloudStack for the state of a scheduled job
func (c CloudStackClient) QueryAsyncJobResult(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	_, err := NewRequest(c, "queryAsyncJobResult", params)
	return "state", err
}

func NewRequest(c CloudStackClient, request string, params url.Values) (interface{}, error) {
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

	log.Printf("response from cloudstack: %d - %s", resp.StatusCode, body)
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received HTTP client/server error from CloudStack: %d", resp.StatusCode))
		return nil,  err
	}

	switch request {
	default:
		log.Printf("Unknown request %s", request)
	case "createSSHKeyPair" :
		var decodedResponse CreateSshKeyPairResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "deleteSSHKeyPair":
		var decodedResponse DeleteSshKeyPairResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "deployVirtualMachine":
		var decodedResponse DeployVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "destroyVirtualMachine":
		var decodedResponse DestroyVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil
	}

	// only reached with unknown request
	return "", nil
}
