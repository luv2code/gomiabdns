package gomiabdns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RecordType is the type of DNS Record.
type RecordType string

const (
	A     RecordType = "A"
	AAAA  RecordType = "AAAA"
	CAA   RecordType = "CAA"
	CNAME RecordType = "CNAME"
	MX    RecordType = "MX"
	NS    RecordType = "NS"
	TXT   RecordType = "TXT"
	SRV   RecordType = "SRV"
	SSHFP RecordType = "SSHFP"
)

type Client struct {
	ApiUrl *url.URL
}

func New(apiUrl, email, password string) *Client {
	parsedUrl, err := url.Parse(apiUrl)
	parsedUrl.User = url.UserPassword(email, password)
	if err != nil {
		panic(err)
	}
	return &Client{
		ApiUrl: parsedUrl,
	}
}

func (c *Client) GetHosts(ctx context.Context, name string, recordType RecordType) ([]DNSRecord, error) {
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodGet, apiUrl.String(), "")
	if err != nil {
		return nil, err
	}
	return unmarshalRecords(apiResp)
}

func (c *Client) AddHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" || recordType == "" || value == "" {
		return fmt.Errorf("Missing parameters to AddHost. all are required. name: %s, recordType: %s, value: %s ", name, recordType, value)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodPost, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

func (c *Client) UpdateHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" || recordType == "" || value == "" {
		return fmt.Errorf("Missing parameters to UpdateHost. all are required. name: %s, recordType: %s, value: %s ", name, recordType, value)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodPut, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

func (c *Client) DeleteHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" {
		return fmt.Errorf("Missing parameter to DeleteHost. Name is required. name: %s", name)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodDelete, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

func doRequest(ctx context.Context, method, requestURL, value string) ([]byte, error) {
	var r io.Reader
	if value != "" {
		r = strings.NewReader(value)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, r)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getApiWithPath(apiUrl *url.URL, name string, rtype RecordType) *url.URL {
	if name != "" {
		if rtype != "" {
			return apiUrl.JoinPath(name, string(rtype))
		} else {
			return apiUrl.JoinPath(name)
		}
	}
	return apiUrl
}

func unmarshalRecords(data []byte) ([]DNSRecord, error) {
	var result []DNSRecord
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

type DNSRecord struct {
	QualifiedName string     `json:"qname"`
	RecordType    RecordType `json:"rtype"`
	SortOrder     struct {
		ByCreated int `json:"created"`
		ByName    int `json:"qname"`
	} `json:"sort-order"`
	Value string `json:"value"`
	Zone  string `json:"zone"`
}
