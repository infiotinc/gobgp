// Copyright (C) 2016 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// +build windows

package server

import (
	"net"
	"syscall"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	tcpMD5SIG       = 14   // TCP MD5 Signature (RFC2385)
	ipv6MinHopCount = 73   // Generalized TTL Security Mechanism (RFC5082)
	IP_MINTTL       = 0x15 // pulled from https://golang.org/pkg/syscall/?GOOS=linux#IP_MINTTL
)

func extractFamilyFromTCPListener(l *net.TCPListener) int {
	family := syscall.AF_INET
	if strings.Contains(l.Addr().String(), "[") {
		family = syscall.AF_INET6
	}
	return family
}

func setTCPMD5SigSockopt(l *net.TCPListener, address string, key string) error {
	sc, err := l.SyscallConn()
	if err != nil {
		return err
	}
	// always enable and assumes that the configuration is done by setkey()
	return setsockOptInt_win(sc, syscall.IPPROTO_TCP, tcpMD5SIG, 1)
}

func setListenTCPTTLSockopt(l *net.TCPListener, ttl int) error {
	family := extractFamilyFromTCPListener(l)
	sc, err := l.SyscallConn()
	if err != nil {
		return err
	}
	return setsockoptIpTtl(sc, family, ttl)
}

func setTCPTTLSockopt(conn *net.TCPConn, ttl int) error {
	family := extractFamilyFromTCPConn(conn)
	sc, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	return setsockoptIpTtl(sc, family, ttl)
}

func setTCPMinTTLSockopt(conn *net.TCPConn, ttl int) error {
	family := extractFamilyFromTCPConn(conn)
	sc, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	level := syscall.IPPROTO_IP
	name := IP_MINTTL
	if family == syscall.AF_INET6 {
		level = syscall.IPPROTO_IPV6
		name = ipv6MinHopCount
	}
	return setsockOptInt_win(sc, level, name, ttl)
}

func dialerControl(network, address string, c syscall.RawConn, ttl, ttlMin uint8, password string, bindInterface string) error {
	if password != "" {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   address,
		}).Warn("setting md5 for active connection is not supported")
	}
	if ttl != 0 {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   address,
		}).Warn("setting ttl for active connection is not supported")
	}
	if ttlMin != 0 {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   address,
		}).Warn("setting min ttl for active connection is not supported")
	}
	return nil
}

func setsockOptInt(sc syscall.RawConn, level, name, value int) error {
	return setsockOptInt_win(sc, level, name, value)
}

func setsockOptInt_win(sc syscall.RawConn, level, name, value int) error {
	var opterr error
	fn := func(s uintptr) {
		opterr = syscall.SetsockoptInt(syscall.Handle(s), level, name, value)
	}
	err := sc.Control(fn)
	if opterr == nil {
		return err
	}
	return opterr
}

func setBindToDevSockopt(sc syscall.RawConn, device string) error {
	return nil
	//return setsockOptString(sc, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, device)
}