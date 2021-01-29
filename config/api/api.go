package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mercury/x/ecode"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var defaultURL = `http://localhost:9600/infra/v1`

type API struct {
	client         *http.Client
	url            string
	header         http.Header
	cookies        []*http.Cookie
	requestTimeout time.Duration
	debug          bool
}

var (
	DefaultRequestTimeout = 30 * time.Second
)

func New(url string) *API {
	if url == "" {
		url = defaultURL
	}
	api := &API{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true, // default behavior, because of `nodeos`'s lack of support for Keep alives.
			},
		},
		url:            url,
		header:         make(http.Header),
		cookies:        make([]*http.Cookie, 0),
		requestTimeout: DefaultRequestTimeout,
	}
	api.header.Set("Accept", "application/json")

	return api
}

func (api *API) WithRequestTimeout(requestTimeout time.Duration) *API {
	api.requestTimeout = requestTimeout
	return api
}

func (api *API) Debug() *API {
	api.debug = true
	return api
}

/* ---------------------------------------- Infra API ---------------------------------------- */

const (
	loadConfigEndpoint = "config"
)

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Load config by the given service name.
func (api *API) LoadConfig(ctx context.Context) (out *LoadConfigResp, err error) {
	err = api.get(ctx, loadConfigEndpoint, nil, &out)
	return
}

func (api *API) get(ctx context.Context, endpoint string, query url.Values, out interface{}) error {
	if api.url == "" {
		return ecode.NewError("api url has not been set")
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s", api.url, endpoint))
	if err != nil {
		return err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("new request: %s", err)
	}

	return api.call(ctx, req, out)
}

func (api *API) call(ctx context.Context, req *http.Request, out interface{}) error {
	var call = func() error {
		for k, v := range api.header {
			if req.Header == nil {
				req.Header = http.Header{}
			}
			req.Header[k] = append(req.Header[k], v...)
		}

		if api.debug {
			// Useful when debugging API calls
			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("-------------------------------")
			fmt.Println(string(requestDump))
			fmt.Println("")
		}

		resp, err := api.client.Do(req.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("%s: %s", req.URL.String(), err)
		}
		defer resp.Body.Close()

		var cnt bytes.Buffer
		_, err = io.Copy(&cnt, resp.Body)
		if err != nil {
			return fmt.Errorf("copy: %s", err)
		}

		if api.debug {
			fmt.Println("RESPONSE:")
			fmt.Println("Header: ", resp.Header)
			fmt.Println("Cookies: ", resp.Cookies())
			fmt.Println("-------------------------------")
			fmt.Println(cnt.String())
			fmt.Println("-------------------------------")
			fmt.Println("")
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
		}

		var jsonResp response
		jsonResp.Data = out
		if err := json.Unmarshal(cnt.Bytes(), &jsonResp); err != nil {
			return fmt.Errorf("unmarshal err: %s", err)
		}

		if jsonResp.Code != ecode.OK.Code() {
			return fmt.Errorf("API error: %s", jsonResp.Message)
		}

		return nil
	}

	_, ok := ctx.Deadline()
	if !ok {
		ctx, _ = context.WithTimeout(ctx, api.requestTimeout)
	}

	ch := make(chan error)
	go func() {
		ch <- call()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %s", fmt.Sprintf("%v", ctx.Err()))
	case err := <-ch:
		if err == nil {
			return nil
		}

		if api.debug {
			fmt.Println("-------------------------------")
			fmt.Printf("API call error: %s, retrying...\n", err.Error())
			fmt.Println("")
		}

		return err
	}
}
