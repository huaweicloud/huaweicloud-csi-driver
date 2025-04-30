package transport

import (
	"net/http"
	"testing"

	"github.com/tjfoc/gmsm/gmtls"
	"github.com/tjfoc/gmsm/x509"
)

func TestNewGMRoundTripper(t *testing.T) {
	tests := []struct {
		name        string
		description string
		cfg         *gmtls.Config
		request     *http.Request
		expected    bool
	}{
		{
			name:        "test1",
			description: "support both GM and no-GM encryption algorithm",
			cfg: &gmtls.Config{
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://sm2test.ovssl.cn", nil, t),
			expected: false,
		},
		{
			name:        "test2",
			description: "support both GM and no-GM encryption algorithm",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeAutoSwitch},
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://sm2test.ovssl.cn", nil, t),
			expected: false,
		},
		{
			name:        "test3",
			description: "support both GM and no-GM encryption algorithm",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeGMSSLOnly},
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://sm2test.ovssl.cn", nil, t),
			expected: false,
		},
		{
			name:        "test4",
			description: "support both GM and no-GM encryption algorithm",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeAutoSwitch},
				InsecureSkipVerify: true,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					for _, v := range rawCerts {
						_, err := x509.ParseCertificate(v)
						if err != nil {
							return err
						}
					}
					return nil
				},
			},
			request:  httpNewRequest("GET", "https://sm2test.ovssl.cn", nil, t),
			expected: false,
		},
		{
			name:        "test5",
			description: "support both GM and no-GM encryption algorithm",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeGMSSLOnly},
				InsecureSkipVerify: true,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					for _, v := range rawCerts {
						_, err := x509.ParseCertificate(v)
						if err != nil {
							return err
						}
					}
					return nil
				},
			},
			request:  httpNewRequest("GET", "https://sm2test.ovssl.cn", nil, t),
			expected: false,
		},

		{
			name:        "test6",
			description: "don not support GM",
			cfg: &gmtls.Config{
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://baidu.com", nil, t),
			expected: true,
		},
		{
			name:        "test7",
			description: "don not support GM",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeAutoSwitch},
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://baidu.com", nil, t),
			expected: true,
		},
		{
			name:        "test8",
			description: "don not support GM",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeGMSSLOnly},
				InsecureSkipVerify: true,
			},
			request:  httpNewRequest("GET", "https://baidu.com", nil, t),
			expected: true,
		},
		{
			name:        "test9",
			description: "don not support GM",
			cfg: &gmtls.Config{
				GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeAutoSwitch},
				InsecureSkipVerify: true,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					for _, v := range rawCerts {
						_, err := x509.ParseCertificate(v)
						if err != nil {
							return err
						}
					}
					return nil
				},
			},
			request:  httpNewRequest("GET", "https://baidu.com", nil, t),
			expected: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			wrapper := NewGMRoundTripper(testCase.cfg)
			wrappedTransport := wrapper(http.DefaultTransport)
			client := &http.Client{
				Transport: wrappedTransport,
			}

			resp, err := client.Do(testCase.request)

			if testCase.expected && err == nil {
				t.Errorf("expected error but got none")
			}
			if !testCase.expected && err != nil {
				t.Errorf("expected: %v, but got err: %v", testCase.expected, err)
			}

			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}
