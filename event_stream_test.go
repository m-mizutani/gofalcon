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
	ch := commonClient.Sensor.EventStream(nil)

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

func TestMultipleEventStream(t *testing.T) {
	qCount1 := 0
	qCount2 := 0
	ch1 := commonClient.Sensor.EventStream(nil)
	ch2 := commonClient.Sensor.EventStream(nil)

Loop:
	for {
		select {
		case q := <-ch1:
			qCount1++
			assert.NoError(t, q.Error)
			if qCount1 > 0 && qCount2 > 0 {
				break Loop
			}

		case q := <-ch2:
			qCount2++
			assert.NoError(t, q.Error)
			if qCount1 > 0 && qCount2 > 0 {
				break Loop
			}

		case <-time.After(time.Second * 10):
			break Loop
		}
	}

	assert.NotEqual(t, 0, qCount1)
	assert.NotEqual(t, 0, qCount2)
}
