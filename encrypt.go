package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var asd = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

var SecretString string = getConfig().Secretstring

func encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func encrypt(text, s string) (string, error) {
	block, err := aes.NewCipher([]byte(s))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, asd)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return encode(cipherText), nil
}

func decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func decrypt(text, s string) (string, error) {
	block, err := aes.NewCipher([]byte(s))
	if err != nil {
		return "", err
	}
	cipherText := decode(text)
	cfb := cipher.NewCFBDecrypter(block, asd)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
