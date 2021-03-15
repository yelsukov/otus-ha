package dialogue

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"

	"github.com/yelsukov/otus-ha/backend/errors"
)

type ServiceProvider struct {
	Token string // auth token
	Url   string // service api url
}

func (n *ServiceProvider) sendRequest(ctx context.Context, method, route string, body io.Reader) (*http.Response, error) {
	// prepare the http client
	client := &http.Client{Timeout: 1 * time.Second}

	// prepare request
	req, err := http.NewRequest(method, n.Url+route, body)
	if err != nil {
		return nil, err
	}
	// Add headers
	req.Header.Add("Authorization", "Entrypoint "+n.Token)
	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}
	// Set pass-through request id
	req.Header.Add(middleware.RequestIDHeader, middleware.GetReqID(ctx))

	// send request to dialogue service
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// prepareResponse parses response to ReadResponse / ReadError structures
func prepareResponse(res *http.Response) ([]byte, error) {
	// If got error
	if res.StatusCode != 200 {
		var r errors.KernelError
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			return nil, err
		}

		return nil, &r
	}

	// On success response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
