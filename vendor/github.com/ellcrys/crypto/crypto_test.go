package crypto

import "testing"
import "github.com/stretchr/testify/assert"

var keys = []string{
	"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCroZieOAo9stcf6R6eWfo51VCv\nK8cLdNS577m/HIFOmEd1CDi/u7agGzpehNAhHpr5NVjQZ4Te+KMRn9SnpUK2hc8d\nUU25PQolsOEwePVQ18hHNK4Y2JvOY/f8KCO2hhrS6uuP6eedpnSdulS1OXHTL6Zx\nQmBd9F33gLT6BERHQwIDAQAB\n-----END PUBLIC KEY-----",
	"-----BEGIN KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCroZieOAo9stcf6R6eWfo51VCv\nK8cLdNS577m/HIFOmEd1CDi/u7agGzpehNAhHpr5NVjQZ4Te+KMRn9SnpUK2hc8d\nUU25PQolsOEwePVQ18hHNK4Y2JvOY/f8KCO2hhrS6uuP6eedpnSdulS1OXHTL6Zx\nQmBd9F33gLT6BERHQwIDAQAB\n-----END PUBLIC KEY-----",
	"-----BEGIN RSA PRIVATE KEY-----\nMIICWwIBAAKBgQCroZieOAo9stcf6R6eWfo51VCvK8cLdNS577m/HIFOmEd1CDi/\nu7agGzpehNAhHpr5NVjQZ4Te+KMRn9SnpUK2hc8dUU25PQolsOEwePVQ18hHNK4Y\n2JvOY/f8KCO2hhrS6uuP6eedpnSdulS1OXHTL6ZxQmBd9F33gLT6BERHQwIDAQAB\nAoGAEZ/0ljrXAmL9KG++DzDaO1omgPaT6B9FQRrXDkMVHEcS/3eqrDXQmTxykAY/\ngUctTu4lgrE+uc76n/Kz2ctkwEKIKet56ylqp+wlEUt1G+udoi07tgd7XyxzoUJm\nZwSm89gKh+mEPxni0FrBNg6dR0n2gvKRecnXqyoGVOHZITECQQDXgRJyrzgc/JhB\nSOBznEjtXAZXRRu3o9UznztjU9Xz7NWXTVuHu8WqYmGWCOqnysMhXJ3xBddJyDTF\njuOJ0123AkEAy+H+3POcT2FDOuluqPmAZQAUU6Nxtbj02/JJtOy7jq5jnN27HVC3\nuQzmfsS5J2XeQQodOUwOy2Ub57/OMrMi1QJAGZsZgQz2wuL0iFVLbhE0zRcxHa91\ncqWB0Kdr3Ap7EoeifV7QsFkMTIlyBOy8TQGXm+AwWBIUmYyzUIIA4UB/EwJAO+Bo\nSB2nZ0yqQO/zVt7HjWIDljinGXZzOvEiImdwAcxHZvdbj5V4D3mxa8N8mQx6xGEj\nCgPDSIquMlaLSSqA7QJAAbQPa0frCkm1rkWWZ7QwGm7ptzOACwFEGefm/1mhmw3a\nvoWRTHhrDuEbeVH3iF8MWhLJLPFtuSShiQMsrVbXPA==\n-----END RSA PRIVATE KEY-----",
	"-----BEGIN PRIVATE KEY-----\nMIICWwIBAAKBgQCroZieOAo9stcf6R6eWfo51VCvK8cLdNS577m/HIFOmEd1CDi/\nu7agGzpehNAhHpr5NVjQZ4Te+KMRn9SnpUK2hc8dUU25PQolsOEwePVQ18hHNK4Y\n2JvOY/f8KCO2hhrS6uuP6eedpnSdulS1OXHTL6ZxQmBd9F33gLT6BERHQwIDAQAB\nAoGAEZ/0ljrXAmL9KG++DzDaO1omgPaT6B9FQRrXDkMVHEcS/3eqrDXQmTxykAY/\ngUctTu4lgrE+uc76n/Kz2ctkwEKIKet56ylqp+wlEUt1G+udoi07tgd7XyxzoUJm\nZwSm89gKh+mEPxni0FrBNg6dR0n2gvKRecnXqyoGVOHZITECQQDXgRJyrzgc/JhB\nSOBznEjtXAZXRRu3o9UznztjU9Xz7NWXTVuHu8WqYmGWCOqnysMhXJ3xBddJyDTF\njuOJ0123AkEAy+H+3POcT2FDOuluqPmAZQAUU6Nxtbj02/JJtOy7jq5jnN27HVC3\nuQzmfsS5J2XeQQodOUwOy2Ub57/OMrMi1QJAGZsZgQz2wuL0iFVLbhE0zRcxHa91\ncqWB0Kdr3Ap7EoeifV7QsFkMTIlyBOy8TQGXm+AwWBIUmYyzUIIA4UB/EwJAO+Bo\nSB2nZ0yqQO/zVt7HjWIDljinGXZzOvEiImdwAcxHZvdbj5V4D3mxa8N8mQx6xGEj\nCgPDSIquMlaLSSqA7QJAAbQPa0frCkm1rkWWZ7QwGm7ptzOACwFEGefm/1mhmw3a\nvoWRTHhrDuEbeVH3iF8MWhLJLPFtuSShiQMsrVbXPA==\n-----END RSA PRIVATE KEY-----",
}

