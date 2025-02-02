package dnstrace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_do(t *testing.T) {
	type args struct {
		server      string
		tcp         bool
		dot         bool
		dohMethod   string
		dohProtocol string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"benchmark against GoogleDNS - DNS over UDP",
			args{
				server: "8.8.8.8",
				tcp:    false,
			},
		},
		{
			"benchmark against GoogleDNS - DNS over TCP",
			args{
				server: "8.8.8.8",
				tcp:    true,
			},
		},
		{
			"benchmark against GoogleDNS - DNS over TLS",
			args{
				server: "8.8.8.8:853",
				tcp:    true,
				dot:    true,
			},
		},
		{
			"benchmark against Cloudflare - DNS over HTTPS",
			args{
				server: "https://1.1.1.1/dns-query",
			},
		},
		{
			"benchmark against Cloudflare - DNS over HTTPS - GET method",
			args{
				server:    "https://1.1.1.1/dns-query",
				dohMethod: "get",
			},
		},
		{
			"benchmark against Cloudflare - DNS over HTTPS - POST method",
			args{
				server:    "https://1.1.1.1/dns-query",
				dohMethod: "post",
			},
		},
		{
			"benchmark against Cloudflare - DNS over HTTPS/1.1",
			args{
				server:      "https://1.1.1.1/dns-query",
				dohProtocol: "1.1",
			},
		},
		{
			"benchmark against Cloudflare - DNS over HTTPS/2",
			args{
				server:      "https://1.1.1.1/dns-query",
				dohProtocol: "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time.Sleep(time.Second) // add delay before tests

			setupBenchmarkTest(tt.args.server, tt.args.tcp, tt.args.dot)
			resetPackageCounters()

			if len(tt.args.dohMethod) > 0 {
				pDoHmethod = &tt.args.dohMethod
			}
			if len(tt.args.dohProtocol) > 0 {
				pDoHProtocol = &tt.args.dohProtocol
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			rs := do(ctx)

			if assert.Len(t, rs, 2, "do(ctx) rstats") {
				if assert.NotNil(t, rs[0].hist, "do(ctx) rstats histogram") {
					assert.NotNil(t, rs[0].codes, "do(ctx) rstats codes")
					assert.Equal(t, int64(1), rs[0].codes[0], "do(ctx) rstats codes NOERROR, state:"+fmt.Sprint(rs[0].codes))
				}

				if assert.NotNil(t, rs[1].hist, "do(ctx) rstats histogram") {
					assert.NotNil(t, rs[1].codes, "do(ctx) rstats codes")
					assert.Equal(t, int64(1), rs[1].codes[0], "do(ctx) rstats codes NOERROR, state:"+fmt.Sprint(rs[1].codes))
				}

				if assert.Len(t, rs[0].timings, 1, "do(ctx) rstats timings") {
					assert.NotZero(t, rs[0].timings[0].duration, "do(ctx) rstats timings duration")
					assert.NotZero(t, rs[0].timings[0].start, "do(ctx) rstats timings start")
				}

				if assert.Len(t, rs[1].timings, 1, "do(ctx) rstats timings") {
					assert.NotZero(t, rs[1].timings[0].duration, "do(ctx) rstats timings duration")
					assert.NotZero(t, rs[1].timings[0].start, "do(ctx) rstats timings start")
				}
			}

			assert.Equal(t, int64(2), count, "total counter")
			assert.Zero(t, cerror, "connection error counter")
			assert.Zero(t, ecount, "error counter")
			assert.Equal(t, int64(2), success, "success counter")
			assert.Equal(t, int64(2), matched, "matched counter")
			assert.Zero(t, mismatch, "mismatch counter")
			assert.Zero(t, truncated, "truncated counter")
		})
	}
}

func setupBenchmarkTest(server string, tcp, dot bool) {
	pQueries = &[]string{"example.com."}

	typ := "A"
	pType = &typ

	pServer = &server
	pTCP = &tcp
	pDOT = &dot

	concurrency := uint32(2)
	pConcurrency = &concurrency

	c := int64(1)
	pCount = &c

	probability := float64(1)
	pProbability = &probability

	writeTimeout := 5 * time.Second
	pWriteTimeout = &writeTimeout

	readTimeout := 5 * time.Second
	pReadTimeout = &readTimeout

	rcodes := true
	pRCodes = &rcodes

	expect := []string{"A"}
	pExpect = &expect

	recurse := true
	pRecurse = &recurse
}
