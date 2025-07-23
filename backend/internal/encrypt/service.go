package encrypt


//Encryptor service interface
type EncryptorService interface {
	EncryptAccessToken(plaintext []byte, keyString string) (string, error)
	DecryptAccessToken(ciphertext, keyString string) ([]byte, error)
}