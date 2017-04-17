package ecdsa

import (
	"crypto/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestSimpleECDSASpec test ../simple_ecdsa.go
func TestSimpleECDSASpec(t *testing.T) {

	Convey("SimpleECDSA", t, func() {

		Convey(".NewSimpleECDSA()", func() {

			Convey("should panic if called with an unsupported elliptic curve", func() {
				So(func() { NewSimpleECDSA("p224") }, ShouldPanicWith, "unsupported elliptic curve")
			})

			Convey("should not panic if called with a supported elliptic curve", func() {
				So(func() { NewSimpleECDSA("p256") }, ShouldNotPanic)
			})

		})

		Convey(".GetPubKey()", func() {
			Convey("should successfully return an ASN.1/DER encoded public key", func() {
				key := NewSimpleECDSA(CurveP256)
				pubFormatted := key.GetPubKey()
				So(pubFormatted, ShouldNotBeNil)
			})
		})

		Convey(".GenerateKey()", func() {
			Convey("should successfully generate a private key", func() {
				key := NewSimpleECDSA(CurveP256)
				currentKey := key.GetPrivKey()
				err := key.GenerateKey()
				So(err, ShouldBeNil)
				So(currentKey, ShouldNotEqual, key.GetPrivKey())
			})
		})

		Convey(".GetPrivKey()", func() {
			Convey("should successfully return a private key", func() {
				key := NewSimpleECDSA(CurveP256)
				privKey := key.GetPrivKey()
				So(privKey, ShouldNotBeNil)
			})
		})

		Convey(".LoadPrivKey()", func() {
			Convey("should load a ANS.1/DER encoded private key and use it to verify a signaturew", func() {
				key := NewSimpleECDSA(CurveP256)
				privKey := key.GetPrivKey()
				pubKey := key.GetPubKeyObj()
				loadedPrivKey, err := LoadPrivKey(privKey, CurveP256)
				So(err, ShouldBeNil)
				key.SetPrivKey(loadedPrivKey)
				sig, err := key.Sign(rand.Reader, []byte("hello"))
				So(err, ShouldBeNil)
				err = Verify(pubKey, []byte("hello"), []byte(sig))
				So(err, ShouldBeNil)
			})
		})

		Convey(".LoadPubKey()", func() {

			Convey("should fail if public key format is invalid", func() {
				pk, err := LoadPubKey("wrong", CurveP256)
				So(pk, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to hex decode public key. encoding/hex: odd length hex string")
			})

			Convey("should fail if elliptic curve is unsupported", func() {
				key := NewSimpleECDSA(CurveP256)
				pubFormatted := key.GetPubKey()
				pk, err := LoadPubKey(pubFormatted, "p224")
				So(pk, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unsupported elliptic curve")
			})

			Convey("should successfully load an ASN.1/DER encoded public key", func() {
				key := NewSimpleECDSA(CurveP256)
				pubFormatted := key.GetPubKey()
				pk, err := LoadPubKey(pubFormatted, CurveP256)
				So(pk, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey(".IsValidPubKey", func() {

			Convey("should fail if public key is not a valid compact/formatted key", func() {
				valid, err := IsValidPubKey("wrong")
				So(valid, ShouldEqual, false)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to hex decode public key. encoding/hex: odd length hex string")
			})

			Convey("should return success if public key if valid", func() {
				key := NewSimpleECDSA(CurveP256)
				validPubKey := key.GetPubKey()
				valid, err := IsValidPubKey(validPubKey)
				So(valid, ShouldEqual, true)
				So(err, ShouldBeNil)
			})
		})

		Convey(".Sign()", func() {

			Convey("should successfully sign text", func() {
				key := NewSimpleECDSA(CurveP256)
				s, err := key.Sign(rand.Reader, []byte("hello"))
				So(err, ShouldBeNil)
				So(s, ShouldNotBeEmpty)
			})
		})

		Convey(".Verify()", func() {

			key := NewSimpleECDSA(CurveP256)
			s, err := key.Sign(rand.Reader, []byte("hello"))
			So(err, ShouldBeNil)

			Convey("should fail if signed hash is invalid hex value", func() {
				key := NewSimpleECDSA(CurveP256)
				pubKey := key.privKey.PublicKey
				err := Verify(&pubKey, []byte("wrong"), []byte("wrong"))
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to hex decode signature. encoding/hex: odd length hex string")
			})

			Convey("should fail if signature could not be verified", func() {
				key := NewSimpleECDSA(CurveP256)
				pubKey := key.privKey.PublicKey
				err := Verify(&pubKey, []byte("hi"), []byte(s))
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "verification failed")
			})

			Convey("should successfully verify signature", func() {
				pubKey := key.privKey.PublicKey
				err := Verify(&pubKey, []byte("hello"), []byte(s))
				So(err, ShouldBeNil)
			})
		})
	})
}
