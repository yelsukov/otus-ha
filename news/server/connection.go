package server

import (
	"io"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mailru/easygo/netpoll"
)

type connection struct {
	id string
	net.Conn
	desc *netpoll.Desc
	t    time.Duration
}

func (c *connection) Write(p []byte) (int, error) {
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.t)); err != nil {
		return 0, err
	}

	w := wsutil.NewWriter(c.Conn, ws.StateServerSide, ws.OpText)
	n, err := w.Write(p)
	if err != nil {
		return 0, err
	}
	if err = w.Flush(); err != nil {
		return 0, err
	}

	return n, nil
}

func (c *connection) Read(p []byte) (int, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.t)); err != nil {
		return 0, err
	}
	return c.Conn.Read(p)
}

func (c *connection) Receive() (io.Reader, error) {
	h, r, err := wsutil.NextReader(c.Conn, ws.StateServerSide)
	if err != nil {
		return nil, err
	}
	if h.OpCode.IsControl() {
		return nil, wsutil.ControlFrameHandler(c.Conn, ws.StateServerSide)(h, r)
	}

	return r, nil
}