// TestParsePublicKey tests that public key is valid
func TestParseGoodPublicKey(t *testing.T) {
	pubKey := keys[0]
	_, err := ParsePublicKey([]byte(pubKey))
	assert.Nil(t, err)
}

// TestUnsupportedPublicKeyType tests that a public key having an unsupported key type will not be parsed
func TestUnsupportedPublicKeyType(t *testing.T) {
	pubKey := keys[1]
	_, err := ParsePublicKey([]byte(pubKey))
	expectedMsg := `unsupported key type "KEY"`
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedMsg)
}

// TestParseGoodPrivateKey tests that a private key is valid
func TestParseGoodPrivateKey(t *testing.T) {
	key := keys[2]
	_, err := ParsePrivateKey([]byte(key))
	assert.Nil(t, err)
}

// TestUnsupportedPrivateKeyType tests that a private key having an unsupported key type will not be parsed
func TestUnsupportedPrivateKeyType(t *testing.T) {
	key := keys[3]
	_, err := ParsePublicKey([]byte(key))
	expectedMsg := `unsupported key type "PRIVATE KEY"`
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedMsg)
}

// TestSignWithPrivateKey test that a valid private key will successfully sign a string
func TestSignWithPrivateKey(t *testing.T) {
	key := keys[2]
	signer, err := ParsePrivateKey([]byte(key))
	assert.Nil(t, err)
	signature, err := signer.Sign([]byte("hello"))
	if !assert.Nil(t, err) {
		t.Error("unable to sign:", err)
	} else {
		expectedSignature := "a527f1e81f65a00e06ed09434d58ae54ec3acb6f097d1fce8d60781b157c0186da0a0dbefc8cceea5df77b2f95658d94f22fb2641eeb33674ebd85e472f65f2bb1243f2ea1d2d4b6cb20b60c77371eee3fe01227e2ccae1f7bb957d54814d1d9ceefd5b789b57fd10da69961d78e5e60a55326de185f51edcb5bf05bfa6c828b"
		if !assert.Equal(t, signature, expectedSignature) {
			t.Errorf("should match expected hex string")
		}
	}
}

// TestVerifyWithPublicKey tests that a valid public key will verify a signature
func TestVerifyWithPublicKey(t *testing.T) {
	pubKey := keys[0]
	signer, err := ParsePublicKey([]byte(pubKey))
	assert.Nil(t, err)
	signature := "a527f1e81f65a00e06ed09434d58ae54ec3acb6f097d1fce8d60781b157c0186da0a0dbefc8cceea5df77b2f95658d94f22fb2641eeb33674ebd85e472f65f2bb1243f2ea1d2d4b6cb20b60c77371eee3fe01227e2ccae1f7bb957d54814d1d9ceefd5b789b57fd10da69961d78e5e60a55326de185f51edcb5bf05bfa6c828b"
	if !assert.Nil(t, err) {
		t.Errorf("could not decode hex signature")
	}
	verified := signer.Verify([]byte("hello"), signature)
	if !assert.Nil(t, verified) {
		t.Errorf("could not verify signature")
	}
}

