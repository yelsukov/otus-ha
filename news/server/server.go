package server

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/mailru/easygo/netpoll"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	"github.com/yelsukov/otus-ha/news/vars"
)

type Server struct {
	pool    entities.Pool
	poller  netpoll.Poller
	ns      map[string]*connection
	exitCh  chan error
	WriteCh chan *models.Event
	mu      sync.RWMutex
	mux     Mux
}

type Mux struct {
	OnConnect  func(w io.Writer, cid string)
	OnReadData func(w io.Writer, r io.Reader, cid string)
}

func NewServer(pool entities.Pool, poller netpoll.Poller, mux Mux) *Server {
	return &Server{
		pool:    pool,
		poller:  poller,
		ns:      make(map[string]*connection),
		exitCh:  make(chan error),
		WriteCh: make(chan *models.Event, 1),
		mux:     mux,
	}
}

func (s *Server) Serve(port string) error {
	// Create incoming connections listener
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Info("websocket is listening on " + ln.Addr().String())

	// Create netpoll descriptor for the listener. OneShot here to manually resume events stream
	acceptDesc := netpoll.Must(netpoll.HandleListener(
		ln, netpoll.EventRead|netpoll.EventOneShot,
	))

	// Subscribe to events about listener
	err = s.poller.Start(acceptDesc, func(_ netpoll.Event) {
		if conn, err := ln.Accept(); err == nil {
			if cid, err := establishWs(conn); err == nil {
				s.registerConnection(cid, conn)
			} else {
				_ = conn.Close()
				log.WithError(err).Error("failed upgrade connection")
			}
		} else if ne, ok := err.(net.Error); !ok || !ne.Temporary() {
			log.WithError(err).Error("failed on connection acceptance")
			s.Shutdown(err)
			return
		}

		_ = s.poller.Resume(acceptDesc)
	})
	if err != nil {
		return err
	}

	go s.writer()

	err = <-s.exitCh
	log.Info("server exited")
	return err
}

func (s *Server) Shutdown(err error) {
	close(s.WriteCh)
	for cid := range s.ns {
		s.removeConnection(cid)
	}
	s.exitCh <- err
}

// Register registers new connection as a User.
func (s *Server) registerConnection(cid string, conn net.Conn) {
	// We want to handle only read events of it.
	desc := netpoll.Must(netpoll.HandleRead(conn))

	// Subscribe to events about conn.
	if err := s.poller.Start(desc, s.receiveData(cid)); err != nil {
		log.WithError(err).Error("failed to subscribe connection")
		_ = conn.Close()
		return
	}

	s.mu.Lock()
	s.ns[cid] = &connection{cid, conn, desc, 1 * time.Second}
	s.mu.Unlock()

	if s.mux.OnConnect != nil {
		s.mux.OnConnect(s.ns[cid], cid)
	}

	log.Infof("%s: has been registered", conn.LocalAddr().String()+" > "+conn.RemoteAddr().String())
}

// Removes connection by id
func (s *Server) removeConnection(cid string) {
	log.Infof("Removing connection #%s", cid)

	conn, has := s.ns[cid]
	if !has {
		return
	}
	s.mu.Lock()
	delete(s.ns, cid)
	s.mu.Unlock()

	// close connection
	if err := conn.Close(); err != nil {
		log.WithError(err).Error("failed to close socket connection")
	}

	// unsubscribe connection listener
	if err := s.poller.Stop(conn.desc); err != nil {
		log.WithError(err).Error("failed to unsubscribe connection")
	}
}

func (s *Server) receiveData(cid string) func(ev netpoll.Event) {
	return func(ev netpoll.Event) {
		// Client has closed connection. Stop poller and disconnect
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			s.removeConnection(cid)
			return
		}
		if s.mux.OnReadData == nil {
			// have no handler for inbound messages
			return
		}

		// try to find connection
		conn, has := s.ns[cid]
		if !has {
			return
		}
		// try to read data from connection
		s.pool.Schedule(func() {
			r, err := conn.Receive()
			if err != nil {
				s.removeConnection(cid)
				return
			}
			s.mux.OnReadData(conn, r, cid)
		})
	}
}

// TODO abstract it
func (s *Server) writer() {
	for event := range s.WriteCh {
		payload, err := json.Marshal(&entities.EventResponse{Object: "event", Event: event})
		if err != nil {
			log.WithError(err).Error("failed to parse event to json")
			continue
		}

		for _, fid := range event.Followers {
			conn, has := s.ns[strconv.Itoa(fid)]
			if !has {
				continue
			}
			s.pool.Schedule(func() {
				if _, err := conn.Write(payload); err != nil {
					log.WithError(err).Error("failed to write message to socket")
				}
			})
		}
	}
	log.Info("Connection writer closed")
}

func establishWs(conn net.Conn) (string, error) {
	var cid string
	var authorized bool

	// Zero-copy upgrade to WebSocket connection.
	upgrader := ws.Upgrader{
		OnHeader: func(key, value []byte) error {
			switch string(key) {
			case "X-Cid":
				cid = string(value) // todo to int
			case "Authorization":
				token := string(value)
				if token != vars.TOKEN {
					return ws.RejectConnectionError(
						ws.RejectionReason("invalid authorization token"),
						ws.RejectionStatus(403),
						ws.RejectionHeader(ws.HandshakeHeaderString("Authorization: "+token+"\r\n")),
					)
				}
				authorized = true
			}

			return nil
		},
		OnBeforeUpgrade: func() (header ws.HandshakeHeader, err error) {
			if cid == "" {
				err = ws.RejectConnectionError(
					ws.RejectionReason("Missed `x-cid` header"),
					ws.RejectionStatus(400),
				)
			}
			if !authorized {
				err = ws.RejectConnectionError(
					ws.RejectionReason("Missed Authorization header"),
					ws.RejectionStatus(401),
				)
			}
			return
		},
	}
	if _, err := upgrader.Upgrade(conn); err != nil {
		return "", err
	}
	if cid == "" {
		return "", errors.New("uid not found in headers")
	}

	return cid, nil
}
