//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------

//go:build windows
// +build windows

package simplefpm

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/osrg/gobgp/v3/pkg/log"
)

func SfpmReadAll(conn net.Conn, length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := io.ReadFull(conn, buf)
	return buf, err
}

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

func (sfpmc *SfpmClient) SfpmReceive() chan *SfpmMessage {
	return sfpmc.SfpmcIncoming
}

func (sfpmc *SfpmClient) SfpmcStartCh() chan bool {
	return sfpmc.SfpmcStart
}

func (sfpmc *SfpmClient) SfpmReceiveOneMsg() (*SfpmMessage, error) {
	hdrBuf, err := SfpmReadAll(sfpmc.SfpmcConn, SfpmGetHeaderSize())
	if err != nil {
		return nil, err
	}

	hdr := &SfpmHeader{}
	hdr.SfpmhParse(hdrBuf)

	msgBodySize := int(hdr.SfpmhLen) - SfpmGetHeaderSize()
	msgBuf, err := SfpmReadAll(sfpmc.SfpmcConn, msgBodySize)
	if err != nil {
		sfpmc.SfpmcLogger.Error("Error reading from msg body of size %v type %d",
			log.Fields{
				"msgSize": msgBodySize,
				"command": hdr.SfpmhCommand,
			})
		return nil, err
	}

	mm, err := SfpmParseBody(msgBuf, hdr)

	return mm, err

}

func (sfpmc *SfpmClient) sfpmStartRxLoop(address string) {

	for {
		if m, err := sfpmc.SfpmReceiveOneMsg(); err != nil {
			sfpmc.SfpmcLogger.Infof("disconnected", log.Fields{"module": "sfpmcrxloop"})
			sfpmc.SfpmcLogger.Infof("exiting", log.Fields{"module": "sfpmrxloop"})
			close(sfpmc.SfpmcIncoming)
			close(sfpmc.SfpmcOutgoing)
			sfpmc.SfpmcClose <- true
			return
		} else if m != nil {
			sfpmc.SfpmcIncoming <- m
		}
	}
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

func (sfpmc *SfpmClient) sfpmStartTxLoop(address string) {

	for {
		select {
		case m, more := <-sfpmc.SfpmcOutgoing:
			if more {
				b1, err := m.SfpmmBody.sfpmbSerialize()
				if err != nil {
					sfpmc.SfpmcLogger.Errorf("unable to serialize cmd %v, body %v, err %v",
						log.Fields{"module": "sfpmtxloop"},
						m.SfpmmHdr.SfpmhCommand,
						m.SfpmmBody,
						err)
				}
				b2 := sfpmSerializeHeader(m, uint16(len(b1)))
				b := append(b2, b1...)
				sfpmc.SfpmcLogger.Infof("sending serialized cmd %v, body %v, data %v, err %v",
					log.Fields{"module": "sfpmtxloop"},
					m.SfpmmHdr.SfpmhCommand,
					m.SfpmmBody,
					b,
					err)
				n, err := sfpmc.SfpmcConn.Write(b)
				if err != nil {
					sfpmc.SfpmcLogger.Errorf("error sending serialized data, err %v",
						log.Fields{"module": "sfpmtxloop"},
						err)
				} else {
					sfpmc.SfpmcLogger.Infof("sent %v bytes of serialized data(len=%d)",
						log.Fields{"module": "sfpmtxloop"},
						n,
						len(b))
				}
			} else {
				sfpmc.SfpmcLogger.Infof("closing tx channel, exiting loop",
					log.Fields{"module": "sfpmtxloop"})
				return
			}
		}
	}
}

func (sfpmc *SfpmClient) sfpmSend(m *SfpmMessage) {
	defer func() {
		if err := recover(); err != nil {
			sfpmc.SfpmcLogger.Infof("recovered from %v",
				log.Fields{
					"Topic": "sfpmsend"},
				err)
		}
	}()
	sfpmc.SfpmcLogger.Infof("send command %v with header %v Body %v to fpm",
		log.Fields{"Topic": "sfpmsend"},
		m.SfpmmHdr.SfpmhCommand,
		m.SfpmmHdr,
		m.SfpmmBody)
	sfpmc.SfpmcOutgoing <- m
}

func (sfpmc *SfpmClient) SfpmSendCommand(cmd SfpmcAPIType,
	vrfID uint32, body SfpmBody) error {
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

	sfpmc.sfpmSend(mm)
	return nil
}

func (sfpmc *SfpmClient) sfpmBlockingConnect(address string) {
	timeout := time.Duration(5 * time.Second)
	sfpmc.SfpmcLogger.Info("connecting to ",
		log.Fields{"address": address})
	for {
		conn, err := winio.DialPipe(address, &timeout)
		if err != nil {
			sfpmc.SfpmcLogger.Errorf("Error connecting to server %v, err %v",
				log.Fields{"module": "sfpmloop"},
				address,
				err)
			time.Sleep(timeout)
			continue
		}

		sfpmc.SfpmcConn = conn
		break
	}

	sfpmc.SfpmcLogger.Infof("connected to %s",
		log.Fields{"module": "sfpmloop"},
		address)
}

func (sfpmc *SfpmClient) SfpmStartWatcher() {
	for {
		select {
		case <-sfpmc.SfpmcClose:
			sfpmc.SfpmcConn.Close()
			sfpmc.SfpmcLogger.Infof("reconnecting to %s",
				log.Fields{"module": "sfpmwatcher"},
				sfpmc.SfpmcAddress)
			sfpmc.sfpmBlockingConnect(sfpmc.SfpmcAddress)
			sfpmc.SfpmResetState()
			go sfpmc.sfpmStartRxLoop(sfpmc.SfpmcAddress)
			go sfpmc.sfpmStartTxLoop(sfpmc.SfpmcAddress)
			sfpmc.SfpmcLogger.Infof("signal route refresh to %s via channel %v",
				log.Fields{"module": "sfpmwatcher"},
				sfpmc.SfpmcAddress,
				sfpmc.SfpmcStartCh())
			sfpmc.SfpmcStartCh() <- true
		}
	}
}

func (sfpc *SfpmClient) SfpmResetState() {
	sfpc.SfpmcOutgoing = make(chan *SfpmMessage)
	sfpc.SfpmcIncoming = make(chan *SfpmMessage)
}

func SfpmNewClient(logger log.Logger, address string) (*SfpmClient, error) {
	// conn, err := net.Dial(network, address)
	// if err != nil {
	// 	return nil, err
	// }

	outgoing := make(chan *SfpmMessage)
	incoming := make(chan *SfpmMessage)
	start := make(chan bool)

	c := &SfpmClient{
		SfpmcConn:     nil,
		SfpmcLogger:   logger,
		SfpmcIncoming: incoming,
		SfpmcOutgoing: outgoing,
		SfpmcClose:    make(chan bool),
		SfpmcAddress:  address,
		SfpmcStart:    start,
	}

	c.sfpmBlockingConnect(address)
	go c.SfpmStartWatcher()
	go c.sfpmStartRxLoop(address)
	go c.sfpmStartTxLoop(address)
	return c, nil
}
