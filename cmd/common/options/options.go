package options

import (
	"github.com/spf13/pflag"
)

type CommonOptions struct {
	TLSCAFile           string
	TLSCertFile         string
	TLSKeyFile          string
	MaxRequestBodyBytes int // default is 2MB.
}

func (common *CommonOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&common.TLSCAFile, "tls-ca-file", "/source/certs/ca.crt",
		"Clien certificate, This must be a valid PEM-encoded CA bundle")
	fs.StringVar(&common.TLSCertFile, "tls-cert-file", "/source/certs/server.crt",
		"Server Cert. File containing the default X509 Certificate for HTTPS.")
	fs.StringVar(&common.TLSKeyFile, "tls-key-file", "/source/certs/server.key",
		"Server Cert Key. File containing the default X509 private key matching --tls-cert-file.")
	fs.IntVar(&common.MaxRequestBodyBytes, "max-request-body-bytes", 2097152,
		"Max request body length, default is 2MB.")
}
