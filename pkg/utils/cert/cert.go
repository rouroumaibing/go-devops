package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// 解析CA证书
func NewPool(filename string) (*x509.CertPool, error) {
	if len(filename) == 0 {
		return nil, fmt.Errorf("Read From an empty filename!")
	}
	pemCerts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	certs, err := ParseCertsPem(pemCerts)
	if err != nil {
		return nil, fmt.Errorf("CA cert read faild: %s: %s", filename, err)
	}

	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}

	return pool, nil
}

// 解析证书数组
func ParseCertsPem(pemCerts []byte) ([]*x509.Certificate, error) {
	ok := false
	certs := []*x509.Certificate{}
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return certs, err
		}
		certs = append(certs, cert)
		ok = true
	}

	if !ok {
		return certs, fmt.Errorf("cloud nor read any cert")
	}

	return certs, nil
}
