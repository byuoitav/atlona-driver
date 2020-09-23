package atlona

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	_omePs62Endpoint = "/cgi-bin/config.cgi"
)

type AtOmePs62 struct {
	Username string
	Password string
	Address  string
}

type config struct {
	Video videoConfig `json:"video"`
	Audio audioConfig `json:"audio"`
}

type videoConfig struct {
	VidOut struct {
		HdmiOut struct {
			Mirror struct {
				Status   bool `json:"status"`
				VideoSrc int  `json:"videoSrc"`
			} `json:"mirror"`

			HdmiOutA struct {
				VideoSrc int `json:"videoSrc"`
			} `json:"hdmiOutA"`

			HdmiOutB struct {
				VideoSrc int `json:"videoSrc"`
			} `json:"hdmiOutB"`
		} `json:"hdmiOut"`
	} `json:"vidOut"`
}

type audioConfig struct {
	AudOut struct {
		ZoneOut1 struct {
			AudioSource string `json:"audioSource"`
			AudioVol    int    `json:"audioVol"`

			AnalogOut struct {
				AudioMute bool `json:"audioMute"`
			} `json:"analogOut"`
		} `json:"zoneOut1"`

		ZoneOut2 struct {
			AudioSource string `json:"audioSource"`
			AudioVol    int    `json:"audioVol"`

			AnalogOut struct {
				AudioMute bool `json:"audioMute"`
			} `json:"analogOut"`
		} `json:"zoneOut2"`
	} `json:"audOut"`
}

func (vs *AtOmePs62) getConfig(ctx context.Context, body string) (config, error) {
	var config config

	url := "http://" + vs.Address + _omePs62Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return config, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(vs.Username, vs.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return config, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return config, fmt.Errorf("unable to decode response: %w", err)
	}

	return config, nil
}

func (vs *AtOmePs62) setConfig(ctx context.Context, body string) error {
	url := "http://" + vs.Address + _omePs62Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(vs.Username, vs.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	var res struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}

	if !strings.EqualFold(res.Message, "OK") {
		return fmt.Errorf("bad response (%d): %s", res.Status, res.Message)
	}

	return nil
}

// GetAudioVideoInputs .
func (vs *AtOmePs62) GetAudioVideoInputs(ctx context.Context) (map[string]string, error) {
	body := `{ "getConfig": { "video": { "vidOut": { "hdmiOut": {}}}}}`

	config, err := vs.getConfig(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("unable to get config: %w", err)
	}

	inputs := make(map[string]string)
	if config.Video.VidOut.HdmiOut.Mirror.Status {
		inputs["mirror"] = strconv.Itoa(config.Video.VidOut.HdmiOut.Mirror.VideoSrc)
	} else {
		inputs["hdmiOutA"] = strconv.Itoa(config.Video.VidOut.HdmiOut.HdmiOutA.VideoSrc)
		inputs["hdmiOutB"] = strconv.Itoa(config.Video.VidOut.HdmiOut.HdmiOutB.VideoSrc)
	}

	return inputs, nil
}

// SetAudioVideoInput .
func (vs *AtOmePs62) SetAudioVideoInput(ctx context.Context, output, input string) error {
	in, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("input must be an int: %w", err)
	}

	body := fmt.Sprintf(`{ "setConfig": { "video": { "vidOut": { "hdmiOut": { "%s": { "videoSrc": %v }}}}}}`, output, in)
	if err := vs.setConfig(ctx, body); err != nil {
		return fmt.Errorf("unable to set config: %w", err)
	}

	return nil
}

// GetVolumes .
func (vs *AtOmePs62) GetVolumes(ctx context.Context, blocks []string) (map[string]int, error) {
	body := `{ "getConfig": "audio": { "audOut": {}}}`

	config, err := vs.getConfig(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("unable to get config: %w", err)
	}

	// always return all of the blocks, regardless of `blocks`
	// (since we don't have to do any extra work)
	vols := make(map[string]int)

	// zoneOut1 volume
	if config.Audio.AudOut.ZoneOut1.AudioVol < -50 {
		vols["zoneOut1"] = 0
	} else {
		vols["zoneOut1"] = 2 * (config.Audio.AudOut.ZoneOut1.AudioVol + 50)
	}

	// zoneOut2 volume
	if config.Audio.AudOut.ZoneOut2.AudioVol < -50 {
		vols["zoneOut2"] = 0
	} else {
		vols["zoneOut2"] = 2 * (config.Audio.AudOut.ZoneOut2.AudioVol + 50)
	}

	return vols, nil
}

// SetVolume .
func (vs *AtOmePs62) SetVolume(ctx context.Context, output string, level int) error {
	if output != "zoneOut1" && output != "zoneOut2" {
		return errors.New("invalid output")
	}

	// Atlona volume levels are from -90 to 10 and the number we receive is 0-100
	// If volume level is supposed to be zero set it -90 on atlona
	if level == 0 {
		level = -90
	} else {
		convertedVolume := -50 + math.Round(float64(level/2))
		level = int(convertedVolume)
	}

	body := fmt.Sprintf(`{ "setConfig": { "audio": { "audOut": { "%s": { "audioVol": %d }}}}}`, output, level)
	if err := vs.setConfig(ctx, body); err != nil {
		return fmt.Errorf("unable to set config: %w", err)
	}

	return nil
}

