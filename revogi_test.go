package revogi

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type mockHttpClient struct {
	err         error
	lastUrl     *url.URL
	body        []byte
	statusCode  int
	lastRequest *http.Request
	tokenLogin  bool
}

func (m mockHttpClient) Get(url string) (resp *http.Response, err error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *mockHttpClient) PostForm(u string, data url.Values) (resp *http.Response, err error) {
	if m.err != nil {
		return nil, m.err
	}

	Url, _ := url.Parse(u)
	if err != nil {
		return
	}

	resp = &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(m.body)),
		StatusCode: m.statusCode,
		Request: &http.Request{
			Form: data,
			URL:  Url,
		},
	}
	return resp, nil
}

func TestRevogi_LoginOk(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":{},"response":101,"sn":""}`),
	}

	// Implement invalid username/password response by code: ???

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Login(); err != nil {
		t.Fatalf("login error: %v", err)
	}
}

func TestRevogi_LoginHttpNonOk(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 500,
		body:       []byte(`{"code":500,"data":{},"response":101,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Login(); err == nil {
		t.Fatal("expected error on login, did not get one")
	}
}

func TestRevogi_LoginErr(t *testing.T) {
	mock := &mockHttpClient{
		err: errors.New("failed to login"),
	}

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Login(); err == nil {
		t.Fatal("expected error on login, did not get one")
	}
}

func TestRevogi_LoginInvalidResponse(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":500,"data":"{}","response":101,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Login(); err == nil {
		t.Fatal("expected error on login, did not get one")
	}
}

func TestRevogi_LoginInvalidData(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":{"incomplete":"data...,"response":101,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Login(); err == nil {
		t.Fatal("expected error on login, did not get one")
	}
}

func TestClient_GetDevice(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":{"dev":[{"ver":"5.12","pname":[],"nver":"0.00","line":1,"socket_type":"WebSocket","ip":"","name":"SmartStrip","gateway_ip":"127.0.0.1","sn":"1234","register":1}],"url":"http:\/\/lon.revogi.net\/services\/ajax.html"},"response":500,"sn":""}`),
	}
	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if _, err := c.GetDevice(Device{Sn: "1234"}); err != nil {
		t.Fatalf("an error occured: %v", err)
	}
}

func TestClient_GetDevices(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":{"dev":[{"ver":"5.12","pname":["PORT 1","PORT 2","PORT 3","PORT 4","PORT 5","PORT 6"],"nver":"0.00","line":1,"socket_type":"WebSocket","ip":"192.168.1.20","mac":"00:00:00:00:00:00","dateAdd":"2018-01-01 00:00:00","name":"SmartStrip","gateway_ip":"127.0.0.1","sn":"SWW6010040000001","protect":0,"sak":"222B2022222B","register":1}],"url":"http:\/\/lon.revogi.net\/services\/ajax.html"},"response":500,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	_, err := c.GetDevices()
	if err != nil {
		t.Fatalf("error getting devices: %v", err)
	}
}

func TestClient_GetDevicesNeedLogin(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":401,"data":"please log in","response":500,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username:   "user",
		Password:   "password",
		MaxRetries: 3,
	})

	_, err := c.GetDevices()
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
}

func TestClient_MaxRetriesReached(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":401,"data":"please log in","response":500,"sn":""}`),
	}

	c := NewClient(mock, Config{
		Username:   "user",
		Password:   "password",
		MaxRetries: 0,
	})

	_, err := c.GetDevices()
	if err == nil || !strings.Contains(err.Error(), "reached max token renewal attempts") {
		t.Fatalf("expected a max retry error")
	}
}

func TestClient_GetDeviceStats(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":[{"softver":"5.12","amp":[0,0,0,0,0,0],"online":1,"sn":"SWW6010040000451","watt":[0,0,0,0,0,0],"switch":[0,0,0,0,0,0]}],"response":511,"sn":""}`),
	}
	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if _, err := c.GetDeviceStats(Device{Sn: "1234"}); err != nil {
		t.Fatalf("an error occured: %v", err)
	}
}

func TestClient_Power(t *testing.T) {
	mock := &mockHttpClient{
		statusCode: 200,
		body:       []byte(`{"code":200,"data":"send2:SWW1234","response":200,"sn":"SWW1234"}`),
	}
	c := NewClient(mock, Config{
		Username: "user",
		Password: "password",
	})

	if err := c.Power(Device{Sn: "SWW1234"}, 0, true); err != nil {
		t.Fatalf("an error occured: %v", err)
	}
}
