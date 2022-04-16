package remotedialer

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"

	"github.com/bhojpur/dcp/pkg/tunnel/remotedialer/metrics"
	"github.com/gorilla/websocket"
)

type sessionListener interface {
	sessionAdded(clientKey string, sessionKey int64)
	sessionRemoved(clientKey string, sessionKey int64)
}

type sessionManager struct {
	sync.Mutex
	clients   map[string][]*Session
	peers     map[string][]*Session
	listeners map[sessionListener]bool
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		clients:   map[string][]*Session{},
		peers:     map[string][]*Session{},
		listeners: map[sessionListener]bool{},
	}
}

func toDialer(s *Session, prefix string) Dialer {
	return func(ctx context.Context, proto, address string) (net.Conn, error) {
		if prefix == "" {
			return s.serverConnectContext(ctx, proto, address)
		}
		return s.serverConnectContext(ctx, prefix+"::"+proto, address)
	}
}

func (sm *sessionManager) removeListener(listener sessionListener) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.listeners, listener)
}

func (sm *sessionManager) addListener(listener sessionListener) {
	sm.Lock()
	defer sm.Unlock()

	sm.listeners[listener] = true

	for k, sessions := range sm.clients {
		for _, session := range sessions {
			listener.sessionAdded(k, session.sessionKey)
		}
	}

	for k, sessions := range sm.peers {
		for _, session := range sessions {
			listener.sessionAdded(k, session.sessionKey)
		}
	}
}

func (sm *sessionManager) getDialer(clientKey string) (Dialer, error) {
	sm.Lock()
	defer sm.Unlock()

	sessions := sm.clients[clientKey]
	if len(sessions) > 0 {
		return toDialer(sessions[0], ""), nil
	}

	for _, sessions := range sm.peers {
		for _, session := range sessions {
			session.Lock()
			keys := session.remoteClientKeys[clientKey]
			session.Unlock()
			if len(keys) > 0 {
				return toDialer(session, clientKey), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find Session for client %s", clientKey)
}

func (sm *sessionManager) add(clientKey string, conn *websocket.Conn, peer bool) *Session {
	sessionKey := rand.Int63()
	session := newSession(sessionKey, clientKey, conn)

	sm.Lock()
	defer sm.Unlock()

	if peer {
		sm.peers[clientKey] = append(sm.peers[clientKey], session)
	} else {
		sm.clients[clientKey] = append(sm.clients[clientKey], session)
	}
	metrics.IncSMTotalAddWS(clientKey, peer)

	for l := range sm.listeners {
		l.sessionAdded(clientKey, session.sessionKey)
	}

	return session
}

func (sm *sessionManager) remove(s *Session) {
	var isPeer bool
	sm.Lock()
	defer sm.Unlock()

	for i, store := range []map[string][]*Session{sm.clients, sm.peers} {
		var newSessions []*Session

		for _, v := range store[s.clientKey] {
			if v.sessionKey == s.sessionKey {
				if i == 0 {
					isPeer = false
				} else {
					isPeer = true
				}
				metrics.IncSMTotalRemoveWS(s.clientKey, isPeer)
				continue
			}
			newSessions = append(newSessions, v)
		}

		if len(newSessions) == 0 {
			delete(store, s.clientKey)
		} else {
			store[s.clientKey] = newSessions
		}
	}

	for l := range sm.listeners {
		l.sessionRemoved(s.clientKey, s.sessionKey)
	}

	s.Close()
}
