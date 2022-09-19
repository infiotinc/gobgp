//-----------------------------------------------------------------------------
//  Copyright (c) 2018 Infiot Inc.
//  All rights reserved.
//-----------------------------------------------------------------------------

package server

import (
	"net"
	"strconv"
	"strings"

	nscom "github.com/infiotinc/gonscom"
	"github.com/osrg/gobgp/v3/internal/pkg/simplefpm"
	"github.com/osrg/gobgp/v3/internal/pkg/table"
	"github.com/osrg/gobgp/v3/pkg/log"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	logrus "github.com/sirupsen/logrus"
)

type SfmpClientCtx struct {
	sccClient *nscom.NSComClient
	sccBgpSrv *BgpServer
}

func (scc *SfmpClientCtx) sccLoop() {
	var ww *watcher
	ww = scc.sccBgpSrv.watch([]watchOption{
		watchBestPath(true),
	}...)

	scc.sccBgpSrv.logger.Infof("start channel %v",
		log.Fields{"module": "NSComClient"},
		scc.sccClient.NscStartCh())
	for {
		select {
		case <-scc.sccClient.NscStartCh():
			if ww != nil {
				scc.sccBgpSrv.logger.Infof("stopping running watcher",
					log.Fields{"module": "NSComClient"})
				ww.Stop()
			}
			scc.sccBgpSrv.logger.Infof("starting new watcher",
				log.Fields{"module": "NSComClient"})
			ww = scc.sccBgpSrv.watch([]watchOption{
				watchBestPath(true),
			}...)
		case msg := <-scc.sccClient.NscReceive():
			if msg == nil {
				break
			}

			mm := simplefpm.SfpmGetMessage(*msg)
			switch body := mm.SfpmmBody.(type) {
			case *simplefpm.SfpmIPRouteBody:
				scc.sccBgpSrv.logger.Info("NSComClient: received from server",
					log.Fields{
						"from": "SfpmServer",
						"body": body,
					})
			}
		case ev := <-ww.Event():
			switch msg := ev.(type) {
			case *watchEventBestPath:
				scc.sccBgpSrv.logger.Infof("best path update %v",
					log.Fields{"module": "NSComClient"},
					msg)
				if table.UseMultiplePaths.Enabled {
					for _, paths := range msg.MultiPathList {
						//FIXME also handle per VRF routes here
						body, isWithdraw := NewSimpleFpmIPRouteBody(paths, 0)
						simplefpm.SfpmSendIPRoute(0,
							body,
							isWithdraw,
							scc.sccClient.NscOutgoing)
					}
				} else {
					for _, path := range msg.PathList {
						//FIXME also handle per VRF routes here
						body, isWithdraw := NewSimpleFpmIPRouteBody([]*table.Path{path}, 0)
						simplefpm.SfpmSendIPRoute(0,
							body,
							isWithdraw,
							scc.sccClient.NscOutgoing)
					}
				}
			}
		}
	}
}

func newSimpleFpmClient(s *BgpServer, url string) (*SfmpClientCtx, error) {
	ll := logrus.New()
	cli, _ := nscom.NscNewClient(ll, url, "GoBGP")

	ww := &SfmpClientCtx{
		sccClient: cli,
		sccBgpSrv: s,
	}

	go ww.sccLoop()

	return ww, nil
}

func NewSimpleFpmIPRouteBody(dst []*table.Path,
	vrfID uint32) (*simplefpm.SfpmIPRouteBody, bool) {

	paths := filterOutExternalPath(dst)
	if len(paths) == 0 {
		return nil, false
	}

	var prefix net.IP
	var nexthop simplefpm.SfpmNexthop
	path := paths[0]
	l := strings.SplitN(path.GetNlri().String(), "/", 2)

	nexthops := make([]simplefpm.SfpmNexthop, 0, len(paths))
	switch path.GetRouteFamily() {
	case bgp.RF_IPv4_UC:
		prefix = path.GetNlri().(*bgp.IPAddrPrefix).IPAddrPrefixDefault.Prefix.To4()
	case bgp.RF_IPv4_VPN:
		prefix = path.GetNlri().(*bgp.LabeledVPNIPAddrPrefix).IPAddrPrefixDefault.Prefix.To4()
	case bgp.RF_IPv6_UC:
		prefix = path.GetNlri().(*bgp.IPv6AddrPrefix).IPAddrPrefixDefault.Prefix.To16()
	case bgp.RF_IPv6_VPN:
		prefix = path.GetNlri().(*bgp.LabeledVPNIPv6AddrPrefix).IPAddrPrefixDefault.Prefix.To16()
	default:
		return nil, false
	}

	for _, p := range paths {
		nexthop.SnhGate = p.GetNexthop()
		nexthop.SnhVrfID = 0 //FIXME: use the actual VRF eventually
		nexthops = append(nexthops, nexthop)
	}

	plen, _ := strconv.ParseUint(l[1], 10, 8)
	return &simplefpm.SfpmIPRouteBody{
		SrbType: simplefpm.SfpmRouteBGP,
		SrbPrefix: simplefpm.SfpmPrefix{
			SfpPrefix:    prefix,
			SfpPrefixLen: uint8(plen),
		},
		SrbNexthops: nexthops,
	}, path.IsWithdraw
}