// GetMutes .
func (vs *AtOmePs62) GetMutes(ctx context.Context, blocks []string) (map[string]bool, error) {
	body := `{ "getConfig": "audio": { "audOut": {}}}`

	config, err := vs.getConfig(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("unable to get config: %w", err)
	}

	// always return all of the blocks, regardless of `blocks`
	// (since we don't have to do any extra work)
	mutes := make(map[string]bool)
	mutes["zoneOut1"] = config.Audio.AudOut.ZoneOut1.AnalogOut.AudioMute
	mutes["zoneOut2"] = config.Audio.AudOut.ZoneOut2.AnalogOut.AudioMute

	return mutes, nil
}

// SetMute .
func (vs *AtOmePs62) SetMute(ctx context.Context, output string, muted bool) error {
	if output != "zoneOut1" && output != "zoneOut2" {
		return errors.New("invalid output")
	}

	body := fmt.Sprintf(`{ "setConfig": { "audio": { "audOut": { "%s": { "analogOut": { "audioMute": %t }}}}}}`, output, muted)
	if err := vs.setConfig(ctx, body); err != nil {
		return fmt.Errorf("unable to set config: %w", err)
	}

	return nil
}

/*
//GetHardwareInfo .
func (vs *AtlonaVideoSwitcher6x2) GetHardwareInfo(ctx context.Context) (structs.HardwareInfo, error) {
	var network atlonaNetwork
	var hardware atlonaHardwareInfo
	var resp structs.HardwareInfo
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)

	//Get network info
	requestBody := fmt.Sprintf(`
	{
		"getConfig": {
			"network": {
				"eth0":{
				}
			}
		}
	}`)
	body, gerr := vs.make6x2request(ctx, url, requestBody)
	if gerr != nil {
		return structs.HardwareInfo{}, fmt.Errorf("An error occured while making the call: %w", gerr)
	}

	err := json.Unmarshal([]byte(body), &network)

	if err != nil {
		return resp, fmt.Errorf("error when unmarshalling the response: %w", err)
	}

	//Get other hardware info
	requestBody = fmt.Sprintf(`
	{
		"getConfig": {
			"system": {}
		}
	}`)
	body, gerr = vs.make6x2request(ctx, url, requestBody)
	if gerr != nil {
		return structs.HardwareInfo{}, fmt.Errorf("An error occured while making the call: %w", gerr)
	}
	err = json.Unmarshal([]byte(body), &hardware)
	if err != nil {
		return resp, fmt.Errorf("error when unmarshalling the response: %w", err)
	}

	//Load up the hardware struct
	resp.Hostname = hardware.System.Model
	resp.ModelName = hardware.System.Model
	resp.NetworkInfo.MACAddress = network.Network.Eth0.MacAddr
	resp.NetworkInfo.IPAddress = network.Network.Eth0.IPSettings.Ipaddr
	resp.NetworkInfo.Gateway = network.Network.Eth0.IPSettings.Gateway
	resp.PowerStatus = hardware.System.PowerStatus
	return resp, nil
}
*/

// GetInfo .
func (vs *AtOmePs62) GetInfo(ctx context.Context) (interface{}, error) {
	return nil, fmt.Errorf("not currently implemented")
}

/*
type atlonaNetwork struct {
	Network struct {
		Eth0 struct {
			MacAddr    string `json:"macAddr"`
			DomainName string `json:"domainName"`
			DNSServer1 string `json:"dnsServer1"`
			DNSServer2 string `json:"dnsServer2"`
			IPSettings struct {
				TelnetPort int    `json:"telnetPort"`
				Ipaddr     string `json:"ipaddr"`
				Netmask    string `json:"netmask"`
				Gateway    string `json:"gateway"`
			} `json:"ipSettings"`
			LastIpaddr string `json:"lastIpaddr"`
			BootProto  string `json:"bootProto"`
		} `json:"eth0"`
	} `json:"network"`
}

type atlonaHardwareInfo struct {
	System struct {
		PowerStatus     string `json:"powerStatus"`
		VendorID        string `json:"vendorID"`
		Model           string `json:"model"`
		SerialNumber    string `json:"serialNumber"`
		FirmwareVersion struct {
			Package          string `json:"package"`
			MasterMCU        string `json:"masterMCU"`
			TransceiverChipB string `json:"transceiverChip_B"`
			TransceiverChipC string `json:"transceiverChip_C"`
			TransceiverChipE string `json:"transceiverChip_E"`
			TransceiverChipF string `json:"transceiverChip_F"`
			Audio            string `json:"audio"`
			Fpga             string `json:"fpga"`
			Usb              string `json:"usb"`
			ScalerChip       string `json:"scalerChip"`
			ValensA          string `json:"valens_A"`
			ValensB          string `json:"valens_B"`
			ValensC          string `json:"valens_C"`
			SlaveMCU         string `json:"slaveMCU"`
			TransceiverChipA string `json:"transceiverChip_A"`
		} `json:"firmwareVersion"`
	} `json:"system"`
}
*/
