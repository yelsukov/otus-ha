package dialogue

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
)

func (sp *ServiceProvider) FetchChats(ctx context.Context, uid, limit int) ([]byte, error) {
	// send request to dialogue service
	resp, err := sp.sendRequest(ctx, "GET", "/chats?uid="+strconv.Itoa(uid)+"&limit="+strconv.Itoa(limit), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return prepareResponse(resp)
}

func (sp *ServiceProvider) GetChat(ctx context.Context, uid int, cid string) ([]byte, error) {
	// send request to dialogue service
	resp, err := sp.sendRequest(ctx, "GET", "/chats/"+cid+"?uid="+strconv.Itoa(uid), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return prepareResponse(resp)
}

func (sp *ServiceProvider) CreateChat(ctx context.Context, users []int) ([]byte, error) {
	// prepare request body
	body, err := json.Marshal(map[string][]int{"users": users})
	if err != nil {
		return nil, err
	}

	resp, err := sp.sendRequest(ctx, "POST", "/chats", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return prepareResponse(resp)
}

func (sp *ServiceProvider) AddUsers2Chat(ctx context.Context, uid int, cid string, users []int) ([]byte, error) {
	// prepare request body
	body, err := json.Marshal(map[string][]int{"users": users})
	if err != nil {
		return nil, err
	}

	resp, err := sp.sendRequest(ctx, "PUT", "/chats/"+cid+"?uid="+strconv.Itoa(uid), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return prepareResponse(resp)
}
