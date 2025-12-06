package main

import (
	"encoding/hex"
	"strings"
	"testing"
)

func Test_fprintID(t *testing.T) {
	tests := []struct {
		id   []byte
		want string
	}{
		{
			id: []byte{0x02, 0x6b, 0x53, 0xd2, 0xbe},
			want: `hex      : 026b53d2be
dec(40)  : 0010390590142
dec(32)  : 1800655550
dec(24)  : 05493438
dec(8+16): 083,53950
`,
		},
		{
			id: []byte{0xff, 0xff, 0xff, 0xff, 0xff},
			want: `hex      : ffffffffff
dec(40)  : 1099511627775
dec(32)  : 4294967295
dec(24)  : 16777215
dec(8+16): 255,65535
`,
		},
		{
			id: []byte{0x00, 0x00, 0x00, 0x00, 0x00},
			want: `hex      : 0000000000
dec(40)  : 0000000000000
dec(32)  : 0000000000
dec(24)  : 00000000
dec(8+16): 000,00000
`,
		},
	}
	for _, tt := range tests {
		t.Run(hex.EncodeToString(tt.id), func(t *testing.T) {
			var got strings.Builder
			fprintID(&got, tt.id)
			if got.String() != tt.want {
				t.Errorf("fprintID() w = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
