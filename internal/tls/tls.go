package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func GenerateCertificate(certFile, keyFile string) error {

	// Comprobar si ya existen los dos archivos.
	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)

	if certErr == nil && keyErr == nil {
		return nil
	}

	// Crear el directorio si no existe.
	dir := filepath.Dir(certFile)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creando directorio del certificado: %w", err)
	}

	// Generar clave privada RSA.
	privateKey, err := rsa.GenerateKey(
		rand.Reader,
		2048,
	)

	if err != nil {
		return fmt.Errorf("error generando clave privada: %w", err)
	}

	// Crear número de serie aleatorio.
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(
		rand.Reader,
		serialLimit,
	)

	if err != nil {
		return fmt.Errorf("error generando número de serie: %w", err)
	}

	// Definir el certificado.
	template := x509.Certificate{
		SerialNumber: serialNumber,

		Subject: pkix.Name{
			CommonName: "GoCockpit",
		},

		NotBefore: time.Now(),

		NotAfter: time.Now().Add(
			365 * 24 * time.Hour,
		),

		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment,

		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},

		BasicConstraintsValid: true,

		DNSNames: []string{
			"localhost",
		},

		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}

	// Crear el certificado autofirmado.
	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)

	if err != nil {
		return fmt.Errorf("error creando certificado: %w", err)
	}

	// Crear archivo del certificado.
	cert, err := os.OpenFile(
		certFile,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)

	if err != nil {
		return fmt.Errorf("error creando certificado: %w", err)
	}

	defer cert.Close()

	err = pem.Encode(
		cert,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		},
	)

	if err != nil {
		return fmt.Errorf("error escribiendo certificado: %w", err)
	}

	// Convertir la clave privada a PKCS#1.
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// Crear archivo de la clave privada.
	key, err := os.OpenFile(
		keyFile,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)

	if err != nil {
		return fmt.Errorf("error creando clave privada: %w", err)
	}

	defer key.Close()

	err = pem.Encode(
		key,
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyDER,
		},
	)

	if err != nil {
		return fmt.Errorf("error escribiendo clave privada: %w", err)
	}

	return nil
}
