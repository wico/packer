// Simple wrapper for Apache CloudStack API.

package cloudstack

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

type StopVirtualMachineResponse struct {
	Stopvirtualmachineresponse struct {
		Jobid string `json:"jobid"`
	} `json:"stopvirtualmachineresponse"`
}

type Nic struct {
	Gateway     string `json:"gateway"`
	ID          string `json:"id"`
	Ipaddress   string `json:"ipaddress"`
	Isdefault   bool   `json:"isdefault"`
	Macaddress  string `json:"macaddress"`
	Netmask     string `json:"netmask"`
	Networkid   string `json:"networkid"`
	Traffictype string `json:"traffictype"`
	Type        string `json:"type"`
}

type Virtualmachine struct {
	Account     string  `json:"account"`
	Cpunumber   float64 `json:"cpunumber"`
	Cpuspeed    float64 `json:"cpuspeed"`
	Created     string  `json:"created"`
	Displayname string  `json:"displayname"`
	Domain      string  `json:"domain"`
	Domainid    string  `json:"domainid"`
	Guestosid   string  `json:"guestosid"`
	Haenable    bool    `json:"haenable"`
	Hypervisor  string  `json:"hypervisor"`
	ID          string  `json:"id"`
	Keypair     string  `json:"keypair"`
	Memory      float64 `json:"memory"`
	Name        string  `json:"name"`
	Nic         []Nic   `json:"nic"`
	Passwordenabled     bool          `json:"passwordenabled"`
	Rootdeviceid        float64       `json:"rootdeviceid"`
	Rootdevicetype      string        `json:"rootdevicetype"`
	Securitygroup       []interface{} `json:"securitygroup"`
	Serviceofferingid   string        `json:"serviceofferingid"`
	Serviceofferingname string        `json:"serviceofferingname"`
	State               string        `json:"state"`
	Tags                []interface{} `json:"tags"`
	Templatedisplaytext string        `json:"templatedisplaytext"`
	Templateid          string        `json:"templateid"`
	Templatename        string        `json:"templatename"`
	Zoneid              string        `json:"zoneid"`
	Zonename            string        `json:"zonename"`
}

type ListVirtualMachinesResponse struct {
	Listvirtualmachinesresponse struct {
		Count          float64 `json:"count"`
		Virtualmachine []Virtualmachine `json:"virtualmachine"`
	} `json:"listvirtualmachinesresponse"`
}

type QueryAsyncJobResultResponse struct {
	Queryasyncjobresultresponse struct {
		Accountid     string  `json:"accountid"`
		Cmd           string  `json:"cmd"`
		Created       string  `json:"created"`
		Jobid         string  `json:"jobid"`
		Jobprocstatus float64 `json:"jobprocstatus"`
		Jobresultcode float64 `json:"jobresultcode"`
		Jobstatus     float64 `json:"jobstatus"`
		Userid        string  `json:"userid"`
	} `json:"queryasyncjobresultresponse"`
}

type Template struct {
	Id   string
	Name string
}

type TemplatesResponse struct {
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

// Deletes an SSH key pair
func (c CloudStackClient) DeleteSSHKeyPair(name string) (string, error) {
	params := url.Values{}
	params.Set("name", name)
	response, err := NewRequest(c, "deleteSSHKeyPair", params)
	if err != nil {
		return "", err
	}
	success := response.(DeleteSshKeyPairResponse).Deletesshkeypairresponse.Success
	return success, err
}

// Deploys a Virtual Machine and returns it's id
func (c CloudStackClient) DeployVirtualMachine(serviceofferingid string, templateid string, zoneid string, networkids []string, keypair string, displayname string, diskoffering string) (string, string, error) {
	params := url.Values{}
	params.Set("serviceofferingid", serviceofferingid)
	params.Set("templateid", templateid)
	params.Set("zoneid", zoneid)
	params.Set("networkids", strings.Join(networkids, ","))
	params.Set("keypair", keypair)
	params.Set("displayname", displayname)
	if diskoffering != "" {
		params.Set("diskoffering", diskoffering)
	}
	response, err := NewRequest(c, "deployVirtualMachine", params)
	if err != nil {
		return "", "", err
	}
	vmid := response.(DeployVirtualMachineResponse).Deployvirtualmachineresponse.ID
	jobid := response.(DeployVirtualMachineResponse).Deployvirtualmachineresponse.Jobid
	return vmid, jobid, nil
}

// Stops a Virtual Machine
func (c CloudStackClient) StopVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "stopVirtualMachine", params)
	if err != nil {
		return "", err
	}
	jobid := response.(StopVirtualMachineResponse).Stopvirtualmachineresponse.Jobid
	return jobid, err
}

// Destroys a Virtual Machine
func (c CloudStackClient) DestroyVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "destroyVirtualMachine", params)
	if err != nil {
		return "", err
	}
	jobid := response.(DestroyVirtualMachineResponse).Destroyvirtualmachineresponse.Jobid
	return jobid, nil
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

// Returns CloudStack string representation of the Virtual Machine state
func (c CloudStackClient) VirtualMachineState(id string) (string, string, error) {
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "listVirtualMachines", params)
	if err != nil {
		return "", "", err
	}

	count := response.(ListVirtualMachinesResponse).Listvirtualmachinesresponse.Count
	if count != 1 {
		// TODO: Return some like no virtual machines found.
		return "", "", err
	}

	state := response.(ListVirtualMachinesResponse).Listvirtualmachinesresponse.Virtualmachine[0].State
	ip := response.(ListVirtualMachinesResponse).Listvirtualmachinesresponse.Virtualmachine[0].Nic[0].Ipaddress

	return ip, state, err
}

// Query CloudStack for the state of a scheduled job
func (c CloudStackClient) QueryAsyncJobResult(jobid string) (float64, error) {
	params := url.Values{}
	params.Set("jobid", jobid)
	response, err := NewRequest(c, "queryAsyncJobResult", params)

	if err != nil {
		return -1, err
	}

	log.Printf("response: %v", response)
	status := response.(QueryAsyncJobResultResponse).Queryasyncjobresultresponse.Jobstatus

	return status, err
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
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signature = url.QueryEscape(signature)

	// Create the final URL before we issue the request
	url := c.BaseURL + "?" + s + "&signature=" + signature

	//log.Printf("Calling %s ", url)

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
		return nil, err
	}

	switch request {
	default:
		log.Printf("Unknown request %s", request)
	case "createSSHKeyPair":
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

	case "stopVirtualMachine":
		var decodedResponse StopVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listVirtualMachines":
		var decodedResponse ListVirtualMachinesResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "queryAsyncJobResult":
		var decodedResponse QueryAsyncJobResultResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil
	}

	// only reached with unknown request
	return "", nil
}
