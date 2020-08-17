package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BaseURL = "https://api.dns.constellix.com/"

var mux, headerMux sync.Mutex

type Client struct {
	httpclient         *http.Client
	apiKey             string //Required
	secretKey          string //Required
	insecure           bool   //Optional
	proxyurl           string //Optional
	reqRemCount        int    //Optional
	reqRefreshInterval int    //Optional
	totalReq           int    //Optional
}

//singleton implementation of a client
var clietnImpl *Client

type Option func(*Client)

func Insecure(insecure bool) Option {
	return func(client *Client) {
		client.insecure = insecure
	}
}

func ProxyUrl(pUrl string) Option {
	return func(client *Client) {
		client.proxyurl = pUrl
	}
}

func initClient(apiKey, secretKey string, options ...Option) *Client {
	//existing information about client
	client := &Client{
		apiKey:             apiKey,
		secretKey:          secretKey,
		reqRemCount:        30,
		reqRefreshInterval: 30,
	}
	for _, option := range options {
		option(client)
	}

	//Setting up the HTTP client for the API call
	var transport *http.Transport
	transport = client.useInsecureHTTPClient(client.insecure)
	if client.proxyurl != "" {
		transport = client.configProxy(transport)
	}
	client.httpclient = &http.Client{
		Transport: transport,
	}
	return client
}

//Returns a singleton
func GetClient(apiKey, secretKey string, options ...Option) *Client {
	clietnImpl = initClient(apiKey, secretKey, options...)
	return clietnImpl
}

func (c *Client) useInsecureHTTPClient(insecure bool) *http.Transport {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       insecure,
			MinVersion:               tls.VersionTLS11,
			MaxVersion:               tls.VersionTLS12,
		},
	}

	return transport
}

func (c *Client) configProxy(transport *http.Transport) *http.Transport {
	pUrl, err := url.Parse(c.proxyurl)
	if err != nil {
		log.Fatal(err)
	}
	transport.Proxy = http.ProxyURL(pUrl)
	return transport
}

func getToken(apiKey, secretKey string) string {
	//Extracts epoch time in miliseconds
	time := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	//Calculate hmac using secrest key and epoch time
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(time))
	sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

	//Building token as 'apikey:hmac:time'
	token := string(apiKey) + ":" + string(sha) + ":" + string(time)
	return token
}

func (c *Client) makeRequest(method, endpoint string, payload []byte) (*http.Request, error) {
	//Defining http request
	mux.Lock()
	if c.reqRemCount <= 2 {
		time.Sleep(time.Second * time.Duration(c.reqRefreshInterval))
		log.Println("Sleeeeeeep")
	}
	mux.Unlock()

	var req *http.Request
	var err error
	if method == "POST" || method == "PUT" {
		req, err = http.NewRequest(method, endpoint, bytes.NewBuffer(payload))
	} else {
		req, err = http.NewRequest(method, endpoint, nil)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("cooouni %d %v", c.reqRemCount, req)
	//Calling for token and setting headers
	token := getToken(c.apiKey, c.secretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-cns-security-token", token)

	return req, nil
}

func (c *Client) Save(obj interface{}, endpoint string) (responce *http.Response, err error) {
	jsonPayload, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}

	urlArr := strings.Split(endpoint, "/")
	var url string
	var flag bool
	if len(urlArr) > 2 && urlArr[2] == "api.sonar.constellix.com" {
		url = endpoint
		flag = true
	} else {
		url = fmt.Sprintf("%s%s", BaseURL, endpoint)
	}

	req, err1 := c.makeRequest("POST", url, jsonPayload)
	log.Println(req, c.reqRemCount)
	if err1 != nil {
		return nil, err1
	}

	resp, err2 := c.Do(req)
	if err2 != nil {
		return nil, err2
	}
	log.Println(resp)
	if flag == false {
		return resp, checkForErrors(resp)
	}
	return resp, checkForErrorsChecks(resp)
}

func checkForErrors(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		var data map[string]interface{}
		json.Unmarshal([]byte(bodyString), &data)

		var errors []interface{}
		errors = data["errors"].([]interface{})

		var allerrs string
		for _, val := range errors {
			allerrs = allerrs + val.(string)
		}
		log.Println(" Errors are .....:: ", allerrs)
		return fmt.Errorf("%s", allerrs)
	}
	return nil
}

func checkForErrorsChecks(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != 201 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		return fmt.Errorf("%s", bodyString)
	}
	return nil
}

func (c *Client) GetbyId(endpoint string) (response *http.Response, err error) {
	var url string
	var flag bool
	urlArr := strings.Split(endpoint, "/")
	if len(urlArr) > 2 && urlArr[2] == "api.sonar.constellix.com" {
		url = endpoint
		flag = true
	} else {
		url = fmt.Sprintf("%s%s", BaseURL, endpoint)
	}

	req, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	log.Println("In GET by ID :", req)

	resp, err1 := c.Do(req)
	if err1 != nil {
		return nil, err1
	}

	log.Println("Response for Get: ", resp)
	if flag == false {
		return resp, checkForErrors(resp)
	}
	return resp, checkForErrorsChecks(resp)
}

func (c *Client) DeletebyId(endpoint string) error {
	var url string
	urlArr := strings.Split(endpoint, "/")
	if len(urlArr) > 2 && urlArr[2] == "api.sonar.constellix.com" {
		url = endpoint
	} else {
		url = fmt.Sprintf("%s%s", BaseURL, endpoint)
	}

	req, err := c.makeRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err1 := c.Do(req)
	if err1 != nil {
		return err1
	}
	return nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpclient.Do(req)
	headerMux.Lock()
	for k, v := range resp.Header {
		log.Printf("herrrr %s", k)
		if k == "requestsRemainingHeader" {
			tempInt, err := strconv.Atoi(v[0])
			if err == nil {
				c.reqRemCount = tempInt
			}
			log.Printf("Remhed found %d", tempInt)
		}
		if k == "requestRefreshInterval" {
			tempInt, err := strconv.Atoi(v[0])
			if err == nil {
				c.reqRefreshInterval = tempInt
			}
		}
	}
	c.totalReq++
	headerMux.Unlock()
	log.Printf("toooot %d", c.totalReq)
	return resp, err
}
func (c *Client) UpdatebyID(obj interface{}, endpoint string) (response *http.Response, err error) {
	jsonPayload, err := json.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}
	urlArr := strings.Split(endpoint, "/")
	var url string
	var flag bool
	if len(urlArr) > 2 && urlArr[2] == "api.sonar.constellix.com" {
		url = endpoint
		flag = true
	} else {
		url = fmt.Sprintf("%s%s", BaseURL, endpoint)
	}

	req, err1 := c.makeRequest("PUT", url, jsonPayload)
	log.Println(req)
	if err1 != nil {
		return nil, err1
	}

	resp, err2 := c.Do(req)
	if err2 != nil {
		return nil, err2
	}
	log.Println(resp)
	if flag == false {
		return resp, checkForErrors(resp)
	}
	return resp, checkForErrorsChecks(resp)
}
