package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"sync"
)

var mu sync.Mutex

func GetRSAPairPaths() (public string, private string, err error) {
	// protect concurrent tests from clashing with each other
	mu.Lock()
	defer mu.Unlock()

	var (
		dirPath = path.Join(os.TempDir(), "occa_keys")

		publicPath  = path.Join(dirPath, "key_pub")
		privatePath = path.Join(dirPath, "key")
	)

	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("creating key directory: %w", err)
	}

	defer func() {
		if err != nil {
			_ = os.Remove(dirPath)
		}
	}()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generating key: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("marshalling private key bytes: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err = os.WriteFile(privatePath, privateKeyPEM, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("writing private key: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("converting public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err = os.WriteFile(publicPath, publicKeyPEM, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("writing public key: %w", err)
	}

	return publicPath, privatePath, nil
}
