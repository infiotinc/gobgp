//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------

//go:build windows
// +build windows

package simplefpm

import "encoding/binary"

func (srb *SfpmIPRouteBody) sfpmbDecodeFromBytes(msgBuf []byte) error {
	// Don't expect a route from the forwarding plane; not implemented
	return nil
}

// sfmpSerialize serialize an SfpmIPRouteBody to SfpmIPRouteWireFmt
func (srb *SfpmIPRouteBody) sfpmbSerialize() ([]byte, error) {
	var buf []byte
	numNexthop := len(srb.SrbNexthops)

	bufInitSize := 16 //fixed fields
	buf = make([]byte, bufInitSize)

	st := 0
	len := 1
	buf[st] = uint8(srb.SrbType)
	st += len

	len = 2
	binary.BigEndian.PutUint16(buf[st:st+len], uint16(srb.SrbInstance))
	st += len

	len = 1
	buf[st] = uint8(srb.Safi)
	st += len

	len = 4
	binary.BigEndian.PutUint32(buf[st:st+len], uint32(srb.SrbMetric))
	st += len

	len = 4
	binary.BigEndian.PutUint32(buf[st:st+len], uint32(srb.SrbMtu))
	st += len

	len = 2
	binary.BigEndian.PutUint16(buf[st:st+len], uint16(numNexthop))
	st += len

	buf = append(buf, srb.SrbPrefix.SfpFamily)
	buf = append(buf, srb.SrbPrefix.SfpPrefixLen)
	byteLen := (int(srb.SrbPrefix.SfpPrefixLen) + 7) / 8
	buf = append(buf, srb.SrbPrefix.SfpPrefix[:byteLen]...)
	for _, nh := range srb.SrbNexthops {
		buf = append(buf, nh.SnhGate.To4()...)
	}
	return buf, nil
}

func (sfpmc *SfpmClient) SfpmSendIPRoute(vrfID uint32,
	body *SfpmIPRouteBody, isWithdraw bool) error {

	cmd := SfpmRouteAdd
	if isWithdraw {
		cmd = SfpmRouteDel
	}

	sfpmc.SfpmSendCommand(cmd, vrfID, body)
	return nil
}
