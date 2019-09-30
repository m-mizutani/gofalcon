package gofalcon_test

import (
	"testing"

	"github.com/k0kubun/pp"
	"github.com/m-mizutani/gofalcon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectionAPI(t *testing.T) {
	output, err := commonClient.Detection.QueriesDetects(&gofalcon.QueriesDetectsInput{
		Limit: gofalcon.Int(1),
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(output.Errors))
	assert.Equal(t, 1, len(output.Resources))

	detectID := output.Resources[0]
	detail, err := commonClient.Detection.EntitySummaries(&gofalcon.EntitySummariesInput{
		ID: []string{detectID},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(detail.Resources))
	assert.NotEmpty(t, detail.Resources[0].Cid)

	if cfg.verbose {
		pp.Println(detail)
	}
}
