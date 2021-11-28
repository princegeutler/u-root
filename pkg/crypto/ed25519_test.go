// Copyright 2017-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crypto

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

const (
	// publicKeyDERFile is a RSA public key in DER format
	publicKeyDERFile string = "tests/public_key.der"
	// publicKeyPEMFile is a RSA public key in PEM format
	publicKeyPEMFile string = "tests/public_key.pem"
	// privateKeyPEMFile is a RSA public key in PEM format
	privateKeyPEMFile string = "tests/private_key.pem"
	// testDataFile which should be verified by the good signature
	testDataFile string = "tests/data"
	// signatureGoodFile is a good signature of testDataFile
	signatureGoodFile string = "tests/verify_rsa_pkcs15_sha256.signature"
	// signatureBadFile is a bad signature which does not work with testDataFile
	signatureBadFile string = "tests/verify_rsa_pkcs15_sha256.signature2"
)

// password is a PEM encrypted passphrase
var password = []byte{'k', 'e', 'i', 'n', 's'}

func TestLoadDERPublicKey(t *testing.T) {
	_, err := LoadPublicKeyFromFile(publicKeyDERFile)
	require.Error(t, err)
}

func TestLoadPEMPublicKey(t *testing.T) {
	_, err := LoadPublicKeyFromFile(publicKeyPEMFile)
	require.NoError(t, err)
}

func TestLoadPEMPrivateKey(t *testing.T) {
	_, err := LoadPrivateKeyFromFile(privateKeyPEMFile, password)
	require.NoError(t, err)
}

func TestLoadBadPEMPrivateKey(t *testing.T) {
	_, err := LoadPrivateKeyFromFile(privateKeyPEMFile, []byte{})
	require.Error(t, err)
}

func TestSignVerifyData(t *testing.T) {
	privateKey, err := LoadPrivateKeyFromFile(privateKeyPEMFile, password)
	require.NoError(t, err)

	publicKey, err := LoadPublicKeyFromFile(publicKeyPEMFile)
	require.NoError(t, err)

	testData, err := ioutil.ReadFile(testDataFile)
	require.NoError(t, err)

	signature := ed25519.Sign(privateKey, testData)
	verified := ed25519.Verify(publicKey, testData, signature)
	require.Equal(t, true, verified)
}

func TestGoodSignature(t *testing.T) {
	publicKey, err := LoadPublicKeyFromFile(publicKeyPEMFile)
	require.NoError(t, err)

	testData, err := ioutil.ReadFile(testDataFile)
	require.NoError(t, err)

	signatureGood, err := ioutil.ReadFile(signatureGoodFile)
	require.NoError(t, err)

	verified := ed25519.Verify(publicKey, testData, signatureGood)
	require.Equal(t, true, verified)
}

func TestBadSignature(t *testing.T) {
	publicKey, err := LoadPublicKeyFromFile(publicKeyPEMFile)
	require.NoError(t, err)

	testData, err := ioutil.ReadFile(testDataFile)
	require.NoError(t, err)

	signatureBad, err := ioutil.ReadFile(signatureBadFile)
	require.NoError(t, err)

	verified := ed25519.Verify(publicKey, testData, signatureBad)
	require.Equal(t, false, verified)
}

func TestGenerateKeys(t *testing.T) {
	// FIXME: move this to testing.TempDir once we require >= Go 1.15
	tmpdir, err := ioutil.TempDir("", "generate-keys")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	err = GeneratED25519Key(password, path.Join(tmpdir, "private_key.pem"), path.Join(tmpdir, "public_key.pem"))
	require.NoError(t, err)
}

func TestGenerateUnprotectedKeys(t *testing.T) {
	// FIXME: move this to testing.TempDir once we require >= Go 1.15
	tmpdir, err := ioutil.TempDir("", "generate-keys")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	err = GeneratED25519Key(nil, path.Join(tmpdir, "private_key.pem"), path.Join(tmpdir, "public_key.pem"))
	require.NoError(t, err)
}
