// [SIKE] http://www.sike.org/files/SIDH-spec.pdf
// [REF] https://github.com/Microsoft/PQCrypto-SIDH
package sike

import (
	"crypto/subtle"
	"errors"
	"io"
	// TODO: Use implementation from xcrypto, once PR below merged
	// https://go-review.googlesource.com/c/crypto/+/111281/
	. "github.com/henrydcase/nobs/dh/sidh"
	cshake "github.com/henrydcase/nobs/hash/sha3"
)

// Constants used for cSHAKE customization
// Those values are different than in [SIKE] - they are encoded on 16bits. This is
// done in order for implementation to be compatible with [REF] and test vectors.
var G = []byte{0x00, 0x00}
var H = []byte{0x01, 0x00}
var F = []byte{0x02, 0x00}

// Generates cShake-256 sum
func cshakeSum(out, in, S []byte) {
	h := cshake.NewCShake256(nil, S)
	h.Write(in)
	h.Read(out)
}

func encrypt(skA *PrivateKey, pkA, pkB *PublicKey, ptext []byte) ([]byte, error) {
	var tmp [40]byte
	var ptextLen = len(ptext)

	if pkB.Variant() != KeyVariant_SIKE {
		return nil, errors.New("wrong key type")
	}

	j, err := DeriveSecret(skA, pkB)
	if err != nil {
		return nil, err
	}

	cshakeSum(tmp[:ptextLen], j, F)
	for i, _ := range ptext {
		tmp[i] ^= ptext[i]
	}

	ret := make([]byte, pkA.Size()+ptextLen)
	copy(ret, pkA.Export())
	copy(ret[pkA.Size():], tmp[:ptextLen])
	return ret, nil
}

// -----------------------------------------------------------------------------
// PKE interface
//

// Uses SIKE public key to encrypt plaintext. Requires cryptographically secure PRNG
// Returns ciphertext in case encryption succeeds. Returns error in case PRNG fails
// or wrongly formated input was provided.
func Encrypt(rng io.Reader, pub *PublicKey, ptext []byte) ([]byte, error) {
	var params = pub.Params()
	var ptextLen = uint(len(ptext))
	// c1 must be security level + 64 bits (see [SIKE] 1.4 and 4.3.3)
	if ptextLen != (params.KemSize + 8) {
		return nil, errors.New("Unsupported message length")
	}

	skA := NewPrivateKey(params.Id, KeyVariant_SIDH_A)
	err := skA.Generate(rng)
	if err != nil {
		return nil, err
	}

	pkA, _ := GeneratePublicKey(skA) // Never fails
	return encrypt(skA, pkA, pub, ptext)
}

// Uses SIKE private key to decrypt ciphertext. Returns plaintext in case
// decryption succeeds or error in case unexptected input was provided.
func Decrypt(prv *PrivateKey, ctext []byte) ([]byte, error) {
	var params = prv.Params()
	var tmp [40]byte
	var c1_len int
	var pk_len = params.PublicKeySize

	if prv.Variant() != KeyVariant_SIKE {
		return nil, errors.New("wrong key type")
	}

	// ctext is a concatenation of (pubkey_A || c1=ciphertext)
	// it must be security level + 64 bits (see [SIKE] 1.4 and 4.3.3)
	c1_len = len(ctext) - pk_len
	if c1_len != (int(params.KemSize) + 8) {
		return nil, errors.New("wrong size of cipher text")
	}

	c0 := NewPublicKey(params.Id, KeyVariant_SIDH_A)
	err := c0.Import(ctext[:pk_len])
	if err != nil {
		return nil, err
	}
	j, err := DeriveSecret(prv, c0)
	if err != nil {
		return nil, err
	}

	cshakeSum(tmp[:c1_len], j, F)
	for i, _ := range tmp[:c1_len] {
		tmp[i] ^= ctext[pk_len+i]
	}

	return tmp[:c1_len], nil
}

