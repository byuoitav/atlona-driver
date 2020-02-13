package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/byuoitav/common/structs"
)

type AtlonaVideoSwitcher6x2 struct {
	Username string
	Password string
	Address  string
}

type atlonaVideo struct {
	Video struct {
		VidOut struct {
			HdmiOut struct {
				HdmiOutA struct {
					VideoSrc int `json:"videoSrc"`
				} `json:"hdmiOutA"`
				HdmiOutB struct {
					VideoSrc int `json:"videoSrc"`
				} `json:"hdmiOutB"`
			} `json:"hdmiOut"`
		} `json:"vidOut"`
	} `json:"video"`
}

type atlonaAudio struct {
	Audio struct {
		AudOut struct {
			ZoneOut1 struct {
				AnalogOut struct {
					AudioMute  bool `json:"audioMute"`
					AudioDelay int  `json:"audioDelay"`
				} `json:"analogOut"`
				AudioVol int `json:"audioVol"`
			} `json:"zoneOut1"`
			ZoneOut2 struct {
				AnalogOut struct {
					AudioMute  bool `json:"audioMute"`
					AudioDelay int  `json:"audioDelay"`
				} `json:"analogOut"`
				AudioVol int `json:"audioVol"`
			} `json:"zoneOut2"`
		} `json:"audOut"`
	} `json:"audio"`
}

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

//Atlona6x2HardwareInfo .
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

func (vs *AtlonaVideoSwitcher6x2) make6x2request(ctx context.Context, url, requestBody string) ([]byte, error) {
	payload := strings.NewReader(requestBody)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("error when creting the request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	//This needs to be replaced with an environmental variable
	req.Header.Add("Authorization", "Basic YWRtaW46QXRsb25h")
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

//GetInputByOutput .
func (vs *AtlonaVideoSwitcher6x2) GetInputByOutput(ctx context.Context, output string) (string, error) {
	var resp atlonaVideo
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)

	requestBody := fmt.Sprintf(`
	{
		"getConfig": {
			"video": {
				"vidOut": {
					"hdmiOut": {
					}
				}
			}
		}
	}`)
	body, gerr := vs.make6x2request(ctx, url, requestBody)
	if gerr != nil {
		return "", fmt.Errorf("An error occured while making the call: %w", gerr)
	}
	err := json.Unmarshal([]byte(body), &resp)
	if err != nil {
		fmt.Printf("%s/n", body)
		return "", fmt.Errorf("error when unmarshalling the response: %w", err)
	}
	//Get the inputsrc for the requested output
	input := ""
	if output == "1" {
		input = strconv.Itoa(resp.Video.VidOut.HdmiOut.HdmiOutA.VideoSrc)
	} else if output == "2" {
		input = strconv.Itoa(resp.Video.VidOut.HdmiOut.HdmiOutB.VideoSrc)
	} else {
		return input, fmt.Errorf("Invalid Output. Valid Output names are 1 and 2 you gave us %s", output)
	}
	return input, nil
}

//SetInputByOutput .
func (vs *AtlonaVideoSwitcher6x2) SetInputByOutput(ctx context.Context, output, input string) error {
	in, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("error when making call: %w", err)
	}
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)
	requestBody := ""
	if output == "1" {
		requestBody = fmt.Sprintf(`
		{
			"setConfig":{
				"video":{
					"vidOut":{
						"hdmiOut":{
							"hdmiOutA":{
								"videoSrc":%v
							}
						}
					}
				}
			}
		}`, in)
	} else if output == "2" {
		requestBody = fmt.Sprintf(`
		{
			"setConfig":{
				"video":{
					"vidOut":{
						"hdmiOut":{
							"hdmiOutB":{
								"videoSrc":%v
							}
						}
					}
				}
			}
		}`, in)
	} else {
		return fmt.Errorf("Invalid Output. Valid Output names are 1 and 2")
	}
	_, gerr := vs.make6x2request(ctx, url, requestBody)
	if gerr != nil {
		return fmt.Errorf("An error occured while making the call: %w", gerr)
	}
	return nil
}

