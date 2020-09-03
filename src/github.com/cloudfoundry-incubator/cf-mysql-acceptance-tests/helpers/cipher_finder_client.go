package helpers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CipherFinderClient struct {
	client *http.Client
	host   string
}

type Pinger interface {
	Ping() error
}

func NewCipherFinderClient(host string, skipSSLValidation bool) CipherFinderClient {
	c := CipherFinderClient{
		host: host,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLValidation},
	}

	c.client = &http.Client{Transport: tr}

	return c
}

func (c CipherFinderClient) Ciphers() (string, error) {
	resp, err := c.do("GET", fmt.Sprintf("%s/ciphers", c.host), "")
	if err != nil {
		return "", err
	}

	var cipher map[string]string

	if err := json.Unmarshal([]byte(resp), &cipher); err != nil {
		return "", err
	}

	return cipher["cipher_used"], nil
}

func (c CipherFinderClient) Ping() error {
	ret, err := c.do("GET", fmt.Sprintf("%s/ping", c.host), "")
	if err != nil {
		return err
	}

	if ret != "OK" {
		return errors.New("did not get an OK")
	}

	return nil
}

func (c CipherFinderClient) do(method, uri, body string) (string, error) {
	req, err := http.NewRequest(method, uri, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s - %s", resp.Status, string(buf))
	}

	return string(buf), nil

}
