//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------

package simplefpm

import (
	"net"
)

// NamedPipeMsgID for all SFPM messages is 1
const SfpmNamedPipeMsgID = uint16(1)

type SfpmcAPIType uint16
type SfpmRouteType uint8
type SfpmFlag uint64
type SfpmMessageFlag uint32
type SfpmSafi uint8
type SfpmNexthopType uint8

const (
	SfpmRouteSystem SfpmRouteType = iota //0
	SfpmRouteBGP                  = 9
)

const (
	SfpmInterfaceAdd SfpmcAPIType = iota
	SfpmRouteAdd
	SfpmRouteDel
)

type SfpmBody interface {
	sfpmbDecodeFromBytes([]byte) error
	sfpmbSerialize() ([]byte, error)
}

type SfpmMessage struct {
	SfpmmHdr  SfpmHeader
	SfpmmBody SfpmBody
}

const SfpmClientReadSize = 65536

type SfpmHeader struct {
	SfpmhLen     uint16
	SfpmchMarker uint8
	SfpmhVersion uint8
	SfpmhVrfID   uint32
	SfpmhCommand SfpmcAPIType
}

type SfpmPrefix struct {
	SfpFamily    uint8
	SfpPrefixLen uint8
	SfpPrefix    net.IP
}

type SfpmNexthop struct {
	SnhType  SfpmNexthopType
	SnhVrfID uint32
	Snhflags uint8
	SnhGate  net.IP
}

type SfpmIPRouteBody struct {
	SrbType           SfpmRouteType
	SrbInstance       uint16
	Safi              SfpmSafi
	SrbPrefix         SfpmPrefix
	SrbSrcPrefix      SfpmPrefix
	SrbNexthops       []SfpmNexthop
	SrbbackupNexthops []SfpmNexthop
	SrbNhgid          uint32
	SrbDistance       uint8
	SrbMetric         uint32
	SrbTag            uint32
	SrbMtu            uint32
	SrbTableID        uint32
}

// this is solely to describe the wire format of a IPRoute Message Body
// on the wire
type SfpmIPRouteWireFmt struct {
	SwfType         SfpmRouteType
	SwfInstance     uint16
	SwfSafi         uint8
	SwfMetric       uint32
	SwfMtu          uint32
	SwfNumNexthops  uint16
	SwfPrefixFamily uint8
	SwfPrefixLen    uint8
	SwfPrefix       []byte //length depends on the prefix len
	SwfNexthops     []byte //SwfNumNexthops entries
}

type SfpmHeaderWireFmt struct {
	SwfMsgLen  uint16
	SwfMarker  uint8
	SwfVersion uint8
	SwfVrfID   uint32
	SwfCmd     uint16
}
