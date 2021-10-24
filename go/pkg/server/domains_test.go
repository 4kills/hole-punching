package server

import (
	"testing"
)

func TestDomainAddrMap_ProcessAddress(t *testing.T) {
	tt := []struct {
		m map[string][]string
		addr string
		domain string
		expectedRet []string
		expectedInner []string
	}{
		{
			m: map[string][]string{
				"myDomain":{
					"143.92.93.227:33333",
				},
			},
			addr: "47.123.241.125:45433",
			domain: "myDomain",

			expectedRet: []string{
				"143.92.93.227:33333",
			},
			expectedInner: []string{
				"143.92.93.227:33333",
				"47.123.241.125:45433",
			},
		},
		{
			m: map[string][]string{
			},
			addr: "47.123.241.125:45433",
			domain: "myDomain",

			expectedRet: []string{
			},
			expectedInner: []string{
				"47.123.241.125:45433",
			},
		},
		{
			m: map[string][]string{
				"myDomain":{
				},
			},
			addr: "47.123.241.125:45433",
			domain: "myDomain",

			expectedRet: []string{
			},
			expectedInner: []string{
				"47.123.241.125:45433",
			},
		},
		{
			m: map[string][]string{
				"myDomain":{
					"47.123.241.125:45433",
				},
			},
			addr: "47.123.241.125:45433",
			domain: "myDomain",

			expectedRet: []string{
			},
			expectedInner: []string{
				"47.123.241.125:45433",
			},
		},
		{
			m: map[string][]string{
				"myDomain":{
					"143.92.93.227:33333",
					"47.123.241.125:45433",
				},
			},
			addr: "47.123.241.125:45433",
			domain: "myDomain",

			expectedRet: []string{
				"143.92.93.227:33333",
			},
			expectedInner: []string{
				"143.92.93.227:33333",
				"47.123.241.125:45433",
			},
		},
	}

	for _, tc := range tt {
		addrStore := domainAddrMap{m: tc.m}

		s, err := addrStore.ProcessAddress(tc.domain, tc.addr, -1)
		if err != nil {
			t.Fatal(err)
		}

		if !strSliceEquals(s, tc.expectedRet) {
			t.Errorf("got %v\n want %v", s, tc.expectedRet)
		}

		inner, ok := addrStore.m[tc.domain]
		if !ok {
			t.Errorf("got %v\n want %v", ok, true)
			continue
		}

		if !strSliceEquals(inner, tc.expectedInner) {
			t.Errorf("got %v\n want %v", inner, tc.expectedInner)
		}
	}
}

func strSliceEquals(got, want []string) bool {
	if got == nil && want == nil {
		return true
	}
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}
