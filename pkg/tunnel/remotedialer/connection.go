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
	"io"
	"net"
	"time"

	"github.com/bhojpur/dcp/pkg/tunnel/remotedialer/metrics"
	"github.com/sirupsen/logrus"
)

type connection struct {
	err           error
	writeDeadline time.Time
	backPressure  *backPressure
	buffer        *readBuffer
	addr          addr
	session       *Session
	connID        int64
}

func newConnection(connID int64, session *Session, proto, address string) *connection {
	c := &connection{
		addr: addr{
			proto:   proto,
			address: address,
		},
		connID:  connID,
		session: session,
	}
	c.backPressure = newBackPressure(c)
	c.buffer = newReadBuffer(connID, c.backPressure)
	metrics.IncSMTotalAddConnectionsForWS(session.clientKey, proto, address)
	return c
}

func (c *connection) tunnelClose(err error) {
	metrics.IncSMTotalRemoveConnectionsForWS(c.session.clientKey, c.addr.Network(), c.addr.String())
	c.writeErr(err)
	c.doTunnelClose(err)
}

func (c *connection) doTunnelClose(err error) {
	if c.err != nil {
		return
	}

	c.err = err
	if c.err == nil {
		c.err = io.ErrClosedPipe
	}

	c.buffer.Close(c.err)
}

func (c *connection) OnData(m *message) error {
	if PrintTunnelData {
		defer func() {
			logrus.Debugf("ONDATA  [%d] %s", c.connID, c.buffer.Status())
		}()
	}
	return c.buffer.Offer(m.body)
}

func (c *connection) Close() error {
	c.session.closeConnection(c.connID, io.EOF)
	c.backPressure.Close()
	return nil
}

func (c *connection) Read(b []byte) (int, error) {
	n, err := c.buffer.Read(b)
	metrics.AddSMTotalReceiveBytesOnWS(c.session.clientKey, float64(n))
	if PrintTunnelData {
		logrus.Debugf("READ    [%d] %s %d %v", c.connID, c.buffer.Status(), n, err)
	}
	return n, err
}

func (c *connection) Write(b []byte) (int, error) {
	if c.err != nil {
		return 0, io.ErrClosedPipe
	}
	c.backPressure.Wait()
	msg := newMessage(c.connID, b)
	metrics.AddSMTotalTransmitBytesOnWS(c.session.clientKey, float64(len(msg.Bytes())))
	return c.session.writeMessage(c.writeDeadline, msg)
}

func (c *connection) OnPause() {
	c.backPressure.OnPause()
}

func (c *connection) OnResume() {
	c.backPressure.OnResume()
}

func (c *connection) Pause() {
	msg := newPause(c.connID)
	_, _ = c.session.writeMessage(c.writeDeadline, msg)
}

func (c *connection) Resume() {
	msg := newResume(c.connID)
	_, _ = c.session.writeMessage(c.writeDeadline, msg)
}

func (c *connection) writeErr(err error) {
	if err != nil {
		msg := newErrorMessage(c.connID, err)
		metrics.AddSMTotalTransmitErrorBytesOnWS(c.session.clientKey, float64(len(msg.Bytes())))
		c.session.writeMessage(c.writeDeadline, msg)
	}
}

func (c *connection) LocalAddr() net.Addr {
	return c.addr
}

func (c *connection) RemoteAddr() net.Addr {
	return c.addr
}

func (c *connection) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	c.buffer.deadline = t
	return nil
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

type addr struct {
	proto   string
	address string
}

func (a addr) Network() string {
	return a.proto
}

func (a addr) String() string {
	return a.address
}
