package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/wspool"
)

const (
	avSettingsPage = "avs"
	infoPage       = "info"
)

type AtlonaVideoSwitcher4x1 struct {
	Username string
	Password string
	Address  string
}

// AVSettings is the response from the switcher for the av settings page
type AVSettings struct {
	HDMIInputAudioBreakout int   `json:"ARC"`
	HDCPSettings           []int `json:"HDCPSet"`
	AudioOutput            int   `json:"HDMIAud"`
	Toslink                int   `json:"Toslink"`
	AutoSwitch             int   `json:"asw"`
	Input                  int   `json:"inp"`
	LoggedIn               int   `json:"login_ur"`
}

// Info is the response from the switcher for the info page
type Info struct {
	SystemInfo []string      `json:"info_val1"`
	VideoInfo  []interface{} `json:"info_val2"`
	LoggedIn   int           `json:"login_ur"`
}

// SystemSettings .
type SystemSettings struct {
}

func getPage(ctx context.Context, address, page string, structToFill interface{}) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/aj.html?a=%s", address, page), nil)
	if err != nil {
		return fmt.Errorf("unable to get page %s on %s", page, address)
	}

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to get page %s on %s", page, address)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to get page %s on %s", page, address)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("unable to get page %s on %s - %v response recevied. body: %s", page, address, resp.StatusCode, b)
	}

	err = json.Unmarshal(b, structToFill)
	if err != nil {
		return fmt.Errorf("unable to get page %s on %s", page, address)
	}

	return nil
}

func sendCommand(ctx context.Context, address, command string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/aj.html?a=command&cmd=%s", address, command), nil)
	if err != nil {
		return fmt.Errorf("unable to send command '%s' to %s", command, address)
	}

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send command '%s' to %s", command, address)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to send command '%s' to %s", command, address)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("unable to send command '%s' to %s - %v response received. body: %s", command, address, resp.StatusCode, b)
	}

	return nil
}

// TODO finish this :)
func getNetworkSettings(ctx context.Context, address string) (structs.NetworkInfo, error) {
	var info structs.NetworkInfo

	// get the ip info (bleh, gross. it's in the html)
	req, gerr := http.NewRequest("GET", fmt.Sprintf("http://%v", address), nil)
	if gerr != nil {
		return info, fmt.Errorf("unable to get network settings from %s:%w", address, gerr)
	}

	req = req.WithContext(ctx)
	resp, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return info, fmt.Errorf("unable to get network settings from %s:%w", address, gerr)
	}
	defer resp.Body.Close()

	b, gerr := ioutil.ReadAll(resp.Body)
	if gerr != nil {
		return info, fmt.Errorf("unable to get network settings from %s:%w", address, gerr)
	}

	if resp.StatusCode/100 != 2 {
		return info, fmt.Errorf("unable to get network settings from %s - %v response received. body: %s", address, resp.StatusCode, b)
	}

	return info, nil
}

// GetInput returns the current input
func (vs *AtlonaVideoSwitcher4x1) GetInputByOutput(ctx context.Context, output string) (string, error) {
	var settings AVSettings
	err := getPage(ctx, vs.Address, avSettingsPage, &settings)
	if err != nil {
		return "", fmt.Errorf("unable to get input: %w", err)
	}

	return fmt.Sprintf("%v", settings.Input-1), nil
}

// GetHardwareInfo returns a hardware info struct
func (vs *AtlonaVideoSwitcher4x1) GetHardwareInfo(ctx context.Context) (structs.HardwareInfo, error) {
	var hwinfo structs.HardwareInfo

	var info Info
	err := getPage(ctx, vs.Address, infoPage, &info)
	if err != nil {
		return hwinfo, fmt.Errorf("unable to get hardware info: %w", err)
	}

	// fill in the hwinfo
	if len(info.SystemInfo) >= 1 {
		hwinfo.ModelName = info.SystemInfo[0]
	}

	if len(info.SystemInfo) >= 2 {
		hwinfo.FirmwareVersion = info.SystemInfo[1]
	}

	return hwinfo, nil
}

// SwitchInput changes the input on the given output to input
func (vs *AtlonaVideoSwitcher4x1) SetInputByOutput(ctx context.Context, output, input string) error {
	// atlona switchers are 1-based
	out, gerr := strconv.Atoi(output)
	if gerr != nil {
		return fmt.Errorf("unable to switch input on %s:%w", vs.Address, gerr)
	}

	in, gerr := strconv.Atoi(input)
	if gerr != nil {
		return fmt.Errorf("unable to switch input on %s:%w", vs.Address, gerr)
	}

	out++
	in++

	// validate that input/output are valid numbers
	var settings AVSettings
	err := getPage(ctx, vs.Address, avSettingsPage, &settings)
	if err != nil {
		return fmt.Errorf("unable to switch input: %w", err)
	}

	if in > len(settings.HDCPSettings) || in <= 0 {
		return fmt.Errorf("unable to switch input on %s - input %s is out of range", vs.Address, input)
	}

	if out != 1 {
		return fmt.Errorf("unable to switch input on %s - output %s is invalid", vs.Address, output)
	}

	err = sendCommand(ctx, vs.Address, fmt.Sprintf("x%vAVx%v", in, out))
	if err != nil {
		return fmt.Errorf("unable to switch input: %w", err)
	}

	return nil
}

//GetInfo .
func (vs *AtlonaVideoSwitcher4x1) GetInfo(ctx context.Context) (interface{}, error) {
	var info interface{}
	return info, fmt.Errorf("not currently implemented")
}

func (vs *AtlonaVideoSwitcher4x1) SetVolumeByBlock(ctx context.Context, block string, volume int) error {
	return fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher4x1) SetMutedByBlock(ctx context.Context, block string, muted bool) error {
	return fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher4x1) GetVolumeByBlock(ctx context.Context, block string) (int, error) {
	return 0, fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher4x1) GetMutedByBlock(ctx context.Context, block string) (bool, error) {
	return false, fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher4x1) SetLogger(logger wspool.Logger) {

}
