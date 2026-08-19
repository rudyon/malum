package safefetch

import (
	"errors"
	"net"
	"testing"
)

func TestValidateAddressesAcceptsOnlyPublicDestinations(t *testing.T) {
	tests := []struct {
		name      string
		addresses []net.IPAddr
		wantError bool
	}{
		{name: "public IPv4", addresses: []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}}},
		{name: "public IPv6", addresses: []net.IPAddr{{IP: net.ParseIP("2606:4700:4700::1111")}}},
		{name: "loopback", addresses: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, wantError: true},
		{name: "private", addresses: []net.IPAddr{{IP: net.ParseIP("192.168.1.2")}}, wantError: true},
		{name: "carrier NAT", addresses: []net.IPAddr{{IP: net.ParseIP("100.64.0.1")}}, wantError: true},
		{name: "documentation range", addresses: []net.IPAddr{{IP: net.ParseIP("203.0.113.4")}}, wantError: true},
		{name: "mixed result", addresses: []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}, {IP: net.ParseIP("10.0.0.2")}}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAddresses(test.addresses)
			if test.wantError && !errors.Is(err, ErrPrivateDestination) {
				t.Fatalf("error = %v, want ErrPrivateDestination", err)
			}
			if !test.wantError && err != nil {
				t.Fatal(err)
			}
		})
	}
}
