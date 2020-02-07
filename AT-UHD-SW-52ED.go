package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/byuoitav/common/structs"
	"github.com/gorilla/websocket"
)

type room struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		AVSettings struct {
			Source          string `json:"source"`
			Autoswitch      int    `json:"Autoswitch"`
			Volume          string `json:"Volume"`
			HDMIAudioMute   int    `json:"HDMI Audio Mute"`
			HDBTAudioMute   int    `json:"HDBT Audio Mute"`
			AnalogAudioMute int    `json:"Analog Audio Mute"`
		} `json:"AV Settings"`
	} `json:"result"`
}

//openWebsocket .
func (vs *AtlonaVideoSwitcher5x1) openWebsocket(c context.Context) error {
	dialer := &websocket.Dialer{}
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	ws, _, err := dialer.DialContext(ctx, fmt.Sprintf("ws://%s:543", vs.Address), nil)
	if err != nil {
		return fmt.Errorf("failed to open websocket: %s", err.Error())
	}
	vs.ws = ws
	return nil
}

//closeWebsocket .
func (vs *AtlonaVideoSwitcher5x1) closeWebsocket(c context.Context) error {
	err := vs.ws.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		return fmt.Errorf("failed to close websocket: %s", err.Error())
	}

	err = vs.ws.Close()
	if err != nil {
		return fmt.Errorf("failed to close websocket: %s", err.Error())
	}

	return nil
}

//GetInputByOutput .
func (vs *AtlonaVideoSwitcher5x1) GetInputByOutput(ctx context.Context, output string) (string, error) {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)
	var roomInfo room
	body := `{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_get",
		"params": {
			"sections": [
				"AV Settings"
			]
		}
	}`

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return "", fmt.Errorf("failed to write message: %s", err.Error())
	}

	_, bytes, err := vs.ws.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("failed to read message: %s", err.Error())
	}

	err = json.Unmarshal(bytes, &roomInfo)

	if err != nil {
		return "", fmt.Errorf("failed to unmarshal message: %s", err.Error())
	}

	return roomInfo.Result.AVSettings.Source[6:], nil
}

//SetInputByOutput .
func (vs *AtlonaVideoSwitcher5x1) SetInputByOutput(ctx context.Context, output, input string) error {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)
	intInput, nerr := strconv.Atoi(input)

	if nerr != nil {
		return fmt.Errorf("error occured when converting input to int: %w", nerr)
	}

	if intInput == 0 || intInput > 5 {
		return fmt.Errorf("Invalid Input. The input requested must be between 1-5. The input you requested was %v", intInput)
	}

	body := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_set",
		"params": {
		  "AV Settings": {
			"source": "input %s"
		  }
		}
	  }`, input)

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to write message: %s", err.Error())
	}

	return nil
}

//SetVolumeByBlock .
func (vs *AtlonaVideoSwitcher5x1) SetVolumeByBlock(ctx context.Context, output string, level int) error {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)

	if level == 0 {
		level = -80
	} else {
		convertedVolume := -35 + math.Round(float64(level/2))
		level = int(convertedVolume)
	}

	body := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_set",
		"params": {
		  "AV Settings": {
			"Volume": "%v"
		  }
		}
	  }`, level)

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to write message: %s", err.Error())

	}
	return nil
}

//GetVolumeByBlock .
func (vs *AtlonaVideoSwitcher5x1) GetVolumeByBlock(ctx context.Context, output string) (int, error) {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)

	var roomInfo room
	body := `{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_get",
		"params": {
			"sections": [
				"AV Settings"
			]
		}
	}`

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return 0, fmt.Errorf("failed to write message: %s", err.Error())
	}

	_, bytes, err := vs.ws.ReadMessage()
	if err != nil {
		return 0, fmt.Errorf("failed to read message: %s", err.Error())
	}

	err = json.Unmarshal(bytes, &roomInfo)

	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %s", err.Error())
	}

	volumeLevel, err := strconv.Atoi(roomInfo.Result.AVSettings.Volume)
	if err != nil {
		return 0, fmt.Errorf("failed to convert volume to int: %s", err.Error())
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
func (vs *AtlonaVideoSwitcher5x1) GetMutedByBlock(ctx context.Context, output string) (bool, error) {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)

	var roomInfo room
	body := `{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_get",
		"params": {
			"sections": [
				"AV Settings"
			]
		}
	}`

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return false, fmt.Errorf("failed to write message: %s", err.Error())
	}

	_, bytes, err := vs.ws.ReadMessage()
	if err != nil {
		return false, fmt.Errorf("failed to read message: %s", err.Error())
	}

	err = json.Unmarshal(bytes, &roomInfo)

	if err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %s", err.Error())
	}

	switch output {
	case "HDMI":
		isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.HDMIAudioMute))
		if err != nil {
			return false, fmt.Errorf("failed to parse bool: %s", err.Error())
		}
		return isMuted, nil
	case "HDBT":
		isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.HDBTAudioMute))
		if err != nil {
			return false, fmt.Errorf("failed to parse bool: %s", err.Error())
		}
		return isMuted, nil
	default:
		// Analog
		isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.AnalogAudioMute))
		if err != nil {
			return false, fmt.Errorf("failed to parse bool: %s", err.Error())
		}
		return isMuted, nil
	}
}

//SetMutedByBlock .
func (vs *AtlonaVideoSwitcher5x1) SetMutedByBlock(ctx context.Context, output string, muted bool) error {
	vs.openWebsocket(ctx)
	defer vs.closeWebsocket(ctx)

	var audioBlock string
	muteInt := 0

	if muted {
		muteInt = 1
	}

	switch output {
	case "HDMI":
		audioBlock = fmt.Sprintf(`"HDMI Audio Mute": %v`, muteInt)
	case "HDBT":
		audioBlock = fmt.Sprintf(`"HDBT Audio Mute": %v`, muteInt)
	default:
		// Analog
		audioBlock = fmt.Sprintf(`"Analog Audio Mute": %v`, muteInt)
	}

	body := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": "<configuration_id>",
		"method": "config_set",
		"params": {
		  "AV Settings": {
			%s
		  }
		}
	  }`, audioBlock)

	err := vs.ws.WriteMessage(websocket.TextMessage, []byte(body))
	if err != nil {
		return fmt.Errorf("failed to write message: %s", err.Error())
	}

	return nil

}

//GetHardwareInfo .
func (vs *AtlonaVideoSwitcher5x1) GetHardwareInfo(ctx context.Context) (structs.HardwareInfo, error) {
	var resp structs.HardwareInfo
	return resp, fmt.Errorf("not currently implemented")
}

//GetInfo .
func (vs *AtlonaVideoSwitcher5x1) GetInfo(ctx context.Context) (interface{}, error) {
	var info interface{}
	return info, fmt.Errorf("not currently implemented")
}
