package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type JSON map[string]interface{}

func Test_main(t *testing.T) {
	port := 40000 + int(rand.Int31n(10000))
	os.Args = []string{"app", "--secret=123", "--port=" + strconv.Itoa(port), "--dsn=host=localhost port=5433 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable"}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}

	var uuid string
	var userId int = int(rand.Int31n(1000))

	{
		resp, err := http.Post(
			fmt.Sprintf("http://localhost:%d/api/v1/events", port),
			"application/json",
			strings.NewReader(`{"user_id": `+fmt.Sprint(userId)+`,"type": "test"}`))

		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		var requestData JSON
		err = json.Unmarshal(respBody, &requestData)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.NotEmpty(t, requestData["uuid"])
		uuid = requestData["uuid"].(string)
	}

	{
        resp, err := http.Post(
                    fmt.Sprintf("http://localhost:%d/api/v1/events/%s", port, uuid),
                    "application/json",
                    strings.NewReader(`{"status": "done"}`))

        require.NoError(t, err)
        defer resp.Body.Close()
        respBody, err := io.ReadAll(resp.Body)
        var requestData JSON
        err = json.Unmarshal(respBody, &requestData)
        assert.NoError(t, err)
        assert.Equal(t, uuid, requestData["uuid"])
        assert.Equal(t, 200, resp.StatusCode)
	}

	{
        resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/events/users/%d", port, userId))

        require.NoError(t, err)
        defer resp.Body.Close()
        respBody, err := io.ReadAll(resp.Body)
        var requestData JSON
        err = json.Unmarshal(respBody, &requestData)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)
    }
}

func waitForHTTPServerStart(port int) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}