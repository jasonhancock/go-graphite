package graphite

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseUnmarshall(t *testing.T) {
	const data = `
[
    {
        "datapoints": [
            [
                null,
                1566577080
            ]
        ],
        "target": "fpi.probes.000.000.001.C"
    },
    {
        "datapoints": [
            [
                null,
                1566577080
            ]
        ],
        "target": "fpi.probes.000.000.002.C"
    },
    {
        "datapoints": [
            [
                null,
                1566577080
            ]
        ],
        "target": "fpi.probes.000.000.003.C"
    }
]
`

	var results []Series
	require.NoError(t, json.Unmarshal([]byte(data), &results))

	// TODO: add assertions for the data
}

func TestUnmarshalDataPoint(t *testing.T) {
	const in = `[null,1566577080]`
	var dp DataPoint
	require.NoError(t, json.Unmarshal([]byte(in), &dp))
}
