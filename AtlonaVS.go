package atlona

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
)

type DeviceType int

const (
	Undefined DeviceType = iota
	Atlona6x2
	Atlona5x1
	Atlona4x1
	Atlona2x1
)

type AtlonaVideoSwitcher interface {
	GetInputByOutput(ctx context.Context, output string) (string, error)
	SetInputByOutput(ctx context.Context, output, input string) error

	SetVolumeByBlock(ctx context.Context, block string, volume int) error
	SetMutedByBlock(ctx context.Context, block string, muted bool) error

	GetVolumeByBlock(ctx context.Context, block string) (int, error)
	GetMutedByBlock(ctx context.Context, block string) (bool, error)

	GetInfo(ctx context.Context) (interface{}, error)
}

type AtlonaVideoSwitcher2x1 struct {
	Username   string
	Password   string
	Address    string
	DeviceType DeviceType
}

type AtlonaVideoSwitcher4x1 struct {
	Username   string
	Password   string
	Address    string
	DeviceType DeviceType
}

type AtlonaVideoSwitcher5x1 struct {
	Username   string
	Password   string
	Address    string
	DeviceType DeviceType
	ws         *websocket.Conn
}

type AtlonaVideoSwitcher6x2 struct {
	Username   string
	Password   string
	Address    string
	DeviceType DeviceType
}

func createVideoSwitcher(ctx context.Context, addr string) (AtlonaVideoSwitcher, error) {

	url := fmt.Sprintf("http://%s/", addr)

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error: %w", err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	//regex black magic
	reg, err := regexp.Compile("<title[^>]*>([^<]+)</title>")
	if err != nil {
		return nil, fmt.Errorf("Error: %w", err)
	}
	regexType := reg.FindAllStringSubmatch(fmt.Sprintf("%s", body), -1)
	deviceType := regexType[0][1]
	deviceType = strings.Replace(deviceType, "Login", "", -1)
	deviceType = strings.Replace(deviceType, " ", "", -1)

	switch deviceType {
	case "AT-OME-PS62":
		Atlonavs := &AtlonaVideoSwitcher6x2{
			Address:    addr,
			DeviceType: Atlona6x2,
		}
		return Atlonavs, nil
	case "AT-UHD-SW-52ED":
		Atlonavs := &AtlonaVideoSwitcher5x1{
			Address:    addr,
			DeviceType: Atlona5x1,
		}
		return Atlonavs, nil
	case "AT-JUNO-451-HDBT":
		Atlonavs := &AtlonaVideoSwitcher4x1{
			Address:    addr,
			DeviceType: Atlona4x1,
		}
		return Atlonavs, nil
	case "AT-HDVS-210U":
		Atlonavs := &AtlonaVideoSwitcher2x1{
			Address:    addr,
			DeviceType: Atlona2x1,
		}
		return Atlonavs, nil
	default:
		return nil, fmt.Errorf("unknown device type")
	}
}
