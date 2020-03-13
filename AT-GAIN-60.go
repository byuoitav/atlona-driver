package atlona

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

// Amp60 represents an Atlona 60 watt amplifier
type Amp60 struct {
	Address string
}

func getR() string {
	return fmt.Sprintf("%v", rand.Float32())
}

func getURL(address, endpoint string) string {
	return "http://" + address + "/action=" + endpoint + getR()
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
	return nil, nil
}

// GetInfo gets the current amp status
func (a *Amp60) GetInfo(ctx context.Context) (interface{}, error) {
	// open a connection with the dsp, return some info about the device...
	return nil, nil
}

// GetVolumeByBlock gets the current volume
func (a *Amp60) GetVolumeByBlock(ctx context.Context, block string) (int, error) {
	// open a connection with the dsp, return the volume for on block...
	return 0, nil
}

// GetMutedByBlock gets the current muted status
func (a *Amp60) GetMutedByBlock(ctx context.Context, block string) (bool, error) {
	// open a connection with the dsp, return the muted status for block...
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
