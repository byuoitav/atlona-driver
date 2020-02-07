package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/byuoitav/common/structs"
)

type wallPlateStruct struct {
	LoginUr   int    `json:"login_ur"`
	LoginUser string `json:"login_user"`
	Inp       int    `json:"inp"`
	Asw       int    `json:"asw"`
	Preport   int    `json:"preport"`
	Aswtime   int    `json:"aswtime"`
	HDMIAud   int    `json:"HDMIAud"`
	HDCPSet   []int  `json:"HDCPSet"`
}

func (vs *AtlonaVideoSwitcher2x1) make2x1request(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error when creting the request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return nil, fmt.Errorf("error when making call: %w", gerr)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error when making call: %w", gerr)
	}
	return body, nil
}

// GetInputByOutput .
func (vs *AtlonaVideoSwitcher2x1) GetInputByOutput(ctx context.Context, output string) (string, error) {
	var resp wallPlateStruct
	url := fmt.Sprintf("http://%s/aj.html?a=avs", vs.Address)
	body, gerr := vs.make2x1request(ctx, url)
	if gerr != nil {
		return "", fmt.Errorf("An error occured while making the call: %w", gerr)
	}
	err := json.Unmarshal([]byte(body), &resp) // here!
	if err != nil {
		return "", fmt.Errorf("error when unmarshalling the response: %w", err)
	}

	return strconv.Itoa(resp.Inp), nil

}

// SetInputByOutput .
func (vs *AtlonaVideoSwitcher2x1) SetInputByOutput(ctx context.Context, output, input string) error {
	intInput, nerr := strconv.Atoi(input)
	if nerr != nil {
		return fmt.Errorf("failed to convert input from string to int: %w", nerr)
	}
	if intInput != 1 && intInput != 2 {
		return fmt.Errorf("Invalid Input, the input you sent was %v the valid inputs are 1 or 2", intInput)
	}
	url := fmt.Sprintf("http://%s/aj.html?a=command&cmd=x%sAVx1", vs.Address, input)
	_, gerr := vs.make2x1request(ctx, url)
	if gerr != nil {
		return fmt.Errorf("An error occured while making the call: %w", gerr)
	}
	return nil
}

//GetHardwareInfo .
func (vs *AtlonaVideoSwitcher2x1) GetHardwareInfo(ctx context.Context) (structs.HardwareInfo, error) {
	var resp structs.HardwareInfo
	return resp, nil
}

//GetInfo .
func (vs *AtlonaVideoSwitcher2x1) GetInfo(ctx context.Context) (interface{}, error) {
	var info interface{}
	return info, fmt.Errorf("not currently implemented")
}

func (vs *AtlonaVideoSwitcher2x1) SetVolumeByBlock(ctx context.Context, block string, volume int) error {
	return fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher2x1) SetMutedByBlock(ctx context.Context, block string, muted bool) error {
	return fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher2x1) GetVolumeByBlock(ctx context.Context, block string) (int, error) {
	return 0, fmt.Errorf("this function is not available for this device type")
}

func (vs *AtlonaVideoSwitcher2x1) GetMutedByBlock(ctx context.Context, block string) (bool, error) {
	return false, fmt.Errorf("this function is not available for this device type")
}
