package gofalcon_test

import (
	"testing"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/gofalcon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryDevices(t *testing.T) {
	client := gofalcon.NewClient(cfg.user, cfg.token)
	output, err := client.Device.QueryDevices(&gofalcon.QueryDevicesInput{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(output.Errors))
	assert.NotEqual(t, 0, len(output.Resources))

	if cfg.verbose {
		pp.Println(output)
	}
}

func TestEntityDevices(t *testing.T) {
	client := gofalcon.NewClient(cfg.user, cfg.token)

	output, err := client.Device.QueryDevices(&gofalcon.QueryDevicesInput{
		Limit: gofalcon.Int(1),
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(output.Errors))
	assert.NotEqual(t, 0, len(output.Resources))

	aid := output.Resources[0]
	detail, err := client.Device.EntityDevices(&gofalcon.EntityDevicesInput{
		ID: []string{aid},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, detail.Resources[0].MacAddress)

	if cfg.verbose {
		pp.Println(detail)
	}
}
