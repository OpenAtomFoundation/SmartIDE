package compose

import (
	"fmt"
	"strings"
	"testing"
)

func TestPortSimple(t *testing.T) {
	tests := []struct {
		item          string
		wantHost      string
		wantContainer string
		wantProtocol  string
		wantErr       bool
	}{
		{item: "\"3000\"", wantHost: "3000", wantErr: false},
		{item: "3000:6000", wantHost: "3000", wantContainer: "6000", wantErr: false},
		{item: "3000-4000:6000", wantHost: "3000-4000", wantContainer: "6000", wantErr: false},
		{item: "3000:6000/udp", wantHost: "3000", wantContainer: "6000", wantProtocol: "udp", wantErr: false},
		{item: "3000-4000:6000/udp", wantHost: "3000-4000", wantContainer: "6000", wantProtocol: "udp", wantErr: false},
		{item: "3000-4000:6000-7000/udp", wantHost: "3000-4000", wantContainer: "6000-7000", wantProtocol: "udp", wantErr: false},
		{item: "127.0.0.1:3000-4000:6000-7000/udp", wantHost: "127.0.0.1:3000-4000", wantContainer: "6000-7000", wantProtocol: "udp", wantErr: false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// MarshalYaml
			if !tt.wantErr {
				item := PortSimple{Host: tt.wantHost, Container: tt.wantContainer, Protocol: tt.wantProtocol}
				content := MarshalYaml(item)
				content = strings.TrimRight(content, "\n")
				if content != tt.item {
					t.Logf("%d %d", len(content), len(tt.item))
					t.Errorf("PortSimple.MarshalYAML() content = %v, wantContent %v", content, tt.item)
					return
				}
			}
			// UnmarshalYaml
			var item PortSimple
			err := UnmarshalYaml(tt.item, &item)
			if (err != nil) != tt.wantErr {
				t.Errorf("PortSimple.UnarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if item.Host != tt.wantHost {
				t.Errorf("PortSimple.UnarshalYAML() host = %v, wantHost %v", item.Host, tt.wantHost)
				return
			}
			if item.Container != tt.wantContainer {
				t.Errorf("Image.UnarshalYAML() container = %v, wantContainer %v", item.Container, tt.wantContainer)
				return
			}
			if item.Protocol != tt.wantProtocol {
				t.Errorf("PortSimple.UnarshalYAML() protocol = %v, wantProtocol %v", item.Protocol, tt.wantProtocol)
				return
			}
		})
	}
}
