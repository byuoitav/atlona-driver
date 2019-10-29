package atlona

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/common/structs"
)

type DeviceType int

const (
	Undefined DeviceType = iota
	Atlona6x2
	Atlona5x1
	Atlona4x1
	Atlona2x1
)

//VideoSwitcher6x2 .
type AtlonaVideoSwitcher struct {
	Username   string
	Password   string
	Address    string
	LastLogin  time.Time
	DeviceType DeviceType
}

func GetDeviceType(ctx context.Context, addr string) (DeviceType, error) {
	url := fmt.Sprintf("http://%s/", addr)

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Undefined, fmt.Errorf("Error: %w", err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	//regex black magic
	reg, err := regexp.Compile("<title[^>]*>([^<]+)</title>")
	if err != nil {
		return Undefined, fmt.Errorf("Error: %w", err)
	}
	regexType := reg.FindAllStringSubmatch(fmt.Sprintf("%s", body), -1)
	deviceType := regexType[0][1]
	deviceType = strings.Replace(deviceType, "Login", "", -1)
	deviceType = strings.Replace(deviceType, " ", "", -1)

	switch deviceType {
	case "AT-OME-PS62":
		return Atlona6x2, nil
	case "AT-UHD-SW-52ED":
		return Atlona5x1, nil
	case "AT-JUNO-451-HDBT":
		return Atlona4x1, nil
	case "AT-HDVS-210U":
		return Atlona2x1, nil
	default:
		return Undefined, fmt.Errorf("unknown device type")
	}
}

//GetInputByOutput .
func (vs *AtlonaVideoSwitcher) GetInputByOutput(ctx context.Context, output string) (string, error) {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.getInputByOutput6x2(ctx, output)
	case Atlona5x1:
		return vs.getInputByOutput6x2(ctx, output)
	case Atlona4x1:
		return vs.getInputByOutput6x2(ctx, output)
	case Atlona2x1:
		return vs.getInputByOutput6x2(ctx, output)
	default:
		return "", fmt.Errorf("unknown device type")
	}
}

//SetInputByOutput .
func (vs *AtlonaVideoSwitcher) SetInputByOutput(ctx context.Context, output, input string) error {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.setInputByOutput6x2(ctx, output, input)
	case Atlona5x1:
		return vs.setInputByOutput5x1(ctx, output, input)
	case Atlona4x1:
		return vs.setInputByOutput4x1(ctx, output, input)
	case Atlona2x1:
		return vs.setInputByOutput2x1(ctx, output, input)
	default:
		return fmt.Errorf("unknown device type")
	}
}

//SetVolumeByBlock .
func (vs *AtlonaVideoSwitcher) SetVolumeByBlock(ctx context.Context, output string, level int) error {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.setVolumeByBlock6x2(ctx, output, level)
	case Atlona5x1:
		return vs.setVolumeByBlock5x1(ctx, output, level)
	default:
		return fmt.Errorf("unknown device type")
	}
}

//GetVolumeByBlock .
func (vs *AtlonaVideoSwitcher) GetVolumeByBlock(ctx context.Context, output string) (int, error) {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.getVolumeByBlock6x2(ctx, output)
	case Atlona5x1:
		return vs.getVolumeByBlock5x1(ctx, output)
	default:
		return 0, fmt.Errorf("unknown device type")
	}
}

//GetMutedByBlock .
func (vs *AtlonaVideoSwitcher) GetMutedByBlock(ctx context.Context, output string) (bool, error) {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.getMutedByBlock6x2(ctx, output)
	case Atlona5x1:
		return vs.getMutedByBlock5x1(ctx, output)
	default:
		return false, fmt.Errorf("unknown device type")
	}
}

//SetMutedByBlock .
func (vs *AtlonaVideoSwitcher) SetMutedByBlock(ctx context.Context, output string, muted bool) error {
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.setMutedByBlock6x2(ctx, output, muted)
	case Atlona5x1:
		return vs.setMutedByBlock5x1(ctx, output, muted)
	default:
		return fmt.Errorf("unknown device type")
	}
}

//GetHardwareInfo .
func (vs *AtlonaVideoSwitcher) GetHardwareInfo(ctx context.Context) (structs.HardwareInfo, error) {
	var resp structs.HardwareInfo
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.getHardwareInfo6x2(ctx)
	case Atlona5x1:
		return vs.getHardwareInfo5x1(ctx)
	case Atlona4x1:
		return vs.getHardwareInfo4x1(ctx)
	case Atlona2x1:
		return vs.getHardwareInfo2x1(ctx)
	default:
		return resp, fmt.Errorf("unknown device type")
	}
}

//GetInfo .
func (vs *AtlonaVideoSwitcher) GetInfo(ctx context.Context) (interface{}, error) {
	var info interface{}
	switch vs.DeviceType {
	case Atlona6x2:
		return vs.getInfo6x2(ctx)
	case Atlona5x1:
		return vs.getInfo5x1(ctx)
	case Atlona4x1:
		return vs.getInfo4x1(ctx)
	case Atlona2x1:
		return vs.getInfo2x1(ctx)
	default:
		return info, fmt.Errorf("unknown device type")
	}
}