// TestToBase64 tests that a string will a base 64 encoded string will always remain the same
func TestToBase64(t *testing.T) {
	str := "john doe"
	b64Str := ToBase64([]byte(str))
	assert.Equal(t, b64Str, "am9obiBkb2U=")
}

// TestFromBase64 tests that a base 64 encoded string will be decoded to it's expected value
func TestFromBase64(t *testing.T) {
	str := "john doe"
	b64Str := ToBase64([]byte(str))
	decStr, err := FromBase64(b64Str)
	assert.Nil(t, err)
	assert.NotEqual(t, decStr, "")
	assert.Equal(t, str, decStr)
}

// TestJWS_RSA_Sign tests that a string can be signed correctly using JWS
func TestJWS_RSA_Sign(t *testing.T) {
	key := keys[2]
	expected := "eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoicTZHWW5qZ0tQYkxYSC1rZW5sbjZPZFZRcnl2SEMzVFV1ZS01dnh5QlRwaEhkUWc0djd1Mm9CczZYb1RRSVI2YS1UVlkwR2VFM3ZpakVaX1VwNlZDdG9YUEhWRk51VDBLSmJEaE1IajFVTmZJUnpTdUdOaWJ6bVAzX0NnanRvWWEwdXJyai1ubm5hWjBuYnBVdFRseDB5LW1jVUpnWGZSZDk0QzAtZ1JFUjBNIiwiZSI6IkFRQUIifX0.aGVsbG8.Y-afJCP7c3jP7Vgl78I6sGApb4S7v717VS-kB8Gx5Owg7ePnOr32icaU_y6ESh6-lB_rXyyqktu3-0lOmFN93LQoo-WQOYdxNoVugBZ4OQRXngF2iM7_2qnu_A6NAhhM-a7LQ_q_pnFCYq8RHQycjRAFJgNbqMAezrob9-1vwDE"
	signer, err := ParsePrivateKey([]byte(key))
	assert.Nil(t, err)
	signature, err := signer.JWS_RSA_Sign("hello")
	assert.Nil(t, err)
	assert.Equal(t, signature, expected)
}

func TestJWS_RSA_Verify(t *testing.T) {
	pubKey := keys[0]
	signature := "eyJhbGciOiJSUzI1NiIsImp3ayI6eyJrdHkiOiJSU0EiLCJuIjoicTZHWW5qZ0tQYkxYSC1rZW5sbjZPZFZRcnl2SEMzVFV1ZS01dnh5QlRwaEhkUWc0djd1Mm9CczZYb1RRSVI2YS1UVlkwR2VFM3ZpakVaX1VwNlZDdG9YUEhWRk51VDBLSmJEaE1IajFVTmZJUnpTdUdOaWJ6bVAzX0NnanRvWWEwdXJyai1ubm5hWjBuYnBVdFRseDB5LW1jVUpnWGZSZDk0QzAtZ1JFUjBNIiwiZSI6IkFRQUIifX0.aGVsbG8.Y-afJCP7c3jP7Vgl78I6sGApb4S7v717VS-kB8Gx5Owg7ePnOr32icaU_y6ESh6-lB_rXyyqktu3-0lOmFN93LQoo-WQOYdxNoVugBZ4OQRXngF2iM7_2qnu_A6NAhhM-a7LQ_q_pnFCYq8RHQycjRAFJgNbqMAezrob9-1vwDE"
	sigPayload := "hello"
	signer, err := ParsePublicKey([]byte(pubKey))
	assert.Nil(t, err)
	payload, err := signer.JWS_RSA_Verify(signature)
	assert.Nil(t, err)
	assert.Equal(t, sigPayload, payload)
}

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("hdg-asfr%$2vaah2s_&739X^*@&^hdsg")
	text := []byte("hello friend")
	encText, err := Encrypt(key, text)
	assert.Nil(t, err)

	decText, err := Decrypt(key, encText)
	assert.Nil(t, err)
	assert.Equal(t, text, decText)
}
