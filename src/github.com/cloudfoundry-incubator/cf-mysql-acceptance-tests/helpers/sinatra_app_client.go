package helpers

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SinatraAppClient struct {
	client          *http.Client
	host            string
	serviceInstance string
}

func NewSinatraAppClient(host string, serviceInstance string, skipSSLValidation bool) SinatraAppClient {
	c := SinatraAppClient{
		host:            host,
		serviceInstance: serviceInstance,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLValidation},
	}

	c.client = &http.Client{Transport: tr}

	return c
}

func (c SinatraAppClient) WriteBulkData(megabytes string) (string, error) {
	return c.do("POST", fmt.Sprintf("%s/service/mysql/%s/write-bulk-data", c.host, c.serviceInstance), megabytes)
}

func (c SinatraAppClient) DeleteBulkData(megabytes string) (string, error) {
	return c.do("POST", fmt.Sprintf("%s/service/mysql/%s/delete-bulk-data", c.host, c.serviceInstance), megabytes)
}

func (c SinatraAppClient) Set(key, value string) (string, error) {
	return c.do("POST", fmt.Sprintf("%s/service/mysql/%s/%s", c.host, c.serviceInstance, key), value)
}

func (c SinatraAppClient) Get(key string) (string, error) {
	return c.do("GET", fmt.Sprintf("%s/service/mysql/%s/%s", c.host, c.serviceInstance, key), "")
}

func (c SinatraAppClient) Ping() error {
	ret, err := c.do("GET", fmt.Sprintf("%s/ping", c.host), "")
	if err != nil {
		return err
	}

	if ret != "OK" {
		return errors.New("did not get an OK")
	}

	return nil
}

func (c SinatraAppClient) do(method, uri, body string) (string, error) {
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