//SetVolumeByBlock .
func (vs *AtlonaVideoSwitcher6x2) SetVolumeByBlock(ctx context.Context, output string, level int) error {
	//Atlona volume levels are from -90 to 10 and the number we recieve is 0-100
	//if volume level is supposed to be zero set it to zero (which is -90) on atlona

	if level == 0 {
		level = -90
	} else {
		convertedVolume := -40 + math.Round(float64(level/2))
		level = int(convertedVolume)
	}
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)
	if output == "1" || output == "2" {
		requestBody := fmt.Sprintf(`
		{
			"setConfig": {
				"audio": {
					"audOut": {
						"zoneOut%s": {
							"audioVol": %d
						}
					}
				}
			}
		}`, output, level)
		_, gerr := vs.make6x2request(ctx, url, requestBody)
		if gerr != nil {
			return fmt.Errorf("An error occured while making the call: %w", gerr)
		}
	} else {
		return fmt.Errorf("Invalid Output. Valid Audio Output names are Audio1 and Audio2: you gave us %s", output)
	}
	return nil
}

//GetVolumeByBlock .
func (vs *AtlonaVideoSwitcher6x2) GetVolumeByBlock(ctx context.Context, output string) (int, error) {
	var resp atlonaAudio
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)
	requestBody := fmt.Sprintf(`
	{
		"getConfig": {
			"audio": {
				"audOut": {
					}
				}
			}
	}`)
	body, gerr := vs.make6x2request(ctx, url, requestBody)
	if gerr != nil {
		return 0, fmt.Errorf("An error occured while making the call: %w", gerr)
	}

	err := json.Unmarshal([]byte(body), &resp) // here!
	if err != nil {
		return 0, fmt.Errorf("error when unmarshalling the response: %w", err)
	}
	if output == "1" {
		if resp.Audio.AudOut.ZoneOut1.AudioVol < -40 {
			return 0, nil
		} else {
			volume := ((resp.Audio.AudOut.ZoneOut1.AudioVol + 40) * 2)
			return volume, nil
		}

	} else if output == "2" {
		return resp.Audio.AudOut.ZoneOut2.AudioVol + 90, nil
	} else {
		return 0, fmt.Errorf("Invalid Output. Valid Output names are 1 and 2 you gave us %s", output)
	}
}

//GetMutedByBlock .
func (vs *AtlonaVideoSwitcher6x2) GetMutedByBlock(ctx context.Context, output string) (bool, error) {
	var resp atlonaAudio
	if output == "1" || output == "2" {
		url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)
		requestBody := fmt.Sprintf(`
		{
			"getConfig": {
				"audio":{
					"audOut":{
						"zoneOut%s":{
							"analogOut": {				
							}
						}
					}
				}	
			}	
		}`, output)
		body, gerr := vs.make6x2request(ctx, url, requestBody)
		if gerr != nil {
			return false, fmt.Errorf("An error occured while making the call: %w", gerr)
		}
		err := json.Unmarshal([]byte(body), &resp)
		if err != nil {
			return false, fmt.Errorf("error when unmarshalling the response: %w", err)
		}
	} else {
		return false, fmt.Errorf("Invalid Output. Valid Output names are 1 and 2 you gave us %s", output)
	}
	if output == "1" {
		return resp.Audio.AudOut.ZoneOut1.AnalogOut.AudioMute, nil
	} else if output == "2" {
		return resp.Audio.AudOut.ZoneOut2.AnalogOut.AudioMute, nil
	} else {
		return false, fmt.Errorf("Invalid Output. Valid Output names are 1 and 2 you gave us %s", output)
	}
}

//SetMutedByBlock .
func (vs *AtlonaVideoSwitcher6x2) SetMutedByBlock(ctx context.Context, output string, muted bool) error {
	url := fmt.Sprintf("http://%s/cgi-bin/config.cgi", vs.Address)
	if output == "1" || output == "2" {
		requestBody := fmt.Sprintf(`
		{
			"setConfig": {
				"audio": {
					"audOut": {
						"zoneOut%s": {
							"analogOut": {
								"audioMute": %v
							}
						}
					}
				}
			}
		}`, output, muted)
		_, gerr := vs.make6x2request(ctx, url, requestBody)
		if gerr != nil {
			return fmt.Errorf("An error occured while making the call: %w", gerr)
		}
	} else {
		return fmt.Errorf("Invalid Output. Valid Output names are Audio1 and Audio2 you gave us %s", output)
	}
	return nil
}

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

//GetInfo .
func (vs *AtlonaVideoSwitcher6x2) GetInfo(ctx context.Context) (interface{}, error) {
	var info interface{}
	return info, fmt.Errorf("not currently implemented")
}
