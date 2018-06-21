package jwt_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"testing"

	"git.lcgc.work/platform/kelp/jwt"
)

const (
	publicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCT6a6va/eQkilmTL04ZCsqp1Be
Yv8/0SFPWGJpyHFaIRAsgzMIJhHvs472ITCCkTLlWTJsBGaYvSS/9p4mKitKYoQl
UZD5rxHe+Mo1lADhvpXIyfcoCok77/d4wGfu9UYUjjcMRfgP6Ubd5xPdoSQ6Y+Ag
uRoicaWVel3yQe2C5QIDAQAB
-----END PUBLIC KEY-----`
	privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCT6a6va/eQkilmTL04ZCsqp1BeYv8/0SFPWGJpyHFaIRAsgzMI
JhHvs472ITCCkTLlWTJsBGaYvSS/9p4mKitKYoQlUZD5rxHe+Mo1lADhvpXIyfco
Cok77/d4wGfu9UYUjjcMRfgP6Ubd5xPdoSQ6Y+AguRoicaWVel3yQe2C5QIDAQAB
AoGAS5iYXjHJMotxO0anQuX3JfKjYcoo+J3S73GVwsjeHhld4dIgR97lNeRIgcUO
vPymzY8b9Rb+tLuex7mstIbC3pV1k8k1JPMeWhtRlL8XB1IGnQ+J+G75qhXoxfJs
Vwch4EPSdYcC3sazNDZES6sOk0kbHKjBNKp6xAN2aiXvqYECQQCnX05Pi/BR4anj
1tWIVn0tun22CLKA1J7305+a6CQpiD2whWyKZ7ixC05XXjH7ovCte/w94Z3K7Jeu
gES+GPnFAkEA4jx8OBoB8NMTldvXk23ALHG8b8i0CNpDA+dROxC6sjzifkI3yQMp
fsrX77N0XMvEBGBqH+9D3VIHmbmbK1OWoQJAJ1MHLT1t228v4W5KgVwA0Uut2aU1
e3t5VjUMnPLJ/FAcXmsa03IHPOGxKGnjSIJCFiC5ZTirQFBSWxecDyYDGQJBAIgH
8nERQkV5xXcAPzlIbprKsJFOTuJbypvYRMGNJ5TwgweD7WMSYar5cKmSb880PmKb
TnRxO48iOau3LJP6qMECQCkAA5dm43xhEg7i+inKJ6iuHeCcgEK2TwsEFZyiOho8
bJB4oo5RIOFbasJJEyKtXe+KxojnBwzW8EJs7RCyqqA=
-----END RSA PRIVATE KEY-----`
)

func TestJWT(t *testing.T) {
	block, _ := pem.Decode([]byte(privateKey))
	priv, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	block, _ = pem.Decode([]byte(publicKey))
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	j := jwt.New(jwt.RS256(priv, pub.(*rsa.PublicKey)))

	// set claims
	j.Claims.SetSubject("sub")

	// sign
	token, err := j.Sign()
	if err == nil {
		fmt.Printf("sign token is %s\n", token)
	} else {
		fmt.Printf("sign error %v", err)
	}

	// verify
	if err := j.Verify(token); err == nil {
		fmt.Printf("verify pass\n")
	} else {
		fmt.Printf("verify error %v", err)
	}

	if err := j.Parse(token); err == nil {
		ret, _ := json.Marshal(j)
		fmt.Printf("claims is %v\n", string(ret))
	} else {
		fmt.Printf("parse error %v", err)
	}
}