// -----------------------------------------------------------------------------
// KEM interface
//

// Encapsulation receives the public key and generates SIKE ciphertext and shared secret.
// The generated ciphertext is used for authentication.
// The rng must be cryptographically secure PRNG.
// Error is returned in case PRNG fails or wrongly formated input was provided.
func Encapsulate(rng io.Reader, pub *PublicKey) (ctext []byte, secret []byte, err error) {
	var params = pub.Params()
	// Buffer for random, secret message
	var ptext = make([]byte, params.MsgLen)
	// r = G(ptext||pub)
	var r = make([]byte, params.SecretKeySize)
	// Resulting shared secret
	secret = make([]byte, params.KemSize)

	// Generate ephemeral value
	_, err = io.ReadFull(rng, ptext)
	if err != nil {
		return nil, nil, err
	}

	h := cshake.NewCShake256(nil, G)
	h.Write(ptext)
	h.Write(pub.Export())
	h.Read(r)

	// cSHAKE256 implementation is byte oriented. Ensure bitlength is less then to E2
	r[len(r)-1] &= params.A.MaskBytes[0]
	r[len(r)-2] &= params.A.MaskBytes[1] // clear high bits, so scalar < 2*732

	// (c0 || c1) = Enc(pkA, ptext; r)
	skA := NewPrivateKey(params.Id, KeyVariant_SIDH_A)
	err = skA.Import(r)
	if err != nil {
		return nil, nil, err
	}

	pkA, _ := GeneratePublicKey(skA) // Never fails
	ctext, err = encrypt(skA, pkA, pub, ptext)
	if err != nil {
		return nil, nil, err
	}

	// K = H(ptext||(c0||c1))
	h = cshake.NewCShake256(nil, H)
	h.Write(ptext)
	h.Write(ctext)
	h.Read(secret)

	return ctext, secret, nil
}

// Decapsulate receives rng - cryptographically secure PRNG, keypair and ciphertext generated
// by Encapsulate().
// It returns shared secret in case cipertext was generated with 'pub' or random value otherwise.
// Key generation, import and export functions ensure that if KEM decapsulation fails, always
// same random value is returned.
// Decapsulation may fail when wrongly formated input is provided or PRNG fails.
func Decapsulate(rng io.Reader, prv *PrivateKey, pub *PublicKey, ctext []byte) ([]byte, error) {
	var params = pub.Params()
	var r = make([]byte, params.SecretKeySize)
	// Resulting shared secret
	var secret = make([]byte, params.KemSize)
	var skA = NewPrivateKey(params.Id, KeyVariant_SIDH_A)

	m, err := Decrypt(prv, ctext)
	if err != nil {
		return nil, err
	}

	// r' = G(m'||pub)
	h := cshake.NewCShake256(nil, G)
	h.Write(m)
	h.Write(pub.Export())
	h.Read(r)

	// cSHAKE256 implementation is byte oriented: Ensure bitlength is equal to E2
	r[len(r)-1] &= params.A.MaskBytes[0]
	r[len(r)-2] &= params.A.MaskBytes[1] // clear high bits, so scalar < 2*732

	err = skA.Import(r)
	if err != nil {
		return nil, err
	}

	pkA, _ := GeneratePublicKey(skA) // Never fails
	c0 := pkA.Export()

	h = cshake.NewCShake256(nil, H)
	if subtle.ConstantTimeCompare(c0, ctext[:len(c0)]) == 1 {
		h.Write(m)
	} else {
		// S is chosen at random when generating a key and unknown to other party. It
		// may seem weird, but it's correct. It is important that S is unpredictable
		// to other party. Without this check, it is possible to recover a secret, by
		// providing series of invalid ciphertexts. It is also important that in case
		//
		// See more details in "On the security of supersingular isogeny cryptosystems"
		// (S. Galbraith, et al., 2016, ePrint #859).
		h.Write(prv.S)
	}
	h.Write(ctext)
	h.Read(secret)
	return secret, nil
}