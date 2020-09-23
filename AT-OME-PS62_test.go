package atlona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetAudioVideoInputsMirrored(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var config config
		config.Video.VidOut.HdmiOut.Mirror.Status = true
		config.Video.VidOut.HdmiOut.Mirror.VideoSrc = 3

		err := json.NewEncoder(w).Encode(config)
		require.NoError(t, err)
	}))
	defer ts.Close()

	vs := &AtOmePs62{
		Address:      strings.TrimPrefix(ts.URL, "http://"),
		RequestDelay: 1 * time.Nanosecond,
	}

	inputs, err := vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string]string{"mirror": "3"}, inputs)
}

func TestGetAudioVideoInputsSeparate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var config config
		config.Video.VidOut.HdmiOut.HdmiOutA.VideoSrc = 2
		config.Video.VidOut.HdmiOut.HdmiOutB.VideoSrc = 3

		err := json.NewEncoder(w).Encode(config)
		require.NoError(t, err)
	}))
	defer ts.Close()

	vs := &AtOmePs62{
		Address:      strings.TrimPrefix(ts.URL, "http://"),
		RequestDelay: 1 * time.Nanosecond,
	}

	inputs, err := vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string]string{"hdmiOutA": "2", "hdmiOutB": "3"}, inputs)
}

func TestAuth(t *testing.T) {
	vs := &AtOmePs62{
		Username:     "username",
		Password:     "password",
		RequestDelay: 1 * time.Nanosecond,
	}

	get := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, _omePs62Endpoint, r.URL.Path)

		uname, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, vs.Username, uname)
		require.Equal(t, vs.Password, pass)

		err := json.NewEncoder(w).Encode(config{})
		require.NoError(t, err)
	}))
	defer get.Close()

	vs.Address = strings.TrimPrefix(get.URL, "http://")

	_, err := vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)

	_, err = vs.GetVolumes(context.Background(), []string{})
	require.NoError(t, err)

	_, err = vs.GetMutes(context.Background(), []string{})
	require.NoError(t, err)

	set := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, _omePs62Endpoint, r.URL.Path)

		uname, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, vs.Username, uname)
		require.Equal(t, vs.Password, pass)

		w.Write([]byte(`{"status": 200, "message": "OK"}`))
		require.NoError(t, err)
	}))
	defer set.Close()

	vs.Address = strings.TrimPrefix(set.URL, "http://")

	err = vs.SetAudioVideoInput(context.Background(), "videoOutA", "1")
	require.NoError(t, err)

	err = vs.SetVolume(context.Background(), "zoneOut1", 40)
	require.NoError(t, err)

	err = vs.SetMute(context.Background(), "zoneOut1", true)
	require.NoError(t, err)
}

func TestRateLimit(t *testing.T) {
	get := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, _omePs62Endpoint, r.URL.Path)

		err := json.NewEncoder(w).Encode(config{})
		require.NoError(t, err)
	}))
	defer get.Close()

	vs := &AtOmePs62{
		Address:      strings.TrimPrefix(get.URL, "http://"),
		RequestDelay: 300 * time.Millisecond,
	}

	start := time.Now()

	// use up the first one
	_, err := vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)

	// this one should be delayed by RequestDelay
	_, err = vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)
	require.WithinDuration(t, start.Add(vs.RequestDelay), time.Now(), 20*time.Millisecond)

	// this one should be delayed by RequestDelay*2
	_, err = vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)
	require.WithinDuration(t, start.Add(2*vs.RequestDelay), time.Now(), 20*time.Millisecond)

	// this one should be delayed by RequestDelay*3
	_, err = vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)
	require.WithinDuration(t, start.Add(3*vs.RequestDelay), time.Now(), 20*time.Millisecond)
}

func TestRateLimitCtxTooSoon(t *testing.T) {
	get := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, _omePs62Endpoint, r.URL.Path)

		err := json.NewEncoder(w).Encode(config{})
		require.NoError(t, err)
	}))
	defer get.Close()

	vs := &AtOmePs62{
		Address:      strings.TrimPrefix(get.URL, "http://"),
		RequestDelay: 500 * time.Millisecond,
	}

	// use up the first one
	_, err := vs.GetAudioVideoInputs(context.Background())
	require.NoError(t, err)

	// have a deadline sooner than RequestDelay
	ctx, cancel := context.WithTimeout(context.Background(), vs.RequestDelay/2)
	defer cancel()

	_, err = vs.GetAudioVideoInputs(ctx)
	require.EqualError(t, err, "unable to get config: unable to wait for ratelimit: rate: Wait(n=1) would exceed context deadline")
}
