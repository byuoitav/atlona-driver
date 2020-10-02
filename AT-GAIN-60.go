package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/common/log"
)

// Amp60 represents an Atlona 60 watt amplifier
type Amp60 struct {
	Username string
	Password string
	Address  string
}

// AmpStatus represents the current amp status
type AmpStatus struct {
	Model         string `json:"101"`
	Firmware      string `json:"102"`
	MACAddress    string `json:"103"`
	SerialNumber  string `json:"104"`
	OperatingTime string `json:"105"`
}

// AmpAudio represents an audio response from an Atlona 60 watt amp
type AmpAudio struct {
	Volume string `json:"608,omitempty"`
	Muted  string `json:"609,omitempty"`
}

type loginResult struct {
	Login bool
}

func getR() string {
	return fmt.Sprintf("%v", rand.Float32())
}

func getURL(address, endpoint string) string {
	return "http://" + address + "/action=" + endpoint + "&r=" + getR()
}

func (a *Amp60) getLoginUrl() string {
	return "http://" + a.Address + "/action=compare&701=" + a.Username + "&702=" + a.Password + "&r=" + getR()
}

func (a *Amp60) sendReq(ctx context.Context, endpoint string) ([]byte, error) {
	// checking to validate that it is logged in
	err := a.login(ctx)
	if err != nil {
		fmt.Errorf("Login failed to device: %v", err)
	}

	var toReturn []byte
	ampUrl := getURL(a.Address, endpoint)
	Client := http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequestWithContext(ctx, "GET", ampUrl, nil)
	req.Header.Set("Context-type", "application/json")
	//req, err := http.NewRequest("GET", ampUrl, nil)
	log.L.Debug("Request Output: %v", req)
	if err != nil {
		return toReturn, fmt.Errorf("unable to make new http request: %w", err)
	}
	resp, err := Client.Do(req)
	log.L.Debug("RESP Output: %v", resp)
	if err != nil {
		if nerr, ok := err.(*url.Error); ok {
			fmt.Printf("%v\n", nerr.Err)
			if !strings.Contains(nerr.Err.Error(), "malformed") {
				return toReturn, fmt.Errorf("unable to perform request: %w", err)
			}
		} else {
			return toReturn, fmt.Errorf("unable to perform request: %w", err)
		}
		return toReturn, nil
	}
	defer resp.Body.Close()
	toReturn, err = ioutil.ReadAll(resp.Body)
	s := string(toReturn)
	log.L.Infof("Response: %v\n", s)

	if err != nil {
		return toReturn, fmt.Errorf("unable to read resp body: %w", err)
	}
	return toReturn, nil
}

// login for device
func (a *Amp60) login(ctx context.Context) error {
	// Check if we are currently logged in
	resp, err := http.Get(a.getLoginUrl())
	if err != nil {
		return fmt.Errorf("Unable to log in: %v", err)
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	s := string(out)
	if err != nil {
		fmt.Errorf("Cannot read body of test: %v", err)
	}

	if strings.Contains(s, "404") == true {
		var toReturn []byte
		loginUrl := a.getLoginUrl()
		Client := http.Client{Timeout: time.Second * 10}
		req, err := http.NewRequestWithContext(ctx, "GET", loginUrl, nil)
		if err != nil {
			return fmt.Errorf("Unable to create request: %v", err)
		}
		resp, err := Client.Do(req)
		if err != nil {
			return fmt.Errorf("Unable to connect to device: %v", err)
		}
		defer resp.Body.Close()
		toReturn, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Cannot read the body of the response")
		}
		data := loginResult{}
		json.Unmarshal(toReturn, &data)
		if data.Login != true {
			return fmt.Errorf("Not able to login: %v", err)
		}
		return nil
	}

	return nil

}

// GetInfo gets the current amp status
func (a *Amp60) GetInfo(ctx context.Context) (interface{}, error) {
	resp, err := a.sendReq(ctx, "devicestatus_get")
	if err != nil {
		return nil, fmt.Errorf("unable to get info: %w", err)
	}
	var info AmpStatus
	err = json.Unmarshal(resp, &info)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal into AmpStatus: %w", err)
	}
	return info, nil
}

// GetVolumeByBlock gets the current volume
func (a *Amp60) GetVolumes(ctx context.Context, blocks []string) (map[string]int, error) {
	resp, err := a.sendReq(ctx, "deviceaudio_get")
	if err != nil {
		return map[string]int{"": -1}, fmt.Errorf("unable to get volume: %w", err)
	}
	var info AmpAudio
	var test map[string]interface{}
	json.Unmarshal(resp, &test)
	for key, value := range test {
		log.L.Debug(key, value.(string))
	}
	log.L.Debug("Testing our json: %v", test)

	err = json.Unmarshal(resp, &info)
	if err != nil {
		return map[string]int{"": -1}, fmt.Errorf("unable to unmarshal into AmpVolume in GetVolume: %w", err)
	}
	toReturn, err := strconv.Atoi(info.Volume)
	if err != nil {
		return map[string]int{"": -1}, fmt.Errorf("Volume is empty")
	}
	return map[string]int{"": toReturn}, nil
}

// GetMutedByBlock gets the current muted status
func (a *Amp60) GetMutes(ctx context.Context, blocks []string) (map[string]bool, error) {
	resp, err := a.sendReq(ctx, "deviceaudio_get")
	if err != nil {

		return map[string]bool{"": false}, fmt.Errorf("unable to get muted: %w", err)
	}
	var info AmpAudio
	err = json.Unmarshal(resp, &info)
	if err != nil {
		return map[string]bool{"": false}, fmt.Errorf("unable to unmarshal into AmpVolume in GetMuted: %w", err)
	}
	if info.Muted == "1" {
		return map[string]bool{"": true}, nil
	}
	return map[string]bool{"": false}, nil
}

// SetVolumeByBlock sets the volume on the amp
func (a *Amp60) SetVolume(ctx context.Context, block string, volume int) error {
	_, err := a.sendReq(ctx, fmt.Sprintf("deviceaudio_set&608=%v", volume))
	if err != nil {
		return fmt.Errorf("unable to set volume: %w", err)
	}
	return nil
}

// SetMutedByBlock sets the current muted status on the amp
func (a *Amp60) SetMute(ctx context.Context, block string, muted bool) error {
	// open a connection with the dsp, set the muted status on block...
	mutedString := "0"
	if muted {
		mutedString = "1"
	}
	_, err := a.sendReq(ctx, fmt.Sprintf("deviceaudio_set&609=%v", mutedString))
	if err != nil {
		return fmt.Errorf("unable to set muted: %w", err)
	}
	return nil
}
