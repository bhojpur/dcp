package server

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
	"strings"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"google.golang.org/grpc/metadata"
)

// explicit interface check
var _ etcdserverpb.ClusterServer = (*KVServerBridge)(nil)

func (s *KVServerBridge) MemberAdd(context.Context, *etcdserverpb.MemberAddRequest) (*etcdserverpb.MemberAddResponse, error) {
	return nil, fmt.Errorf("member add is not supported")
}

func (s *KVServerBridge) MemberRemove(context.Context, *etcdserverpb.MemberRemoveRequest) (*etcdserverpb.MemberRemoveResponse, error) {
	return nil, fmt.Errorf("member remove is not supported")
}

func (s *KVServerBridge) MemberUpdate(context.Context, *etcdserverpb.MemberUpdateRequest) (*etcdserverpb.MemberUpdateResponse, error) {
	return nil, fmt.Errorf("member update is not supported")
}

func (s *KVServerBridge) MemberList(ctx context.Context, r *etcdserverpb.MemberListRequest) (*etcdserverpb.MemberListResponse, error) {
	listenURL := authorityURL(ctx, s.limited.scheme)
	return &etcdserverpb.MemberListResponse{
		Header: &etcdserverpb.ResponseHeader{},
		Members: []*etcdserverpb.Member{
			{
				Name:       "statebase",
				ClientURLs: []string{listenURL},
				PeerURLs:   []string{listenURL},
			},
		},
	}, nil
}

func (s *KVServerBridge) MemberPromote(context.Context, *etcdserverpb.MemberPromoteRequest) (*etcdserverpb.MemberPromoteResponse, error) {
	return nil, fmt.Errorf("member promote is not supported")
}

// authorityURL returns the URL of the authority (host) that the client connected to.
// If no scheme is included in the authority data, the provided scheme is used. If no
// authority data is provided, the default etcd endpoint is used.
func authorityURL(ctx context.Context, scheme string) string {
	authority := "127.0.0.1:2379"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		authList := md.Get(":authority")
		if len(authList) > 0 {
			authority = authList[0]
			// etcd v3.5 encodes the endpoint address list as "#initially=[ADDRESS1;ADDRESS2]"
			if strings.HasPrefix(authority, "#initially=[") {
				authority = strings.TrimPrefix(authority, "#initially=[")
				authority = strings.TrimSuffix(authority, "]")
				authority = strings.ReplaceAll(authority, ";", ",")
				return authority
			}
		}
	}
	return scheme + "://" + authority
}
