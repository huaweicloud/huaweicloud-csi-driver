package services

import (
	"testing"

	"github.com/chnsz/golangsdk"
)

// newErr400 builds an ErrDefault400 whose Body carries the given SFS Turbo
// error envelope JSON. This mirrors what the golangsdk returns when the SFS
// Turbo dir-quota API responds with HTTP 400.
func newErr400(body string) error {
	return golangsdk.ErrDefault400{
		ErrUnexpectedResponseCode: golangsdk.ErrUnexpectedResponseCode{
			BaseError: golangsdk.BaseError{},
			URL:       "https://sfs-turbo.example/v1/shares/x/fs/dir-quota?path=%2Fa",
			Method:    "GET",
			RequestId: "req-1",
			Expected:  []int{200},
			Actual:    400,
			Body:      []byte(body),
		},
	}
}

func TestParseSFSTurboErrCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
		wantOk   bool
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "non-HTTP error",
			err:      errSimple{},
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "400 with SFS.TURBO.0113 body",
			err:      newErr400(`{"errCode":"SFS.TURBO.0113","errMsg":"dir not create quota"}`),
			wantCode: "SFS.TURBO.0113",
			wantOk:   true,
		},
		{
			name:     "400 with empty body",
			err:      newErr400(``),
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "400 with non-JSON body",
			err:      newErr400(`not json`),
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "400 with JSON but no errCode field",
			err:      newErr400(`{"errMsg":"something"}`),
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "400 with different errCode",
			err:      newErr400(`{"errCode":"SFS.TURBO.0102","errMsg":"Path is not directory"}`),
			wantCode: "SFS.TURBO.0102",
			wantOk:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotOk := parseSFSTurboErrCode(tt.err)
			if gotCode != tt.wantCode || gotOk != tt.wantOk {
				t.Fatalf("parseSFSTurboErrCode() = (%q, %v), want (%q, %v)",
					gotCode, gotOk, tt.wantCode, tt.wantOk)
			}
		})
	}
}

func TestIsDirQuotaNotCreated(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "0113 dir not create quota",
			err:  newErr400(`{"errCode":"SFS.TURBO.0113","errMsg":"dir not create quota"}`),
			want: true,
		},
		{
			name: "0102 path is not directory (different code)",
			err:  newErr400(`{"errCode":"SFS.TURBO.0102","errMsg":"Path is not directory"}`),
			want: false,
		},
		{
			name: "non-HTTP error",
			err:  errSimple{},
			want: false,
		},
		{
			name: "400 unparseable body",
			err:  newErr400(`garbage`),
			want: false,
		},
		{
			name: "404 not found (not 0113)",
			err:  golangsdk.ErrDefault404{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDirQuotaNotCreated(tt.err); got != tt.want {
				t.Fatalf("isDirQuotaNotCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}

// errSimple is a plain non-HTTP error used to verify parseSFSTurboErrCode
// returns false for errors that are not golangsdk HTTP errors.
type errSimple struct{}

func (errSimple) Error() string { return "simple error" }
