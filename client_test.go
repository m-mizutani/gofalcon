package gofalcon_test

import (
	"log"
	"os"
	"testing"

	"github.com/m-mizutani/gofalcon"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	user     string
	token    string
	clientID string
	secret   string
	verbose  bool
}

var cfg testConfig
var commonClient *gofalcon.Client

func init() {
	cfg = testConfig{
		user:     os.Getenv("FALCON_USER"),
		token:    os.Getenv("FALCON_TOKEN"),
		clientID: os.Getenv("FALCON_CLIENT_ID"),
		secret:   os.Getenv("FALCON_SECRET"),
		verbose:  (os.Getenv("FALCON_TEST_VERBOSE") != ""),
	}

	commonClient = gofalcon.NewClient()
	err := commonClient.EnableOAuth2(cfg.clientID, cfg.secret)
	if err != nil {
		log.Fatal("Fail oauth2", err)
	}

	gofalcon.Logger.SetLevel(logrus.DebugLevel)
}

func TestClientOAuth2(t *testing.T) {
	client := gofalcon.NewClient()
	err := client.EnableOAuth2(cfg.clientID, cfg.secret)
	require.NoError(t, err)

	output, err := client.Device.QueryDevices(&gofalcon.QueryDevicesInput{
		Limit: gofalcon.Int(1),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(output.Resources))
}
