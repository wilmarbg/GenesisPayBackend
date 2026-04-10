package payments

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

func Encrypt(text, keyString string) ([]byte, error) {
	key := []byte(keyString)
	if len(key) != 32 {
		return nil, errors.New("la clave debe tener 32 bytes de longitud")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte, keyString string) (string, error) {
	key := []byte(keyString)
	if len(key) != 32 {
		return "", errors.New("la clave debe tener 32 bytes de longitud")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("texto cifrado demasiado corto")
	}

	nonce, ciphertextMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertextMessage, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func GenerateCardNumber() string {
	number := ""
	for len(number) < 16 {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		digit := int(randomByte[0]) % 10
		number += fmt.Sprintf("%d", digit)
	}
	return number
}

func MaskCardNumber(number string) string {
	if len(number) < 4 {
		return number
	}
	last4 := number[len(number)-4:]
	return "**** **** **** " + last4
}
