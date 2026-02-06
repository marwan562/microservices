package tenant

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// KeyManager manages tenant-specific encryption keys for CMK (Customer Managed Keys).
type KeyManager struct {
	masterKey []byte
	keyStore  KeyStore
}

// KeyStore is the interface for storing and retrieving tenant keys.
type KeyStore interface {
	// GetKey retrieves the encryption key for a tenant.
	GetKey(tenantID string) ([]byte, error)

	// SetKey stores the encryption key for a tenant.
	SetKey(tenantID string, key []byte) error

	// RotateKey generates and stores a new key for a tenant.
	RotateKey(tenantID string) ([]byte, error)

	// DeleteKey removes the key for a tenant.
	DeleteKey(tenantID string) error
}

// NewKeyManager creates a new key manager with the master encryption key.
func NewKeyManager(masterKey []byte, store KeyStore) (*KeyManager, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes for AES-256")
	}

	return &KeyManager{
		masterKey: masterKey,
		keyStore:  store,
	}, nil
}

// GenerateTenantKey generates a new AES-256 key for a tenant.
func (km *KeyManager) GenerateTenantKey(tenantID string) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Encrypt the tenant key with master key before storing
	encryptedKey, err := km.encryptWithMaster(key)
	if err != nil {
		return nil, err
	}

	if err := km.keyStore.SetKey(tenantID, encryptedKey); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	return key, nil
}

// GetTenantKey retrieves and decrypts the key for a tenant.
func (km *KeyManager) GetTenantKey(tenantID string) ([]byte, error) {
	encryptedKey, err := km.keyStore.GetKey(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	return km.decryptWithMaster(encryptedKey)
}

// Encrypt encrypts data using the tenant's key.
func (km *KeyManager) Encrypt(tenantID string, plaintext []byte) (string, error) {
	key, err := km.GetTenantKey(tenantID)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using the tenant's key.
func (km *KeyManager) Decrypt(tenantID string, ciphertext string) ([]byte, error) {
	key, err := km.GetTenantKey(tenantID)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, cipherData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, cipherData, nil)
}

// RotateTenantKey rotates the encryption key for a tenant.
func (km *KeyManager) RotateTenantKey(tenantID string) ([]byte, error) {
	return km.GenerateTenantKey(tenantID)
}

func (km *KeyManager) encryptWithMaster(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(km.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (km *KeyManager) decryptWithMaster(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(km.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, data := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, data, nil)
}

// InMemoryKeyStore is a simple in-memory key store for testing.
type InMemoryKeyStore struct {
	keys map[string][]byte
}

// NewInMemoryKeyStore creates a new in-memory key store.
func NewInMemoryKeyStore() *InMemoryKeyStore {
	return &InMemoryKeyStore{
		keys: make(map[string][]byte),
	}
}

// GetKey retrieves a key from memory.
func (s *InMemoryKeyStore) GetKey(tenantID string) ([]byte, error) {
	key, ok := s.keys[tenantID]
	if !ok {
		return nil, fmt.Errorf("key not found for tenant %s", tenantID)
	}
	return key, nil
}

// SetKey stores a key in memory.
func (s *InMemoryKeyStore) SetKey(tenantID string, key []byte) error {
	s.keys[tenantID] = key
	return nil
}

// RotateKey generates a new key.
func (s *InMemoryKeyStore) RotateKey(tenantID string) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	s.keys[tenantID] = key
	return key, nil
}

// DeleteKey removes a key from memory.
func (s *InMemoryKeyStore) DeleteKey(tenantID string) error {
	delete(s.keys, tenantID)
	return nil
}
