package libs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"log"
)

// GenerateSSHKeys creates SSH Keys for LabRequest
func GenerateSSHKeys(uuid string) (publickey []byte, privatekey []byte){
	// TODO: #1 instead of saving key to local file create OpenShift/K8s secret
	//PrivateKeyFile := "/tmp/" + uuid
	//PublicKeyFile := "/tmp/" + uuid + ".pub"
	keyBitSize := 4096

	generatedPrivateKey, err := generatePrivateKey(keyBitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	extractedPublicKeyBytes, err := generatePublicKey(&generatedPrivateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	generatedPrivateKeyBytes := encodePrivateKeyToPEM(generatedPrivateKey)

	// TODO: linked to #1
	//err = writeKeyToFile(generatedPrivateKeyBytes, PrivateKeyFile)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//
	//err = writeKeyToFile([]byte(extractedPublicKeyBytes), PublicKeyFile)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}

	return extractedPublicKeyBytes, generatedPrivateKeyBytes
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(keyBitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	generatedPrivateKey, err := rsa.GenerateKey(rand.Reader, keyBitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = generatedPrivateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return generatedPrivateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(generatedPrivateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privateKeyDERFormat := x509.MarshalPKCS1PrivateKey(generatedPrivateKey)

	// pem.Block
	privateKeyPEMBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDERFormat,
	}

	// Private key in PEM format
	privateKeyPEMFormat := pem.EncodeToMemory(&privateKeyPEMBlock)

	return privateKeyPEMFormat
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(generatedPrivateKey *rsa.PublicKey) ([]byte, error) {
	extractedPublicKey, err := ssh.NewPublicKey(generatedPrivateKey)
	if err != nil {
		return nil, err
	}

	extractedPublicKeyBytes := ssh.MarshalAuthorizedKey(extractedPublicKey)

	log.Println("Public key generated")
	return extractedPublicKeyBytes, nil
}

// TODO: linked to #1
// writePemToFile writes keys to a file
//func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
//	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
//	if err != nil {
//		return err
//	}
//
//	log.Printf("Key saved to: %s", saveFileTo)
//	return nil
//}
