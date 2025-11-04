package utils

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/astaxie/beego"

	"github.com/rouroumaibing/go-devops/cmd/common/options"
	"github.com/rouroumaibing/go-devops/pkg/utils/cert"
)

func InitServerSecureConfig(bCfg beego.Config, opt options.CommonOptions) error {
	beego.BConfig.Listen.EnableHTTPS = bCfg.Listen.EnableHTTPS
	beego.BConfig.Listen.HTTPSPort = bCfg.Listen.HTTPSPort
	beego.BConfig.Listen.HTTPSAddr = bCfg.Listen.HTTPSAddr
	beego.BConfig.Listen.EnableHTTP = bCfg.Listen.EnableHTTP
	beego.BConfig.Listen.EnableAdmin = bCfg.Listen.EnableAdmin

	if beego.BConfig.Listen.EnableHTTPS {
		beego.BeeApp.Server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
		}

		if len(opt.TLSCertFile) != 0 && len(opt.TLSKeyFile) != 0 {
			keyBuf, err := os.ReadFile(opt.TLSKeyFile)
			if err != nil {
				return err
			}
			certBuf, err := os.ReadFile(opt.TLSCertFile)
			if err != nil {
				return err
			}
			certi, err := tls.X509KeyPair(certBuf, []byte(keyBuf))
			if err != nil {
				return fmt.Errorf("Unable to load server certificate: %v", err)
			}
			beego.BeeApp.Server.TLSConfig.Certificates = append(beego.BeeApp.Server.TLSConfig.Certificates, certi)
			beego.BeeApp.Server.TLSConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &beego.BeeApp.Server.TLSConfig.Certificates[0], nil
			}

		}

		if len(opt.TLSCAFile) > 0 {
			clientCAs, err := cert.NewPool(opt.TLSCAFile)
			if err != nil {
				return fmt.Errorf("Unable to load client CA file: %v", err)
			}
			beego.BeeApp.Server.TLSConfig.ClientAuth = tls.RequestClientCert
			beego.BeeApp.Server.TLSConfig.ClientCAs = clientCAs
		}
	}
	return nil
}
