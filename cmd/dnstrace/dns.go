package dnstrace

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/miekg/dns"
)

func dialConnection(srv, network string, m *dns.Msg) (*dns.Conn, error) {
	co, err := dial(srv, network)
	if err != nil {
		atomic.AddInt64(&cerror, 1)

		if *pIOErrors {
			fmt.Fprintln(os.Stderr, "i/o error dialing: ", err)
		}
		return nil, err
	}
	if udpSize := *pUDPSize; udpSize > 0 {
		m.SetEdns0(udpSize, true)
		co.UDPSize = udpSize
	}
	if ednsOpt := *pEdnsOpt; len(ednsOpt) > 0 {
		o := m.IsEdns0()
		if o == nil {
			m.SetEdns0(4096, true)
			o = m.IsEdns0()
		}
		s := strings.Split(ednsOpt, ":")
		data, err := hex.DecodeString(s[1])
		if err != nil {
			panic(err)
		}
		code, err := strconv.ParseUint(s[0], 10, 16)
		if err != nil {
			panic(err)
		}
		o.Option = append(o.Option, &dns.EDNS0_LOCAL{Code: uint16(code), Data: data})
	}
	return co, err
}

func dial(srv string, network string) (*dns.Conn, error) {
	if *pDOT {
		return dns.DialTimeoutWithTLS(network, srv, &tls.Config{}, dnsTimeout)
	}
	return dns.DialTimeout(network, srv, dnsTimeout)
}
