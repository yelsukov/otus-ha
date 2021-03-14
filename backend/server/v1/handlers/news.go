package handlers

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/backend/server/v1/responses"
)

// authorize and proxy socket connection to client
func GetNews(addr, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cid := strconv.Itoa(int(r.Context().Value("currentUserId").(int64)))
		// add auth headers
		r.Header.Set("Authorization", token)
		r.Header.Add("x-cid", cid)

		// dial new service
		peer, err := net.Dial("tcp", addr)
		if err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		if err = r.Write(peer); err != nil {
			responses.ResponseWithError(w, err)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			responses.ResponseWithError(w, errors.New("failed to create hijacker"))
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			responses.ResponseWithError(w, errors.New("failed to hijack connection"))
			return
		}

		log.Debugf(
			"serving %s < %s <~> %s > %s",
			peer.RemoteAddr(), peer.LocalAddr(), conn.RemoteAddr(), conn.LocalAddr(),
		)

		go func() {
			defer func() { _ = peer.Close() }()
			defer func() { _ = conn.Close() }()
			_, _ = io.Copy(peer, conn)
			log.Debugf("disconnected %s < %s", peer.RemoteAddr(), peer.LocalAddr())
		}()
		go func() {
			defer func() { _ = peer.Close() }()
			defer func() { _ = conn.Close() }()
			_, _ = io.Copy(conn, peer)
			log.Debugf(
				"disconnected %s > %s",
				conn.RemoteAddr(), conn.LocalAddr(),
			)
		}()
	}
}
