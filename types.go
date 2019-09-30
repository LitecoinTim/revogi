package revogi

import (
	"encoding/json"
	"log"
)

type Revogi struct {
	client  Client
	config  Config
	domain  string
	token   string
	devices []Device
	logger  *log.Logger
}

type Config struct {
	Username   string
	Password   string
	Interval   int
	MaxRetries int
	Cooldown   int
	Devices    map[string]DeviceConfig
	Logger     *log.Logger
}

type DeviceConfig struct {
	State map[int]int
	Usage map[int]int
}

type SocketNumber int

type DeviceStats struct {
	Softver string `json:"softver"`
	Amp     []int  `json:"amp"`
	Online  int    `json:"online"`
	Sn      string `json:"sn"`
	Watt    []int  `json:"watt"`
	Switch  []int  `json:"switch"`
}

type Device struct {
	Ver        string   `json:"ver"`
	Pname      []string `json:"pname"`
	Nver       string   `json:"nver"`
	Line       int      `json:"line"`
	SocketType string   `json:"socket_type"`
	Ip         string   `json:"ip"`
	Mac        string   `json:"mac"`
	DateAdded  string   `json:"dateAdd"`
	Name       string   `json:"name"`
	GatewayIp  string   `json:"gateway_ip"`
	Sn         string   `json:"sn"`
	Protect    int      `json:"protect"`
	Sak        string   `json:"sak"`
	Register   int      `json:"register"`
	Stats      *DeviceStats
}

type devicesData struct {
	Dev []Device `json:"dev"`
	Url string   `json:"url"`
}

type revogiResponseBody struct {
	Code     int             `json:"code"`
	Data     json.RawMessage `json:"data"`
	Response int             `json:"response"`
	Sn       string          `json:"sn"`
}

type loginResponse struct {
	revogiResponseBody
	Data loginData `json:"data"`
}

type loginData struct {
	UserId  string `json:"user_id"`
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	RegId   string `json:"regid"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Url     string `json:"url"`
	Token   string `json:"token"`
}

type revogiDevicesResponseBody struct {
	revogiResponseBody
	Data devicesData `json:"data"`
}

type revogiDeviceStatsResponseBody struct {
	revogiResponseBody
	Data []DeviceStats `json:"data"`
}