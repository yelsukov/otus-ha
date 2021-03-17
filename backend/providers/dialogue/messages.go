package dialogue

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
)

func (sp *ServiceProvider) FetchMessages(ctx context.Context, uid int, cid string, limit int) ([]byte, error) {
	resp, err := sp.sendRequest(ctx, "GET", "/messages?uid="+strconv.Itoa(uid)+"&cid="+cid+"&limit="+strconv.Itoa(limit), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return prepareResponse(resp)
}

type PostMessageBody struct {
	UserId int    `json:"uid"`
	ChatId string `json:"cid"`
	Text   string `json:"txt"`
}

func (sp *ServiceProvider) CreateMessage(ctx context.Context, message PostMessageBody) ([]byte, error) {
	body, err := json.Marshal(&message)
	if err != nil {
		return nil, err
	}

	resp, err := sp.sendRequest(ctx, "POST", "/messages", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return prepareResponse(resp)
}
