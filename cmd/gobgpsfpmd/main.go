//go:build windows
// +build windows

//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------
package main

import (
	"io"
	"net"
	"os"

	"github.com/Microsoft/go-winio"
	"github.com/jessevdk/go-flags"
	"github.com/osrg/gobgp/v3/pkg/simplefpm"
	log "github.com/sirupsen/logrus"
)

func sfpmsHandleClient(c net.Conn) {
	defer c.Close()
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())

	buf := make([]byte, 65536)
	for {
		n, err := c.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error: %v\n", err)
			}
			break
		}
		str := string(buf[:n])
		log.Printf("read %d bytes: %q\n", n, str)
		hdr := &simplefpm.SfpmHeader{}
		hdr.SfpmhParse(buf)
		//mm, err := simplefpm.SfpmParseBody(buf[10:], hdr)
		log.Printf("hdr cmd:%v, len:%d", hdr.SfpmhCommand, hdr.SfpmhLen)
	}

	log.Println("Client disconnected")
}

func sfpmsStartServer(pipePath string) {
	cc := &winio.PipeConfig{
		MessageMode:      true,
		InputBufferSize:  65536,
		OutputBufferSize: 65536,
	}
	l, err := winio.ListenPipe(pipePath, cc)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()
	log.Printf("Server listening op pipe %v\n", pipePath)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go sfpmsHandleClient(conn)
	}
}

func main() {

	var opts struct {
		SfpmsPipePath string `short:"p" long:"path" description:"specify the listen path"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Errorf("unable to parse flags, err %v", err)
		os.Exit(1)
	}

	if len(opts.SfpmsPipePath) == 0 {
		log.Fatalf("named pipe path cannot be empty, exiting")
		os.Exit(1)
	}

	sfpmsStartServer(opts.SfpmsPipePath)
}
