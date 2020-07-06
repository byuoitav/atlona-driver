package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/wspool"
	"github.com/gorilla/websocket"
)

type AtlonaVideoSwitcher5x1 struct {
	Username string
	Password string
	Address  string
	once     sync.Once
	pool     wspool.Pool
	Logger   wspool.Logger
}

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

func (vs *AtlonaVideoSwitcher5x1) createPool() {
	if vs.Logger != nil {
		vs.Logger.Infof("creating pool")
	}

	vs.pool = wspool.Pool{
		NewConnection: createConnectionFunc(vs.Address),
		TTL:           10 * time.Second,
		Delay:         75 * time.Millisecond,
		Logger:        vs.Logger,
	}

}

func createConnectionFunc(address string) wspool.NewConnectionFunc {
	return func(ctx context.Context) (*websocket.Conn, error) {
		dialer := &websocket.Dialer{}
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		ws, _, err := dialer.DialContext(ctx, fmt.Sprintf("ws://%s:543", address), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to open websocket: %s", err.Error())
		}
		return ws, nil
	}
}

//GetAudioVideoInputs .
func (vs *AtlonaVideoSwitcher5x1) GetAudioVideoInputs(ctx context.Context) (map[string]string, error) {
	toReturn := make(map[string]string)
	vs.once.Do(vs.createPool)

	var roomInfo room
	var bytes []byte

	err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

		vs.pool.Logger.Infof("writing message to Get Input")

		err := ws.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to write message: %s", err.Error())
		}

		if vs.Logger != nil {
			vs.Logger.Infof("reading message from websocket")
		}

		timeout := time.Now()
		timeout = timeout.Add(time.Second * 5)

		err = ws.SetReadDeadline(timeout)
		if err != nil {
			return fmt.Errorf("failed to set readDeadline: %s", err)
		}

		_, bytes, err = ws.ReadMessage()
		if err != nil {
			vs.Logger.Errorf("failed reading message from websocket: %s", err)
			return fmt.Errorf("failed to read message: %s", err)
		}

		vs.pool.Logger.Infof("read message from Get Input")

		return nil
	})

	if err != nil {
		return toReturn, fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	err = json.Unmarshal(bytes, &roomInfo)

	if err != nil {
		return toReturn, fmt.Errorf("failed to unmarshal message: %s", err.Error())
	}

	toReturn[""] = roomInfo.Result.AVSettings.Source[6:]
	return toReturn, nil
}

//SetAudioVideoInput .
func (vs *AtlonaVideoSwitcher5x1) SetAudioVideoInput(ctx context.Context, output, input string) error {
	vs.once.Do(vs.createPool)

	intInput, nerr := strconv.Atoi(input)

	if nerr != nil {
		return fmt.Errorf("error occured when converting input to int: %w", nerr)
	}

	if intInput == 0 || intInput > 5 {
		return fmt.Errorf("Invalid Input. The input requested must be between 1-5. The input you requested was %v", intInput)
	}

	err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

		if vs.Logger != nil {
			vs.Logger.Infof("writing message")
		}

		vs.pool.Logger.Infof("writing message to Set Input")

		err := ws.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to write message: %s", err.Error())
		}

		vs.pool.Logger.Infof("successful wrote message to set Input")

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	if err != nil {
		return fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	return nil
}

//SetVolume .
func (vs *AtlonaVideoSwitcher5x1) SetVolume(ctx context.Context, output string, level int) error {
	vs.once.Do(vs.createPool)

	if level == 0 {
		level = -80
	} else {
		convertedVolume := -35 + math.Round(float64(level/2))
		level = int(convertedVolume)
	}

	err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

		vs.pool.Logger.Infof("writing message to Set Volume")

		err := ws.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to write message: %s", err.Error())
		}

		vs.pool.Logger.Infof("successfully wrote to Set Volume")

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	if err != nil {
		return fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	return nil
}

