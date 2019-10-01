package gofalcon_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"github.com/m-mizutani/gofalcon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensorAPI(t *testing.T) {
	appID := uuid.New().String()[0:8]
	output, err := commonClient.Sensor.EntitiesDatafeed(&gofalcon.EntitiesDatafeedInput{
		AppID: &appID,
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(output.Errors))
	assert.Equal(t, 1, len(output.Resources))

	if cfg.verbose {
		pp.Println(output.Resources)
	}

	ch := gofalcon.ReadEventStreamFeed(output.Resources[0])
	require.NoError(t, err)

	q := <-ch
	assert.NoError(t, q.Error)

	partition, err := output.Resources[0].Partition()
	require.NoError(t, err)

	action, err := commonClient.Sensor.EntitiesDatafeedAction(&gofalcon.EntitiesDatafeedActionInput{
		AppID:      &appID,
		ActionName: gofalcon.String("refresh_active_stream_session"),
		Partition:  &partition,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, len(action.Errors))
}

func TestEventStream(t *testing.T) {
	qCount := 0
	ch := commonClient.Sensor.EventStream()

	select {
	case q := <-ch:
		qCount++
		if cfg.verbose {
			pp.Println(*q)
		}
		assert.NoError(t, q.Error)
		break
	case <-time.After(time.Second * 10):
		break
	}

	assert.Equal(t, 1, qCount)
}
