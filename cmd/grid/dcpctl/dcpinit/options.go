package dcpinit

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
	"fmt"
	"net"

	"github.com/pkg/errors"

	"github.com/bhojpur/dcp/pkg/client/constants"
)

// InitOptions defines all the init options exposed via flags by dcpctl init.
type InitOptions struct {
	AdvertiseAddress    string
	TunnelServerAddress string
	ServiceSubnet       string
	PodSubnet           string
	Password            string
	ImageRepository     string
	BhojpurDcpVersion   string
}

func NewInitOptions() *InitOptions {
	return &InitOptions{
		ImageRepository:   constants.DefaultDcpImageRegistry,
		BhojpurDcpVersion: constants.DefaultDcpVersion,
	}
}

func (o *InitOptions) Validate() error {
	if err := validateServerAddress(o.AdvertiseAddress); err != nil {
		return err
	}
	if o.TunnelServerAddress != "" {
		if err := validateServerAddress(o.TunnelServerAddress); err != nil {
			return err
		}
	}
	if o.Password == "" {
		return fmt.Errorf("password can't be empty.")
	}

	if o.PodSubnet != "" {
		if err := validateCidrString(o.PodSubnet); err != nil {
			return err
		}
	}
	if o.ServiceSubnet != "" {
		if err := validateCidrString(o.ServiceSubnet); err != nil {
			return err
		}
	}
	return nil
}

func validateServerAddress(address string) error {
	ip := net.ParseIP(address)
	if ip == nil {
		return errors.Errorf("cannot parse IP address: %s", address)
	}
	if !ip.IsGlobalUnicast() {
		return errors.Errorf("cannot use %q as the bind address for the API Server", address)
	}
	return nil
}

func validateCidrString(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	return nil
}
