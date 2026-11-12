package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeCA(t *testing.T) string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test root"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTLSConfig(t *testing.T) {
	notPEM := filepath.Join(t.TempDir(), "not.pem")
	if err := os.WriteFile(notPEM, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	ca := writeCA(t)

	cases := []struct {
		name     string
		insecure bool
		caFile   string
		wantErr  bool
		wantPool bool
	}{
		{name: "default"},
		{name: "insecure", insecure: true},
		{name: "ca-file", caFile: ca, wantPool: true},
		{name: "insecure wins over ca-file", insecure: true, caFile: ca},
		{name: "missing ca-file", caFile: filepath.Join(t.TempDir(), "none.crt"), wantErr: true},
		{name: "ca-file without certificates", caFile: notPEM, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &CloudCredentials{}
			c.Global.Insecure = tc.insecure
			c.Global.CAFile = tc.caFile
			cfg, err := c.tlsConfig()
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, want error %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if cfg.InsecureSkipVerify != tc.insecure {
				t.Fatalf("InsecureSkipVerify = %v", cfg.InsecureSkipVerify)
			}
			if (cfg.RootCAs != nil) != tc.wantPool {
				t.Fatalf("RootCAs set = %v, want %v", cfg.RootCAs != nil, tc.wantPool)
			}
		})
	}
}
