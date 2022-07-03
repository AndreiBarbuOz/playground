package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	dockerRegistry "github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"           // used for docker test registry
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory" // used for docker test registry
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/crypto/bcrypt"
)

// generateCertificates create a CA certificate and use it to sign a server certificate to be used by the docker registry
// return []bytes containing the PEM encoded certificate and secret key
func generateCertificates() (*bytes.Buffer, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"UiPath"},
			Country:      []string{"US"},
			Province:     []string{""},
			Locality:     []string{"New York"},
			PostalCode:   []string{"98004"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate CA key: %w", err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create certificate: %w", err)
	}

	caPEM := &bytes.Buffer{}
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("could not encode CA certificate: %w", err)
	}

	caPrivateKeyPEM := &bytes.Buffer{}
	err = pem.Encode(caPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("could not encode CA private key: %w", err)
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"UiPath"},
			Country:      []string{"US"},
			Province:     []string{""},
			Locality:     []string{"New York"},
			PostalCode:   []string{"98004"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"www.example.com"},
	}

	certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate certificate key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create certificate: %w", err)
	}

	certPEM := &bytes.Buffer{}
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("could not encode certificate: %w", err)
	}

	certPrivateKeyPEM := &bytes.Buffer{}
	err = pem.Encode(certPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("could not encode certificate private key: %w", err)
	}

	return certPEM, certPrivateKeyPEM, nil
}

// generatePassword returns a username/password combination, encrypted with bcrypt algorithm
// the format is used by the docker registry when configuring authentication
// see https://httpd.apache.org/docs/2.4/misc/password_encryptions.html
func generatePassword(username string, password string) (*bytes.Buffer, error) {
	pwBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error generating bcrypt password for test htpasswd file: %w", err)
	}

	ret := &bytes.Buffer{}
	ret.WriteString(username)
	ret.WriteRune(':')
	ret.Write(pwBytes)

	return ret, nil
}

// RegistryServer docker registry V2 API implementation. uses the distribution/registry library behind the scenes
// contains a client which can be used to perform actions on the registry
type RegistryServer struct {
	*dockerRegistry.Registry
	RegistryURL  string
	TestUsername string
	TestPassword string
	Client       crane.Options
}

// createServer returns the docker registry server. Accepts 3 paths for the certificate, key and authentication information
func createServer(certPath string, keyPath string, passPath string) (*RegistryServer, error) {
	config := &configuration.Configuration{}

	//port, err := freeport.GetFreePort()
	//if err != nil {
	//	return nil, fmt.Errorf("error finding free port for test registry: %w", err)
	//}
	port := 30071
	config.HTTP.Addr = fmt.Sprintf("127.0.0.1:%d", port)
	config.HTTP.DrainTimeout = time.Duration(10) * time.Second
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]interface{}{}}
	config.Auth = configuration.Auth{
		"htpasswd": configuration.Parameters{
			"realm": "localhost",
			"path":  passPath,
		},
	}
	config.HTTP.TLS.Key = keyPath
	config.HTTP.TLS.Certificate = certPath
	config.Log.Level = "error"
	config.Log.AccessLog.Disabled = true

	registryURL := fmt.Sprintf("127.0.0.1:%d", port)

	r, err := dockerRegistry.NewRegistry(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker registry server: %w", err)
	}

	return &RegistryServer{
		Registry:    r,
		RegistryURL: registryURL,
	}, nil
}

// generateClient initialize the crane.Options object correctly configured to work with the docker registry
func (r *RegistryServer) generateClient() {
	var options []crane.Option
	options = append(options, crane.Insecure)

	transport := remote.DefaultTransport.Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	options = append(options, crane.WithTransport(transport))

	authorizer := &authn.Basic{
		Username: r.TestUsername,
		Password: r.TestPassword,
	}
	options = append(options, crane.WithAuth(authorizer))

	r.Client = crane.GetOptions(options...)
}

// NewRegistryServer returns a docker registry server, configured with TLS and username/password authentication
// includes a client which can be used to interact with the server
func NewRegistryServer() (*RegistryServer, error) {
	dir, _ := ioutil.TempDir("", "registry")

	cert, key, err := generateCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificates: %w", err)
	}

	htpasswd, err := generatePassword("username", "password")
	if err != nil {
		return nil, fmt.Errorf("failed to generate htpasswd: %w", err)
	}

	configFiles := []struct {
		fileName string
		content  *bytes.Buffer
	}{
		{
			fileName: "cert.pem",
			content:  cert,
		},
		{
			fileName: "key.pem",
			content:  key,
		},
		{
			fileName: "auth.htpasswd",
			content:  htpasswd,
		},
	}

	for _, f := range configFiles {
		filePath := filepath.Join(dir, f.fileName)
		contentBytes := f.content.Bytes()

		err := os.WriteFile(filePath, contentBytes, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write to file %s: %w", f.fileName, err)
		}
	}

	server, err := createServer(filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), filepath.Join(dir, "auth.htpasswd"))
	if err != nil {
		return nil, fmt.Errorf("could not create server: %w", err)
	}

	server.TestUsername = "username"
	server.TestPassword = "password"
	server.generateClient()

	return server, nil
}

func populateRegistry(server *RegistryServer, tags ...string) {
	for _, t := range tags {
		tag := fmt.Sprintf("%s/%s", server.RegistryURL, t)

		img, err := random.Image(1024, 1)
		if err != nil {
			log.Fatalf("%s", err)
		}

		ref, err := name.ParseReference(tag, server.Client.Name...)
		if err != nil {
			log.Fatalf("%s", err)
		}

		if err := remote.Write(ref, img, server.Client.Remote...); err != nil {
			log.Fatalf("%s", err)
		}
	}
}

func main() {
	srv, err := NewRegistryServer()
	if err != nil {
		panic("cannot start server")
	}
	finished := make(chan bool)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			fmt.Printf("docker registry server failed: %s", err)
		}
		finished <- true
	}()
	populateRegistry(srv, []string{"johnny:v1", "bravo:latest", "bravo:v1", "bravo:v2", "dexter"}...)

	<-finished
}