//GetVolumes .
func (vs *AtlonaVideoSwitcher5x1) GetVolumes(ctx context.Context, blocks []string) (map[string]int, error) {
	toReturn := make(map[string]int)

	vs.once.Do(vs.createPool)

	var roomInfo room
	var bytes []byte

	err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

		vs.pool.Logger.Infof("writing message to Get Volume")

		err := ws.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to write message: %s", err.Error())
		}

		timeout := time.Now()
		timeout = timeout.Add(time.Second * 5)

		err = ws.SetReadDeadline(timeout)
		if err != nil {
			return fmt.Errorf("failed to set readDeadline: %s", err)
		}

		_, bytes, err = ws.ReadMessage()
		if err != nil {
			vs.Logger.Errorf("failed reading message from websocket: %s", err)
			return fmt.Errorf("failed to read message: %s", err)
		}

		vs.pool.Logger.Infof("read message from Get volume")

		return nil
	})
	if err != nil {
		return toReturn, fmt.Errorf("failed to read message from channel: %s", err.Error())
	}

	err = json.Unmarshal(bytes, &roomInfo)

	if err != nil {
		return toReturn, fmt.Errorf("failed to unmarshal response: %s", err.Error())
	}

	volumeLevel, err := strconv.Atoi(roomInfo.Result.AVSettings.Volume)
	if err != nil {
		return toReturn, fmt.Errorf("failed to convert volume to int: %s", err.Error())
	}

	if volumeLevel < -35 {
		toReturn[""] = 0
	} else {
		volume := ((volumeLevel + 35) * 2)
		if volume%2 != 0 {
			volume = volume + 1
		}
		toReturn[""] = volume
	}

	return toReturn, nil
}

//GetMutes .
func (vs *AtlonaVideoSwitcher5x1) GetMutes(ctx context.Context, blocks []string) (map[string]bool, error) {
	toReturn := make(map[string]bool)

	for _, block := range blocks {
		vs.once.Do(vs.createPool)

		var roomInfo room
		var bytes []byte

		err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

			vs.pool.Logger.Infof("writing message to Get Muted")

			err := ws.WriteMessage(websocket.TextMessage, []byte(body))
			if err != nil {
				return fmt.Errorf("failed to write message: %s", err.Error())
			}

			timeout := time.Now()
			timeout = timeout.Add(time.Second * 5)

			err = ws.SetReadDeadline(timeout)
			if err != nil {
				return fmt.Errorf("failed to set readDeadline: %s", err)
			}

			_, bytes, err = ws.ReadMessage()
			if err != nil {
				vs.Logger.Errorf("failed reading message from websocket: %s", err)
				return fmt.Errorf("failed to read message: %s", err)
			}

			vs.pool.Logger.Infof("read message from Get Muted")

			return nil
		})
		if err != nil {
			return toReturn, fmt.Errorf("failed to read message from channel: %s", err.Error())
		}

		err = json.Unmarshal(bytes, &roomInfo)

		if err != nil {
			return toReturn, fmt.Errorf("failed to unmarshal response: %s", err.Error())
		}

		switch block {
		case "HDMI":
			isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.HDMIAudioMute))
			if err != nil {
				return toReturn, fmt.Errorf("failed to parse bool: %s", err.Error())
			}
			toReturn[block] = isMuted
		case "HDBT":
			isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.HDBTAudioMute))
			if err != nil {
				return toReturn, fmt.Errorf("failed to parse bool: %s", err.Error())
			}
			toReturn[block] = isMuted
		default:
			// Analog
			isMuted, err := strconv.ParseBool(fmt.Sprintf("%v", roomInfo.Result.AVSettings.AnalogAudioMute))
			if err != nil {
				return toReturn, fmt.Errorf("failed to parse bool: %s", err.Error())
			}
			toReturn[block] = isMuted
		}
	}

	return toReturn, nil
}

//SetMute .
func (vs *AtlonaVideoSwitcher5x1) SetMute(ctx context.Context, output string, muted bool) error {
	vs.once.Do(vs.createPool)

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

	err := vs.pool.Do(ctx, func(ws *websocket.Conn) error {
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

		vs.pool.Logger.Infof("writing message to set Muted")

		err := ws.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			return fmt.Errorf("failed to write message: %s", err.Error())
		}

		vs.pool.Logger.Infof("wrote message to set Muted")

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to read message from channel: %s", err.Error())
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
