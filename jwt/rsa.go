package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
)

// rsasha implement Alg interface
type rsasha struct {
	name       string
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	hash       crypto.Hash
}

type RsaKeySet struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey

	Kty string `json:"kty"`
	Kid string `json:"kid"`
	E   string `json:"e"`
	N   string `json:"n"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

func NewRsaKeySet(kid, publicKey, privateKey, alg string) (*RsaKeySet, error) {
	pub, err := RSAPublicKeyFromString(publicKey)
	if err != nil {
		return nil, err
	}
	priv, err := RSAPrivateKeyFromString(privateKey)
	if err != nil {
		return nil, err
	}
	ks := &RsaKeySet{
		publicKey:  pub,
		privateKey: priv,

		Kty: "RSA",
		Kid: kid,
		E:   base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		N:   base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(pub.N.Bytes()),
		Use: "sig",
		Alg: alg,
	}
	return ks, nil
}

func (this *RsaKeySet) Pair() (*rsa.PrivateKey, *rsa.PublicKey) {
	return this.privateKey, this.publicKey
}

// RSAPublicKeyFromString build a rsa.PublicKey from string
func RSAPublicKeyFromString(str string) (*rsa.PublicKey, error) {
	block, rest := pem.Decode([]byte(str))
	if block == nil {
		return nil, fmt.Errorf("key pem decode faild: %s", rest)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}

// RSAPrivateKeyFromString build a rsa.PrivateKey from string
func RSAPrivateKeyFromString(str string) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode([]byte(str))
	if block == nil {
		return nil, fmt.Errorf("key pem decode faild: %s", rest)
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// RS256 is an crypto algorithm using RSA and SHA-256
func RS256(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Alg {
	return &rsasha{
		"RS256",
		publicKey,
		privateKey,
		crypto.SHA256,
	}
}

// RS384 is an crypto algorithm using RSA and SHA-384
func RS384(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Alg {
	return &rsasha{
		"RS384",
		publicKey,
		privateKey,
		crypto.SHA384,
	}
}

// RS512 is an crypto algorithm using RSA and SHA-512
func RS512(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Alg {
	return &rsasha{
		"RS512",
		publicKey,
		privateKey,
		crypto.SHA512,
	}
}

func (this *rsasha) Name() string {
	return this.name
}

func (this *rsasha) Sign(data []byte) ([]byte, error) {
	if this.privateKey == nil {
		return nil, ErrorInvalidPrivateKey
	}

	h := this.hash.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}

	sign, err := rsa.SignPKCS1v15(rand.Reader, this.privateKey, this.hash, h.Sum(nil))
	if err != nil {
		return nil, err
	}
	return sign, nil

}

func (this *rsasha) Verify(data, sign []byte) error {
	if this.publicKey == nil {
		return ErrorInvalidPublicKey
	}
	h := this.hash.New()
	if _, err := h.Write(data); err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(this.publicKey, this.hash, h.Sum(nil), sign)
}
