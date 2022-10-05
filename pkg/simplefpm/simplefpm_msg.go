//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------

//go:build windows
// +build windows

package simplefpm

import (
	"encoding/binary"

	log "github.com/sirupsen/logrus"
)

var gSfpmLogger log.Logger

func SfpmGetHeaderSize() int {
	return 10
}

func SfpmParseBody(msgBuf []byte, hdr *SfpmHeader) (*SfpmMessage, error) {
	m := &SfpmMessage{
		SfpmmHdr: *hdr,
	}
	cmd := hdr.SfpmhCommand

	switch cmd {
	case SfpmRouteAdd:
		m.SfpmmBody = &SfpmIPRouteBody{}
	}

	m.SfpmmBody.sfpmbDecodeFromBytes(msgBuf)
	return m, nil
}

func (Sfpmh *SfpmHeader) SfpmhParse(hdrBuf []byte) {
	Sfpmh.SfpmhLen = binary.BigEndian.Uint16(hdrBuf[0:2])
	Sfpmh.SfpmchMarker = hdrBuf[2]
	Sfpmh.SfpmhVersion = hdrBuf[3]
	Sfpmh.SfpmhVrfID = binary.BigEndian.Uint32(hdrBuf[4:8])
	Sfpmh.SfpmhCommand = SfpmcAPIType(binary.BigEndian.Uint16(hdrBuf[8:10]))
}

func sfpmSerializeHeader(mm *SfpmMessage, bodyLen uint16) []byte {
	headerLen := 10
	buf := make([]byte, headerLen)
	binary.BigEndian.PutUint16(buf[0:2], uint16(10+bodyLen))
	buf[2] = mm.SfpmmHdr.SfpmchMarker
	buf[3] = mm.SfpmmHdr.SfpmhVersion
	binary.BigEndian.PutUint32(buf[4:8], uint32(mm.SfpmmHdr.SfpmhVrfID))
	binary.BigEndian.PutUint16(buf[8:10], uint16(mm.SfpmmHdr.SfpmhCommand))

	return buf
}

func SfpmGetMessage(msg []byte) *SfpmMessage {
	var hdr SfpmHeader

	hdr.SfpmhParse(msg)
	mm, _ := SfpmParseBody(msg, &hdr)

	return mm
}

func sfpmSend(m *SfpmMessage, txChan chan *[]byte) {
	defer func() {
		if err := recover(); err != nil {
			gSfpmLogger.Infof("recovered from %v",
				log.Fields{
					"Topic": "sfpmsend"},
				err)
		}
	}()
	gSfpmLogger.Infof("send command %v with header %v Body %v to fpm",
		log.Fields{"Topic": "sfpmsend"},
		m.SfpmmHdr.SfpmhCommand,
		m.SfpmmHdr,
		m.SfpmmBody)
	b1, err := m.SfpmmBody.sfpmbSerialize()
	if err != nil {
		gSfpmLogger.Errorf("unable to serialize cmd %v, body %v, err %v",
			log.Fields{"module": "sfpmtxloop"},
			m.SfpmmHdr.SfpmhCommand,
			m.SfpmmBody,
			err)
	}
	b2 := sfpmSerializeHeader(m, uint16(len(b1)))
	b := append(b2, b1...)
	gSfpmLogger.Infof("sending serialized cmd %v, body %v, data %v, err %v",
		log.Fields{"module": "sfpmtxloop"},
		m.SfpmmHdr.SfpmhCommand,
		m.SfpmmBody,
		b,
		err)
	txChan <- &b
}

func SfpmSendCommand(cmd SfpmcAPIType,
	vrfID uint32,
	body SfpmBody,
	txChan chan *[]byte) error {
	mm := &SfpmMessage{
		SfpmmHdr: SfpmHeader{
			SfpmhLen:     uint16(SfpmGetHeaderSize()),
			SfpmchMarker: 0,
			SfpmhVersion: 0,
			SfpmhVrfID:   vrfID,
			SfpmhCommand: cmd,
		},
		SfpmmBody: body,
	}

	sfpmSend(mm, txChan)
	return nil
}
