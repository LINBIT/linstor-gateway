package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseURL(str string) *url.URL {
	u, _ := url.Parse(str)
	return u
}

type testData struct {
	A string `json:"a"`
	B int    `json:"b"`
}

func TestNewRequest(t *testing.T) {
	type params struct {
		method string
		path   string
		body   interface{}
	}
	type wantData struct {
		Method      string
		URL         *url.URL
		Body        string
		ContentType string
	}
	cases := []struct {
		name      string
		in        params
		want      wantData
		wantError bool
	}{{
		name:      `invalid URL`,
		in:        params{method: "GET", path: "ht tp://localhost", body: nil},
		wantError: true,
	}, {
		name: `default case`,
		in:   params{method: "GET", path: "/test", body: nil},
		want: wantData{Method: "GET", URL: parseURL("http://localhost:8337/test"), Body: ""},
	}, {
		name: `body`,
		in: params{method: "POST", path: "/test", body: testData{
			A: "test",
			B: 4711,
		}},
		want: wantData{Method: "POST", URL: parseURL("http://localhost:8337/test"), Body: "{\"a\":\"test\",\"b\":4711}\n", ContentType: "application/json"},
	}, {
		name:      `invalid body`,
		in:        params{method: "POST", path: "/test", body: make(chan int)}, // channels cannot be marshalled, causing json.Marshal to fail,
		wantError: true,
	}, {
		name:      `invalid method`,
		in:        params{method: "PO:ST"},
		wantError: true,
	}}

	cli, err := NewClient(Log(t))
	assert.NoError(t, err)

	t.Parallel()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cli.newRequest(tt.in.method, tt.in.path, tt.in.body)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Method, got.Method)
				assert.Equal(t, tt.want.URL, got.URL)
				if tt.want.Body == "" {
					assert.Nil(t, got.Body)
				} else {
					assert.NotNil(t, got.Body)
					gotBytes, err := ioutil.ReadAll(got.Body)
					assert.NoError(t, err)
					assert.Equal(t, tt.want.Body, string(gotBytes))
				}
				if tt.want.ContentType != "" {
					assert.Contains(t, got.Header, "Content-Type")
					assert.Len(t, got.Header["Content-Type"], 1)
					assert.Equal(t, tt.want.ContentType, got.Header["Content-Type"][0])
				}
			}
		})
	}
}

func TestDo(t *testing.T) {
	type wantData struct {
		status int
		body   interface{}
	}
	cases := []struct {
		name      string
		method    string
		handler   http.HandlerFunc
		want      wantData
		wantError bool
	}{{
		name:   `default case`,
		method: "GET",
		handler: func(w http.ResponseWriter, r *http.Request) {
			want := testData{A: "teststring", B: 4711}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&want)
			if err != nil {
				t.Fatal(err)
			}
		},
		want: wantData{
			status: 200,
			body:   testData{A: "teststring", B: 4711},
		},
	}, {
		name:   `error 404`,
		method: "GET",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		wantError: true,
	}, {
		name:   `error 500`,
		method: "GET",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		wantError: true,
	}}

	t.Parallel()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			base, err := url.Parse(server.URL)
			require.NoError(t, err)

			cli, err := NewClient(BaseURL(base), Log(t))
			require.NoError(t, err)

			req, err := cli.newRequest(tt.method, "/testurl", nil)
			require.NoError(t, err)

			var got testData
			resp, err := cli.do(context.Background(), req, &got)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want.status, resp.StatusCode)
				require.Equal(t, tt.want.body, got)
			}
		})
	}
}
