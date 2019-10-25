package atlona

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/byuoitav/common/structs"
)

//VideoSwitcher5x1 .
type VideoSwitcher5x1 struct {
	Username string
	Password string
}

//Login .
func login(ctx context.Context, addr string) error {
	url := fmt.Sprintf("http://%s/ajlogin.html?value=login&usn=root&pwd=Atlona", addr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Errorf("error when making call: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return fmt.Errorf("Atlona returned an ER in the response of the login request: %v", body)
	}
	defer res.Body.Close()
	return nil
}

//GetInputByOutput .
func (v *VideoSwitcher5x1) GetInputByOutput(ctx context.Context, addr, output string) (string, error) {
	loginerr := login(ctx, addr)
	if loginerr != nil {
		return "", fmt.Errorf("error logging in to Atlona to make the request: %w", loginerr)
	}
	url := fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&ro=0", addr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error when creting the request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return "", fmt.Errorf("error when making call: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return "", fmt.Errorf("Atlona returned an ER in the response of the login request: %v", body)
	}
	defer res.Body.Close()
	return splitRes[1], nil
}

//SetInputByOutput .
func (v *VideoSwitcher5x1) SetInputByOutput(ctx context.Context, addr, input string) error {
	intInput, nerr := strconv.Atoi(input)
	if nerr != nil {
		return fmt.Errorf("error occured when converting input to int: %w", nerr)
	}
	if intInput == 0 || intInput > 5 {
		return fmt.Errorf("Invalid Input. The input requested must be between 1-5. The input you requested was %v", intInput)
	}
	//decrement IntInput by 1 because the 5x1 is 0 based
	intInput = intInput - 1
	loginerr := login(ctx, addr)
	if loginerr != nil {
		return fmt.Errorf("error logging in to make a request: %w", loginerr)
	}
	url := fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&mlf=1&inp=%v", addr, input)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error when creating request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Errorf("error when sending request: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return fmt.Errorf("Atlona returned an ER in the response of the login request: %v", body)
	}
	defer res.Body.Close()
	return nil
}

//SetVolumeByBlock .
func (v *VideoSwitcher5x1) SetVolumeByBlock(ctx context.Context, addr, output string, level int) error {
	loginerr := login(ctx, addr)
	if loginerr != nil {
		return fmt.Errorf("error logging into Atlona to make request: %w", loginerr)
	}
	if level == 0 {
		level = -80
	} else {
		convertedVolume := -35 + math.Round(float64(level/2))
		level = int(convertedVolume)
	}
	url := fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Z3&mlf=1&vol=%v", addr, level)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error when creating request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Errorf("error when making request: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	fmt.Println(resp)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return fmt.Errorf("Atlona returned an error when making this request: %v", body)
	}
	defer res.Body.Close()
	return nil
}

//GetVolumeByBlock .
func (v *VideoSwitcher5x1) GetVolumeByBlock(ctx context.Context, addr, output string) (int, error) {
	loginerr := login(ctx, addr)
	if loginerr != nil {
		return 0, fmt.Errorf("error logging into Atlona make a request: %w", loginerr)
	}
	url := fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&mlf=1&inp=", addr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return 0, fmt.Errorf("error when making call: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return 0, fmt.Errorf("Atlona has returned an error to the request", "")
	}
	defer res.Body.Close()
	//convert response back to 0-100 value
	volumeLevel, gerr := strconv.Atoi(splitRes[2])
	if gerr != nil {
		return 0, fmt.Errorf("error when converting volume to int: %w", gerr)
	}
	if volumeLevel < -35 {
		return 0, nil
	} else {
		volume := ((volumeLevel + 35) * 2)
		if volume%2 != 0 {
			volume = volume + 1
		}
		return volume, nil
	}
}

//GetMutedByBlock .
func (v *VideoSwitcher5x1) GetMutedByBlock(ctx context.Context, addr, output string) (bool, error) {
	loginerr := login(ctx, addr)
	if loginerr != nil {
		return false, fmt.Errorf("Error logging into the Atlona to make the request: %w", loginerr)
	}
	url := fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&ro=0", addr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return false, fmt.Errorf("error when making call: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		fmt.Println(resp)
		return false, fmt.Errorf("Atlona returned an error when making this request: %v", body)
	}
	defer res.Body.Close()
	if splitRes[(len(splitRes)-2)] == "1" {
		return true, nil
	}
	return false, nil
}

//SetMutedByBlock .
func (v *VideoSwitcher5x1) SetMutedByBlock(ctx context.Context, addr, output string, muted bool) error {
	loginerr := login(ctx, addr)
	var url string
	if loginerr != nil {
		return fmt.Errorf("Error logging into the Atlona to make the request: %w", loginerr)
	}
	if muted {
		url = fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&mlf=1&lraud=1", addr)
	} else {
		url = fmt.Sprintf("http://%s/ajstatus.html?value=status&uid=Y1&mlf=1&lraud=0", addr)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error when making request: %w", err)
	}
	req = req.WithContext(ctx)
	res, gerr := http.DefaultClient.Do(req)
	if gerr != nil {
		return fmt.Errorf("error when making call: %w", gerr)
	}
	body, _ := ioutil.ReadAll(res.Body)
	resp := string(body)
	splitRes := strings.Split(resp, ";")
	if splitRes[0] == "ER" {
		return fmt.Errorf("Atlona returned an error when making this request: %v", body)
	}
	defer res.Body.Close()
	return nil
}

//GetHardwareInfo .
func (v *VideoSwitcher5x1) GetHardwareInfo(ctx context.Context, addr string) (structs.HardwareInfo, error) {
	var resp structs.HardwareInfo
	return resp, nil
}
