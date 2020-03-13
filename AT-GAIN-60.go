package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

// Amp60 represents an Atlona 60 watt amplifier
type Amp60 struct {
	Address string
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
	Volume int `json:"608,omitempty"`
	Muted  int `json:"609,omitempty"`
}

func getR() string {
	return fmt.Sprintf("%v", rand.Float32())
}

func getURL(address, endpoint string) string {
	return "http://" + address + "/action=" + endpoint + "&" + getR()
}

func (a *Amp60) sendReq(ctx context.Context, endpoint string) ([]byte, error) {
	var toReturn []byte
	url := getURL(a.Address, endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return toReturn, fmt.Errorf("unable to make new http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return toReturn, fmt.Errorf("unable to perform request: %w", err)
	}
	defer resp.Body.Close()
	toReturn, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return toReturn, fmt.Errorf("unable to read resp body: %w", err)
	}
	return toReturn, nil
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
func (a *Amp60) GetVolumeByBlock(ctx context.Context, block string) (int, error) {
	resp, err := a.sendReq(ctx, "deviceaudio_get")
	if err != nil {
		return -1, fmt.Errorf("unable to get volume: %w", err)
	}
	var info AmpAudio
	err = json.Unmarshal(resp, &info)
	if err != nil {
		return -1, fmt.Errorf("unable to unmarshal into AmpVolume: %w", err)
	}
	return info.Volume, nil
}

// GetMutedByBlock gets the current muted status
func (a *Amp60) GetMutedByBlock(ctx context.Context, block string) (bool, error) {
	resp, err := a.sendReq(ctx, "deviceaudio_get")
	if err != nil {
		return false, fmt.Errorf("unable to get volume: %w", err)
	}
	var info AmpAudio
	err = json.Unmarshal(resp, &info)
	if err != nil {
		return false, fmt.Errorf("unable to unmarshal into AmpVolume: %w", err)
	}
	if info.Muted == 1 {
		return true, nil
	}
	return false, nil
}

// SetVolumeByBlock sets the volume on the amp
func (a *Amp60) SetVolumeByBlock(ctx context.Context, block string, volume int) error {
	// open a connection with the dsp, set the volume on block...
	return nil
}

// SetMutedByBlock sets the current muted status on the amp
func (a *Amp60) SetMutedByBlock(ctx context.Context, block string, muted bool) error {
	// open a connection with the dsp, set the muted status on block...
	return nil
}