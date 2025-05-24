package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"time"
)

func VerifySignature(message string, signatureBase64 string, pubKey *rsa.PublicKey) error {
	log.Println(message)
	hashed := sha256.Sum256([]byte(message))
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return err
	}

	return nil
}

func ValidateTimestamp(ts string) error {
	clientTime, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return errors.New(("invalid timestamp format"))
	}

	diff := time.Since(clientTime)
	if diff > 5*time.Minute || diff < -5*time.Minute {
		return errors.New("timestamp expired or invalid")
	}

	return nil
}
