package graphite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// Doer is an abstraction around an http client.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client talks to the Graphite HTTP API.
type Client struct {
	client Doer
	url    url.URL
}

// NewClient initializes a client. The url passed here should be the base url of
// your graphite installation. Thus, if your render API url is located at
// "https://www.example.com:9090/foo/render", it should be
// "https://www.example.com:9090/foo".
func NewClient(doer Doer, u url.URL) *Client {
	u.Path = path.Join(u.Path, "render")
	return &Client{
		client: doer,
		url:    u,
	}
}

// Render initiates a request to the graphite server, parses and returns the results.
func (c *Client) Render(ctx context.Context, r RenderRequest) (RenderResponse, error) {
	u := c.url
	u.Path = path.Join(u.Path, "render")
	u.RawQuery = r.Values().Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("constructing request: %w", err)
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RenderResponse{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var data []Series
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return RenderResponse{}, fmt.Errorf("decoding response: %w", err)
	}
	return RenderResponse{Series: data}, nil
}

// Time is an interface for defining time periods.
type Time interface {
	String() string
}

// RelativeTime is a relative time format. Examples would include "-4h", "-6d", etc.
type RelativeTime struct {
	Time string
}

// String returns the string representation of the time.
func (t *RelativeTime) String() string {
	return t.Time
}

// AbsoluteTime represents and absolute moment in time.
type AbsoluteTime struct {
	Time time.Time
}

// String returns the unix timestamp of the time.
func (t *AbsoluteTime) String() string {
	return strconv.FormatInt(t.Time.Unix(), 10)
}

// RenderRequest represents a request to the graphite server to retrieve one or more time series.
type RenderRequest struct {
	From    Time
	Until   Time
	Targets []string
}

// Values returns a url.Values with all of the attributes of the request.
func (r RenderRequest) Values() url.Values {
	values := url.Values{
		"format": []string{"json"},
		"target": r.Targets,
	}
	if r.From != nil {
		values.Set("from", r.From.String())
	}
	if r.Until != nil {
		values.Set("until", r.Until.String())
	}

	return values
}

// RenderResponse is the response from the graphite server.
type RenderResponse struct {
	Series []Series `json:"series"`
}

// Series represents one time series worth of data.
type Series struct {
	DataPoints []DataPoint `json:"datapoints"`
	Target     string      `json:"target"`
}

// DataPoint represents a single point of data. It has a timestamp and value.
type DataPoint struct {
	Timestamp int64
	Value     *float64
}

// UnmarshalJSON is a custom unmarshaler for a DataPoint. This is because graphite
// returns the data as an array with the timestamp and value ("[<timestamp>,<value>]")
// and we need to unwind that into something more usable.
func (p *DataPoint) UnmarshalJSON(buf []byte) error {
	tmp := []interface{}{&p.Value, &p.Timestamp}
	wantLen := len(tmp)
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if g, e := len(tmp), wantLen; g != e {
		return fmt.Errorf("wrong number of fields in datapoint: %d != %d", g, e)
	}
	return nil
}
