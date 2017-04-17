package ecdsa

// Simple ECDSA supports creation of keypair,
// signing and verification. Public Key, Private Key and Signatures
// are ASN.1/DER encoded.
//
// Examples
//
// ecd := NewSimpleECDSA()
// ecd.GenerateKey()
// pubKey, _ := LoadPubKey(ecd.GetPubKey())
// fmt.Println("Private Key: ", ecd.GetPrivKey())
// signature, _ := ecd.Sign(rand.Reader, []byte("hello"))
// verified := Verify(pubKey, []byte("hello"), []byte(signature))

import (
	goecdsa "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
)

const (
	// CurveP256 P256 elliptic curve
	CurveP256 = "p256"
)

// ASN1PubKey defines a public key structure for ANS.1/DER encoding
type ASN1PubKey struct {
	X string `asn1:"utf8"`
	Y string `asn1:"utf8"`
}

// ASN1PrivKey defines a private key structure for ANS.1/DER encoding
type ASN1PrivKey struct {
	D string `asn1:"utf8"`
}

// ASN1Sig defines an elliptic signature for ANS.1/DER encoding
type ASN1Sig struct {
	R string `asn1:"utf8"`
	S string `asn1:"utf8"`
}

// SimpleECDSA defines code for creating ECDSA keys,
// signing and verifying data.
type SimpleECDSA struct {
	curve   elliptic.Curve
	privKey *goecdsa.PrivateKey
}

// NewSimpleECDSA creates a new SimpleECDSA object
func NewSimpleECDSA(curveName string) *SimpleECDSA {
	var curve elliptic.Curve
	switch curveName {
	case CurveP256:
		curve = elliptic.P256()
	default:
		panic("unsupported elliptic curve")
	}
	var se = &SimpleECDSA{curve: curve}
	if err := se.GenerateKey(); err != nil {
		panic(err)
	}
	return se
}

// Removes white spaces from a string
func clean(key string) string {
	re := regexp.MustCompile(`\r?\n|\s`)
	return re.ReplaceAllString(key, "")
}

// LoadPrivKey a formatted private key and return a ecdsa.PrivateKey
func LoadPrivKey(privKey, curveName string) (*goecdsa.PrivateKey, error) {

	privKey = clean(privKey)

	dBytes, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hex decode private key. %s", err)
	}

	var asn1PrivKey ASN1PrivKey
	_, err = asn1.Unmarshal(dBytes, &asn1PrivKey)
	if err != nil {
		return nil, errors.New("failed to unmarshal ASN.1/DER private key")
	}

	var curve elliptic.Curve
	switch curveName {
	case CurveP256:
		curve = elliptic.P256()
	default:
		return nil, errors.New("unsupported elliptic curve")
	}

	d, ok := new(big.Int).SetString(asn1PrivKey.D, 10)
	if !ok {
		return nil, errors.New("invalid private key")
	}

	return &goecdsa.PrivateKey{
		PublicKey: goecdsa.PublicKey{Curve: curve},
		D:         d,
	}, nil
}

// LoadPubKey creates a public key object and returns
// an ASN.1/DER encoded string.
func LoadPubKey(pubKey string, curveName string) (*goecdsa.PublicKey, error) {

	pubKey = clean(pubKey)

	pubBS, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hex decode public key. %s", err)
	}

	var asn1Pub ASN1PubKey
	_, err = asn1.Unmarshal(pubBS, &asn1Pub)
	if err != nil {
		return nil, errors.New("failed to unmarshal ASN.1/DER public key")
	}

	var curve elliptic.Curve
	switch curveName {
	case CurveP256:
		curve = elliptic.P256()
	default:
		return nil, errors.New("unsupported elliptic curve")
	}

	x, ok := new(big.Int).SetString(asn1Pub.X, 10)
	if !ok {
		return nil, errors.New("invalid x value")
	}
	y, ok := new(big.Int).SetString(asn1Pub.Y, 10)
	if !ok {
		return nil, errors.New("invalid y value")
	}

	return &goecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// GetPubKey encodes the public key in a DER-encoded ASN.1 data structure
// returns the hex encoded value.
func (se *SimpleECDSA) GetPubKey() string {
	asn1PubKey := ASN1PubKey{
		X: se.privKey.Public().(*goecdsa.PublicKey).X.String(),
		Y: se.privKey.Public().(*goecdsa.PublicKey).Y.String(),
	}
	bs, err := asn1.Marshal(asn1PubKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", hex.EncodeToString(bs))
}

// GetPubKeyObj returns the public key object
func (se *SimpleECDSA) GetPubKeyObj() *goecdsa.PublicKey {
	return &se.privKey.PublicKey
}

// GetPrivKey returns an ASN.1/DER encoded private key
func (se *SimpleECDSA) GetPrivKey() string {
	var asn1PrivKey = ASN1PrivKey{
		D: se.privKey.D.String(),
	}
	bs, err := asn1.Marshal(asn1PrivKey)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", hex.EncodeToString(bs))
}

// GenerateKey creates a new ECDSA private key
func (se *SimpleECDSA) GenerateKey() error {

	privKey, err := goecdsa.GenerateKey(se.curve, rand.Reader)
	if err != nil {
		return err
	}

	se.privKey = privKey
	return nil
}

// SetPrivKey sets the private key
func (se *SimpleECDSA) SetPrivKey(privKey *goecdsa.PrivateKey) {
	se.privKey = privKey
}

// Sign a byte slice. Return a ASN.1/DER-encoded signature
func (se *SimpleECDSA) Sign(rand io.Reader, hashed []byte) (string, error) {
	r, s, err := goecdsa.Sign(rand, se.privKey, hashed)
	if err != nil {
		return "", err
	}

	var asn1Sig = ASN1Sig{
		R: r.String(),
		S: s.String(),
	}

	bs, _ := asn1.Marshal(asn1Sig)

	return fmt.Sprintf("%s", hex.EncodeToString(bs)), nil
}

// IsValidPubKey checks whether the public key
// pass hex and ASN.1/DER decoding operations.
func IsValidPubKey(pubKey string) (bool, error) {
	_, err := LoadPubKey(pubKey, CurveP256)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Verify a signature. Expects a ASN.1/DER encoded signature
func Verify(pubKey *goecdsa.PublicKey, hash []byte, sig []byte) error {

	decSig, err := hex.DecodeString(string(sig))
	if err != nil {
		return fmt.Errorf("failed to hex decode signature. %s", err)
	}

	var asn1Sig ASN1Sig
	_, err = asn1.Unmarshal(decSig, &asn1Sig)
	if err != nil {
		return errors.New("failed to unmarshal ASN.1/DER signature")
	}

	r, ok := new(big.Int).SetString(asn1Sig.R, 10)
	if !ok {
		return errors.New("invalid signature r value")
	}
	s, ok := new(big.Int).SetString(asn1Sig.S, 10)
	if !ok {
		return errors.New("invalid signature s value")
	}

	// verify signature
	if goecdsa.Verify(pubKey, hash, r, s) {
		return nil
	}
	return errors.New("verification failed")
}
