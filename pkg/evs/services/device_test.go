package services

import "testing"

func TestNextFreeDevice(t *testing.T) {
	all := make([]string, 0, 25)
	for ch := 'a'; ch <= 'z'; ch++ {
		all = append(all, "/dev/vd"+string(ch))
	}

	cases := []struct {
		name    string
		used    []string
		want    string
		wantErr bool
	}{
		{name: "only system disk", used: []string{"/dev/vda"}, want: "/dev/vdb"},
		{name: "no attachments", used: nil, want: "/dev/vdb"},
		{name: "fills the first gap", used: []string{"/dev/vda", "/dev/vdb", "/dev/vdd"}, want: "/dev/vdc"},
		{name: "sd names count as vd", used: []string{"/dev/vda", "/dev/sdb", "/dev/sdc"}, want: "/dev/vdd"},
		{name: "full server", used: all, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nextFreeDevice(tc.used)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("nextFreeDevice(%v) = %q, want an error", tc.used, got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("nextFreeDevice(%v) = %q, %v; want %q", tc.used, got, err, tc.want)
			}
		})
	}
}
