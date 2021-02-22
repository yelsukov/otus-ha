package news

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/backend/errors"
	"github.com/yelsukov/otus-ha/backend/models"
)

type NewsService struct {
	token string // auth token
	url   string // service api url
}

type ReadResponse struct {
	Object string        `json:"object"`
	Data   []models.News `json:"data"`
}

type ReadError struct {
	Code    string `json:"code"`
	Message string
}

func NewNewsService(token, url string) *NewsService {
	return &NewsService{token, url}
}

func (n *NewsService) Read(uid int) (*ReadResponse, error) {
	// prepare the http client
	client := &http.Client{Timeout: 1 * time.Second}

	// prepare request
	req, err := http.NewRequest("GET", n.url+"/events?fid="+strconv.Itoa(uid), nil)
	if err != nil {
		return nil, err
	}
	// Add auth token to headers
	req.Header.Add("Authorization", "Backend "+n.token)
	// send request to news service
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.WithError(err).Error("can not close response body")
		}
	}()

	// If got error
	if resp.StatusCode != 200 {
		var r ReadError
		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, err
		}
		return nil, errors.New(r.Code, r.Message)
	}

	// On success response
	var r ReadResponse
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}
