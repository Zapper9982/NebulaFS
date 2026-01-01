package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	data := []byte("Hello, NebulaFS!")

	encrypted, err := EncryptAES256(data, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptAES256(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(data, decrypted) {
		t.Fatalf("Decrypted data mismatch. Expected %s, got %s", data, decrypted)
	}
}

func TestHashSHA1(t *testing.T) {
	data := []byte("NebulaFS")
	expected := "954a90dcbb7e33e1e7661730b55ef050dcd3b7b7" // python3 -c "import hashlib; print(hashlib.sha1(b'NebulaFS').hexdigest())"

	hash := HashSHA1(data)
	if hash != expected {
		t.Errorf("Hash mismatch. Expected %s, got %s", expected, hash)
	}
}
