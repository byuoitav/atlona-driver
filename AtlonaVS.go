package atlona

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/wspool"
)

type DeviceType int

type AtlonaVideoSwitcher interface {
	GetAudioVideoInputs(ctx context.Context) (map[string]string, error)
	SetAudioVideoInput(ctx context.Context, output, input string) error

	SetVolume(ctx context.Context, block string, volume int) error
	SetMute(ctx context.Context, block string, muted bool) error

	GetVolumes(ctx context.Context, blocks []string) (map[string]int, error)
	GetMutes(ctx context.Context, blocks []string) (map[string]bool, error)

	GetInfo(ctx context.Context) (interface{}, error)
}

func CreateVideoSwitcher(ctx context.Context, addr, username, password string, log wspool.Logger) (AtlonaVideoSwitcher, error) {
	url := fmt.Sprintf("http://%s/", addr)

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response: %w", err)
	}

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
			Username: username,
			Password: password,
			Address:  addr,
		}
		return Atlonavs, nil
	case "AT-UHD-SW-52ED":
		Atlonavs := &AtlonaVideoSwitcher5x1{
			Username: username,
			Password: password,
			Address:  addr,
			Logger:   log,
		}
		return Atlonavs, nil
	case "AT-JUNO-451-HDBT", "AT-JUNO-451":
		Atlonavs := &AtlonaVideoSwitcher4x1{
			Username: username,
			Password: password,
			Address:  addr,
		}
		return Atlonavs, nil
	case "AT-HDVS-210U":
		Atlonavs := &AtlonaVideoSwitcher2x1{
			Username: username,
			Password: password,
			Address:  addr,
		}
		return Atlonavs, nil
	default:
		return nil, fmt.Errorf("unknown device type %v", deviceType)
	}
}
