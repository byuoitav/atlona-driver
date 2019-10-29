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

// GetInputByOutput .
func (vs *AtlonaVideoSwitcher) getInputByOutput2x1(ctx context.Context, output string) (string, error) {
	var resp wallPlateStruct
	url := fmt.Sprintf("http://%s/aj.html?a=avs", vs.Address)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error when making call: %w", err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(body), &resp) // here!
	if err != nil {
		return "", fmt.Errorf("error when unmarshalling the response: %w", err)
	}
	defer res.Body.Close()

	return strconv.Itoa(resp.Inp), nil

}

// SetInputByOutput .
func (vs *AtlonaVideoSwitcher) setInputByOutput2x1(ctx context.Context, output, input string) error {
	intInput, nerr := strconv.Atoi(input)
	if nerr != nil {
		return fmt.Errorf("failed to convert input from string to int: %w", nerr)
	}
	if intInput != 1 && intInput != 2 {
		return fmt.Errorf("Invalid Input, the input you sent was %v the valid inputs are 1 or 2", intInput)
	}
	url := fmt.Sprintf("http://%s/aj.html?a=command&cmd=x%sAVx1", vs.Address, input)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Errorf("error when making call: %w", gerr)
	}
	defer res.Body.Close()
	return nil
}

//GetHardwareInfo .
func (vs *AtlonaVideoSwitcher) getHardwareInfo2x1(ctx context.Context) (structs.HardwareInfo, error) {
	var resp structs.HardwareInfo
	return resp, nil
}

//GetInfo .
func (vs *AtlonaVideoSwitcher) getInfo2x1(ctx context.Context) (interface{}, error) {
	var info interface{}
	return info, fmt.Errorf("not currently implemented")
}
