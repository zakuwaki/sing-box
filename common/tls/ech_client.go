//go:build with_ech

package tls

import (
	"crypto/x509"
	"net"
	"net/netip"
	"os"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/cloudflaretls"
	E "github.com/sagernet/sing/common/exceptions"
	"encoding/base64"
)

type echClientConfig struct {
	config *tls.Config
}

func (e *echClientConfig) Config() (*STDConfig, error) {
	return nil, E.New("unsupported usage for ECH")
}

func (e *echClientConfig) Client(conn net.Conn) Conn {
	return tls.Client(conn, e.config)
}

func newECHClient(serverAddress string, options option.OutboundTLSOptions) (Config, error) {
	var serverName string
	if options.ServerName != "" {
		serverName = options.ServerName
	} else if serverAddress != "" {
		if _, err := netip.ParseAddr(serverName); err != nil {
			serverName = serverAddress
		}
	}
	if serverName == "" && !options.Insecure {
		return nil, E.New("missing server_name or insecure=true")
	}

	var tlsConfig tls.Config
	if options.DisableSNI {
		tlsConfig.ServerName = "127.0.0.1"
	} else {
		tlsConfig.ServerName = serverName
	}
	if options.Insecure {
		tlsConfig.InsecureSkipVerify = options.Insecure
	} else if options.DisableSNI {
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			verifyOptions := x509.VerifyOptions{
				DNSName:       serverName,
				Intermediates: x509.NewCertPool(),
			}
			for _, cert := range state.PeerCertificates[1:] {
				verifyOptions.Intermediates.AddCert(cert)
			}
			_, err := state.PeerCertificates[0].Verify(verifyOptions)
			return err
		}
	}
	if len(options.ALPN) > 0 {
		tlsConfig.NextProtos = options.ALPN
	}
	if options.MinVersion != "" {
		minVersion, err := ParseTLSVersion(options.MinVersion)
		if err != nil {
			return nil, E.Cause(err, "parse min_version")
		}
		tlsConfig.MinVersion = minVersion
	}
	if options.MaxVersion != "" {
		maxVersion, err := ParseTLSVersion(options.MaxVersion)
		if err != nil {
			return nil, E.Cause(err, "parse max_version")
		}
		tlsConfig.MaxVersion = maxVersion
	}
	if options.CipherSuites != nil {
	find:
		for _, cipherSuite := range options.CipherSuites {
			for _, tlsCipherSuite := range tls.CipherSuites() {
				if cipherSuite == tlsCipherSuite.Name {
					tlsConfig.CipherSuites = append(tlsConfig.CipherSuites, tlsCipherSuite.ID)
					continue find
				}
			}
			return nil, E.New("unknown cipher_suite: ", cipherSuite)
		}
	}
	var certificate []byte
	if options.Certificate != "" {
		certificate = []byte(options.Certificate)
	} else if options.CertificatePath != "" {
		content, err := os.ReadFile(options.CertificatePath)
		if err != nil {
			return nil, E.Cause(err, "read certificate")
		}
		certificate = content
	}
	if len(certificate) > 0 {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(certificate) {
			return nil, E.New("failed to parse certificate:\n\n", certificate)
		}
		tlsConfig.RootCAs = certPool
	}

	// ECH Config

	tlsConfig.ECHEnabled = true
	tlsConfig.PQSignatureSchemesEnabled = options.ECH.PQSignatureSchemesEnabled
	tlsConfig.DynamicRecordSizingDisabled = options.ECH.DynamicRecordSizingDisabled
	if options.ECH.Config != "" {
		clientConfigContent, err := base64.StdEncoding.DecodeString(options.ECH.Config)
		if err != nil {
			return nil, err
		}
		clientConfig, err := tls.UnmarshalECHConfigs(clientConfigContent)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientECHConfigs = clientConfig
	}
	return &echClientConfig{&tlsConfig}, nil
}
