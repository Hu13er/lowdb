package lowdb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	UnexpectedServerValue = errors.New("unexpected value")
	UnknownError          = errors.New("unknown error")
)

type KVStoreHTTPClient struct {
	Client http.Client
}

var _ KVStore = (*KVStoreHTTPClient)(nil)

func NewKVStoreHTTPClient(socketFile string) KVStore {
	dialer := func(network, addr string) (net.Conn, error) {
		return net.Dial("unix", socketFile)
	}

	return &KVStoreHTTPClient{
		Client: http.Client{
			Transport: &http.Transport{
				Dial: dialer,
			},
		},
	}
}

func (c *KVStoreHTTPClient) Keys() ([]string, error) {
	resp, err := c.Client.Get("http://localhost/")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	keys := strings.Split(strings.TrimSpace(string(body)), "\n")
	return keys, nil
}

func (c *KVStoreHTTPClient) Get(key string) (KeyValueMetadata, error) {
	var kvm KeyValueMetadata
	url := fmt.Sprintf("http://localhost/%s/", key)
	resp, err := c.Client.Get(url)
	if err != nil {
		return kvm, err
	}

	kvm.Key = resp.Header.Get("Lowdb-Key")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return kvm, err
	}
	kvm.Value = body

	revStr := resp.Header.Get("Lowdb-Revision")
	rev, err := strconv.Atoi(revStr)
	if err != nil {
		return kvm, UnexpectedServerValue
	}
	kvm.Revision = rev

	createdAtStr := resp.Header.Get("Lowdb-Created_at")
	createdAt, err := time.Parse(time.ANSIC, createdAtStr)
	if err != nil {
		return kvm, UnexpectedServerValue
	}
	kvm.CreatedAt = createdAt

	kvm.Headers = make(map[string][]string)
	for k, vs := range resp.Header {
		trimedk := strings.TrimPrefix(k, "Lowdb-Meta-")
		if k == trimedk {
			continue
		}
		kvm.Headers[trimedk] = slices.Clone(vs)
	}
	return kvm, nil
}

func (c *KVStoreHTTPClient) Set(data KeyValueMetadata) error {
	url := fmt.Sprintf("http://localhost/%s/", data.Key)
	buffer := bytes.NewReader(data.Value)
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return err
	}
	req.Header.Add("Lowdb-Key", data.Key)
	if data.Revision >= 0 {
		req.Header.Add("Lowdb-Revision", strconv.Itoa(data.Revision))
	}
	for k, vs := range data.Headers {
		for _, v := range vs {
			req.Header.Add("Lowdb-Meta-"+k, v)
		}
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return InvalidRevsion
	default:
		return UnknownError
	}
}

func (c *KVStoreHTTPClient) Delete(data KeyValueMetadata) error {
	url := fmt.Sprintf("http://localhost/%s/", data.Key)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Lowdb-Key", data.Key)
	if data.Revision >= 0 {
		req.Header.Add("Lowdb-Revision", strconv.Itoa(data.Revision))
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return InvalidRevsion
	default:
		return UnknownError
	}
}
