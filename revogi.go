package revogi

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	apiHost  = "server.revogi.net"
	protocol = "3"
)

type Client interface {
	Login() error
	GetDevice(dev Device) (Device, error)
	GetDevices() ([]Device, error)
	GetDeviceStats(device Device) (DeviceStats, error)
	Power(device Device, port int, state bool) error
	doRequest(cmd int, postData string, tokenLogin bool) (respBody *revogiResponseBody, err error)
}

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

type client struct {
	logger *log.Logger
	client HttpClient
	config Config

	domain  string
	token   string
	retries int
}

func NewClient(httpClient HttpClient, config Config) Client {
	c := &client{
		client: httpClient,
		config: config,
	}
	if config.Logger == nil {
		c.logger = log.New(ioutil.Discard, "", 0)
	}
	return c
}

func (c *client) Login() error {
	postData, err := json.Marshal(map[string]string{
		"username": c.config.Username,
		"password": c.config.Password,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal login post data")
	}

	body, err := c.doRequest(101, string(postData), false)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	loginResult := loginResponse{*body, loginData{}}
	err = json.Unmarshal(body.Data, &loginResult.Data)
	if err != nil {
		return err
	}

	c.domain = loginResult.Data.Domain
	c.token = loginResult.Data.Token

	return nil
}

func (c *client) doRequest(cmd int, postData string, tokenLogin bool) (respBody *revogiResponseBody, err error) {
	postBody := url.Values{
		"cmd":  {strconv.Itoa(cmd)},
		"json": {postData},
	}

	if tokenLogin {
		postBody.Add("tokenlogin", c.token)
	}

	var target = c.domain
	if c.domain == "" {
		target = apiHost
	}

	resp, err := c.client.PostForm(fmt.Sprintf("https://%s/services/ajax.html", target), postBody)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("unexpected http response code: %d", resp.StatusCode)
	}

	respBody = &revogiResponseBody{}
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}

	if tokenLogin && respBody.Code == 401 {
		if c.retries >= c.config.MaxRetries {
			return nil, errors.Errorf("reached max token renewal attempts (%d)", c.config.MaxRetries)
		}

		c.logger.Println("token expired. attempting to renew")
		err := c.Login()

		if err != nil {
			c.retries++
			return nil, errors.Wrapf(err, "renew failed")
		}
		c.retries = 0
		return c.doRequest(cmd, postData, tokenLogin)
	}

	if respBody.Code != 200 {
		if respBody.Code == 500 {
			time.Sleep(time.Second * time.Duration(c.config.Cooldown))
		}
		return nil, errors.Errorf("unexpected json response code: %d, body: %s\n", respBody.Code, string(respBody.Data))
	}

	return
}

func (c *client) GetDevice(dev Device) (Device, error) {
	var device Device

	postData, err := json.Marshal(map[string]string{
		"protocol": protocol,
		"dev":      dev.Sn,
	})
	if err != nil {
		return device, err
	}

	body, err := c.doRequest(500, string(postData), true)
	if err != nil {
		return device, err
	}

	devicesResult := &revogiDevicesResponseBody{*body, devicesData{}}
	err = json.Unmarshal(body.Data, &devicesResult.Data)
	if err != nil {
		return device, err
	}

	if len(devicesResult.Data.Dev) > 0 {
		device = devicesResult.Data.Dev[0]
	}
	return device, nil
}

func (c *client) GetDevices() ([]Device, error) {
	postData, err := json.Marshal(map[string]string{
		"protocol": protocol,
		"dev":      "all",
	})
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(500, string(postData), true)
	if err != nil {
		return nil, err
	}

	devicesResult := &revogiDevicesResponseBody{*body, devicesData{}}
	err = json.Unmarshal(body.Data, &devicesResult.Data)
	if err != nil {
		return nil, err
	}

	return devicesResult.Data.Dev, nil
}

func (c *client) GetDeviceStats(device Device) (DeviceStats, error) {
	stats := DeviceStats{}
	postData, err := json.Marshal(map[string]interface{}{
		"protocol": protocol,
		"sn":       []string{device.Sn},
	})
	if err != nil {
		return stats, err
	}

	body, err := c.doRequest(511, string(postData), true)
	if err != nil {
		return stats, err
	}

	deviceStatsResult := &revogiDeviceStatsResponseBody{*body, []DeviceStats{}}
	err = json.Unmarshal(body.Data, &deviceStatsResult.Data)
	if err != nil {
		return stats, err
	}

	return deviceStatsResult.Data[0], nil
}

func (c *client) Power(device Device, port int, state bool) error {
	var intState int
	if state {
		intState = 1
	} else {
		intState = 0
	}

	postData, err := json.Marshal(map[string]interface{}{
		"protocol": protocol,
		"sn":       device.Sn,
		"port":     port,
		"state":    intState,
	})
	if err != nil {
		return err
	}

	body, err := c.doRequest(200, string(postData), true)
	if err != nil {
		return err
	}

	result, err := strconv.Unquote(string(body.Data))
	if err != nil {
		return errors.Wrapf(err, "failed to unquote body: %s", string(body.Data))
	}

	if results := strings.Split(result, ":"); results[0] == "send2" && results[1] == device.Sn {
		return nil
	}

	return errors.Errorf("unexpected result: %s", result)
}
