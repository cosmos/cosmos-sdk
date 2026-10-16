package crypto

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/cometbft/cometbft/crypto"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bcrypt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/xsalsa20symmetric"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	blockTypePrivKey = "TENDERMINT PRIVATE KEY"
	blockTypeKeyInfo = "TENDERMINT KEY INFO"
	blockTypePubKey  = "TENDERMINT PUBLIC KEY"

	defaultAlgo = "secp256k1"

	headerVersion = "version"
	headerType    = "type"
)

var (
	kdfHeader = "kdf"
	kdfBcrypt = "bcrypt"
	kdfArgon2 = "argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
)

// BcryptSecurityParameter is security parameter var, and it can be changed within the lcd test.
// Making the bcrypt security parameter a var shouldn't be a security issue:
// One can't verify an invalid key by maliciously changing the bcrypt
// parameter during a runtime vulnerability. The main security
// threat this then exposes would be something that changes this during
// runtime before the user creates their key. This vulnerability must
// succeed to update this to that same value before every subsequent call
// to the keys command in future startups / or the attacker must get access
// to the filesystem. However, with a similar threat model (changing
// variables in runtime), one can cause the user to sign a different tx
// than what they see, which is a significantly cheaper attack then breaking
// a bcrypt hash. (Recall that the nonce still exists to break rainbow tables)
// For further notes on security parameter choice, see README.md
var BcryptSecurityParameter uint32 = 12

//-----------------------------------------------------------------
// add armor

// ArmorInfoBytes returns the info from info bytes.
func ArmorInfoBytes(bz []byte) string {
	header := map[string]string{
		headerType:    "Info",
		headerVersion: "0.0.0",
	}

	return EncodeArmor(blockTypeKeyInfo, header, bz)
}

// ArmorPubKeyBytes return the pubkey from info bytes.
func ArmorPubKeyBytes(bz []byte, algo string) string {
	header := map[string]string{
		headerVersion: "0.0.1",
	}
	if algo != "" {
		header[headerType] = algo
	}

	return EncodeArmor(blockTypePubKey, header, bz)
}

//-----------------------------------------------------------------
// remove armor

// UnarmorInfoBytes returns the info bytes from an armored string.
func UnarmorInfoBytes(armorStr string) ([]byte, error) {
	bz, header, err := unarmorBytes(armorStr, blockTypeKeyInfo)
	if err != nil {
		return nil, err
	}

	if header[headerVersion] != "0.0.0" {
		return nil, fmt.Errorf("unrecognized version: %v", header[headerVersion])
	}

	return bz, nil
}

// UnarmorPubKeyBytes returns the pubkey byte slice, a string of the algo type, and an error
func UnarmorPubKeyBytes(armorStr string) (bz []byte, algo string, err error) {
	bz, header, err := unarmorBytes(armorStr, blockTypePubKey)
	if err != nil {
		return nil, "", fmt.Errorf("couldn't unarmor bytes: %w", err)
	}

	switch header[headerVersion] {
	case "0.0.0":
		return bz, defaultAlgo, err
	case "0.0.1":
		if header[headerType] == "" {
			header[headerType] = defaultAlgo
		}

		return bz, header[headerType], err
	case "":
		return nil, "", fmt.Errorf("header's version field is empty")
	default:
		err = fmt.Errorf("unrecognized version: %v", header[headerVersion])
		return nil, "", err
	}
}

func unarmorBytes(armorStr, blockType string) (bz []byte, header map[string]string, err error) {
	bType, header, bz, err := DecodeArmor(armorStr)
	if err != nil {
		return bz, header, err
	}

	if bType != blockType {
		err = fmt.Errorf("unrecognized armor type %q, expected: %q", bType, blockType)
		return bz, header, err
	}

	return bz, header, err
}

//-----------------------------------------------------------------
// encrypt/decrypt with armor

// EncryptArmorPrivKey encrypts and armors the private key.
func EncryptArmorPrivKey(privKey cryptotypes.PrivKey, passphrase, algo string) string {
	saltBytes, encBytes := encryptPrivKey(privKey, passphrase)
	header := map[string]string{
		kdfHeader: kdfArgon2,
		"salt":    fmt.Sprintf("%X", saltBytes),
	}

	if algo != "" {
		header[headerType] = algo
	}

	armorStr := EncodeArmor(blockTypePrivKey, header, encBytes)

	return armorStr
}

func encryptPrivKey(privKey cryptotypes.PrivKey, passphrase string) (saltBytes, encBytes []byte) {
	saltBytes = crypto.CRandBytes(16)

	key := argon2.IDKey([]byte(passphrase), saltBytes, argon2Time, argon2Memory, argon2Threads, chacha20poly1305.KeySize)
	privKeyBytes := legacy.Cdc.MustMarshal(privKey)

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		panic(errorsmod.Wrap(err, "error generating cypher from key"))
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(privKeyBytes)+aead.Overhead()) // Nonce is fixed to maintain consistency, each key is generated at every encryption using a random salt.

	encBytes = aead.Seal(nil, nonce, privKeyBytes, nil)

	return saltBytes, encBytes
}

// UnarmorDecryptPrivKey returns the privkey byte slice, a string of the algo type, and an error
func UnarmorDecryptPrivKey(armorStr, passphrase string) (privKey cryptotypes.PrivKey, algo string, err error) {
	blockType, header, encBytes, err := DecodeArmor(armorStr)
	if err != nil {
		return privKey, "", err
	}

	if blockType != blockTypePrivKey {
		return privKey, "", fmt.Errorf("unrecognized armor type: %v", blockType)
	}

	if header[kdfHeader] != kdfBcrypt && header[kdfHeader] != kdfArgon2 {
		return privKey, "", fmt.Errorf("unrecognized KDF type: %v", header[kdfHeader])
	}

	if header["salt"] == "" {
		return privKey, "", fmt.Errorf("missing salt bytes")
	}

	saltBytes, err := hex.DecodeString(header["salt"])
	if err != nil {
		return privKey, "", fmt.Errorf("error decoding salt: %v", err.Error())
	}

	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase, header[kdfHeader])

	if header[headerType] == "" {
		header[headerType] = defaultAlgo
	}

	return privKey, header[headerType], err
}

func decryptPrivKey(saltBytes, encBytes []byte, passphrase, kdf string) (privKey cryptotypes.PrivKey, err error) {
	// Key derivation
	var (
		key          []byte
		privKeyBytes []byte
	)

	// Since the argon2 key derivation and chacha encryption was implemented together, it is not possible to have mixed kdf and encryption algorithms
	switch kdf {
	case kdfArgon2:
		key = argon2.IDKey([]byte(passphrase), saltBytes, argon2Time, argon2Memory, argon2Threads, chacha20poly1305.KeySize)

		aead, err := chacha20poly1305.New(key)
		if err != nil {
			return privKey, errorsmod.Wrap(err, "Error generating aead cypher for key.")
		} else if len(encBytes) < aead.NonceSize() {
			return privKey, errorsmod.Wrap(nil, "Encrypted bytes length is smaller than aead nonce size.")
		}
		nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(privKeyBytes)+aead.Overhead())
		privKeyBytes, err = aead.Open(nil, nonce, encBytes, nil) // Decrypt the message and check it wasn't tampered with.
		if err != nil {
			return privKey, sdkerrors.ErrWrongPassword
		}
	case kdfBcrypt:
		key, err = bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), BcryptSecurityParameter)
		if err != nil {
			return privKey, errorsmod.Wrap(err, "Error generating bcrypt cypher for key.")
		}
		key = crypto.Sha256(key) // Get 32 bytes
		privKeyBytes, err = xsalsa20symmetric.DecryptSymmetric(encBytes, key)

		if errors.Is(err, xsalsa20symmetric.ErrCiphertextDecrypt) {
			return privKey, sdkerrors.ErrWrongPassword
		}
	default:
		return privKey, errorsmod.Wrap(nil, fmt.Sprintf("Unrecognized key derivation function (kdf) header: %s.", kdf))
	}

	if err != nil {
		return privKey, err
	}

	return legacy.PrivKeyFromBytes(privKeyBytes)
}

//-----------------------------------------------------------------
// encode/decode with armor

func EncodeArmor(blockType string, headers map[string]string, data []byte) string {
	buf := new(bytes.Buffer)
	w, err := encodeArmorBlock(buf, blockType, headers)
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %w", err))
	}
	_, err = w.Write(data)
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %w", err))
	}
	err = w.Close()
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %w", err))
	}
	return buf.String()
}

func DecodeArmor(armorStr string) (blockType string, headers map[string]string, data []byte, err error) {
	buf := bytes.NewBufferString(armorStr)
	block, err := decodeArmorBlock(buf)
	if err != nil {
		return "", nil, nil, err
	}
	data, err = io.ReadAll(block.Body)
	if err != nil {
		return "", nil, nil, err
	}
	return block.Type, block.Header, data, nil
}

//-----------------------------------------------------------------
// OpenPGP ASCII armor, see RFC 4880. OpenPGP armor is very similar to PEM
// except that it has an additional CRC checksum.
//
// MODIFIED BY COSMOS SDK: copied from golang.org/x/crypto/openpgp/armor
// (v0.57.0), which is deprecated and flagged as GO-2026-5932. Behavior is
// unchanged. Changes:
//   - identifiers are unexported, and errArmorCorrupt replaces ArmorCorrupt
//     with the same message, so that golang.org/x/crypto/openpgp/errors is not
//     imported;
//   - naked returns are made explicit and the deprecation notice is dropped.
//
// The code below is covered by the following license:
//
// Copyright 2009 The Go Authors.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google LLC nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// armorBlock represents an OpenPGP armored structure.
//
// The encoded form is:
//
//	-----BEGIN Type-----
//	Headers
//
//	base64-encoded Bytes
//	'=' base64 encoded checksum
//	-----END Type-----
//
// where Headers is a possibly empty sequence of Key: Value lines.
//
// Since the armored data can be very large, this package presents a streaming
// interface.
type armorBlock struct {
	Type    string            // The type, taken from the preamble (i.e. "PGP SIGNATURE").
	Header  map[string]string // Optional headers.
	Body    io.Reader         // A Reader from which the contents can be read
	lReader lineReader
	oReader openpgpReader
}

// errArmorCorrupt is returned when an armored block is malformed or its contents do
// not match its CRC-24 checksum.
var errArmorCorrupt = errors.New("openpgp: invalid data: armor invalid")

const (
	crc24Init = 0xb704ce
	crc24Poly = 0x1864cfb
	crc24Mask = 0xffffff
)

// crc24 calculates the OpenPGP checksum as specified in RFC 4880, section 6.1
func crc24(crc uint32, d []byte) uint32 {
	for _, b := range d {
		crc ^= uint32(b) << 16
		for range 8 {
			crc <<= 1
			if crc&0x1000000 != 0 {
				crc ^= crc24Poly
			}
		}
	}
	return crc
}

var (
	armorStart     = []byte("-----BEGIN ")
	armorEnd       = []byte("-----END ")
	armorEndOfLine = []byte("-----")
)

// lineReader wraps a line based reader. It watches for the end of an armor
// block and records the expected CRC value.
type lineReader struct {
	in     *bufio.Reader
	buf    []byte
	eof    bool
	crc    uint32
	crcSet bool
}

func (l *lineReader) Read(p []byte) (int, error) {
	if l.eof {
		return 0, io.EOF
	}

	if len(l.buf) > 0 {
		n := copy(p, l.buf)
		l.buf = l.buf[n:]
		return n, nil
	}

	line, isPrefix, err := l.in.ReadLine()
	if err != nil {
		return 0, err
	}
	if isPrefix {
		return 0, errArmorCorrupt
	}

	if bytes.HasPrefix(line, armorEnd) {
		l.eof = true
		return 0, io.EOF
	}

	if len(line) == 5 && line[0] == '=' {
		// This is the checksum line
		var expectedBytes [3]byte
		m, err := base64.StdEncoding.Decode(expectedBytes[0:], line[1:])
		if m != 3 || err != nil {
			return 0, err
		}
		l.crc = uint32(expectedBytes[0])<<16 |
			uint32(expectedBytes[1])<<8 |
			uint32(expectedBytes[2])

		line, _, err = l.in.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}
		if !bytes.HasPrefix(line, armorEnd) {
			return 0, errArmorCorrupt
		}

		l.eof = true
		l.crcSet = true
		return 0, io.EOF
	}

	if len(line) > 96 {
		return 0, errArmorCorrupt
	}

	n := copy(p, line)
	bytesToSave := len(line) - n
	if bytesToSave > 0 {
		if cap(l.buf) < bytesToSave {
			l.buf = make([]byte, 0, bytesToSave)
		}
		l.buf = l.buf[0:bytesToSave]
		copy(l.buf, line[n:])
	}

	return n, nil
}

// openpgpReader passes Read calls to the underlying base64 decoder, but keeps
// a running CRC of the resulting data and checks the CRC against the value
// found by the lineReader at EOF.
type openpgpReader struct {
	lReader    *lineReader
	b64Reader  io.Reader
	currentCRC uint32
}

func (r *openpgpReader) Read(p []byte) (int, error) {
	n, err := r.b64Reader.Read(p)
	r.currentCRC = crc24(r.currentCRC, p[:n])

	if errors.Is(err, io.EOF) && r.lReader.crcSet && r.lReader.crc != r.currentCRC&crc24Mask {
		return 0, errArmorCorrupt
	}

	return n, err
}

// decodeArmorBlock reads a PGP armored block from the given Reader. It will ignore
// leading garbage. If it doesn't find a block, it will return nil, io.EOF. The
// given Reader is not usable after calling this function: an arbitrary amount
// of data may have been read past the end of the block.
func decodeArmorBlock(in io.Reader) (*armorBlock, error) {
	r := bufio.NewReaderSize(in, 100)
	var line []byte
	var err error
	ignoreNext := false

TryNextBlock:
	// Skip leading garbage
	for {
		ignoreThis := ignoreNext
		line, ignoreNext, err = r.ReadLine()
		if err != nil {
			return nil, err
		}
		if ignoreNext || ignoreThis {
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) > len(armorStart)+len(armorEndOfLine) && bytes.HasPrefix(line, armorStart) {
			break
		}
	}

	p := new(armorBlock)
	p.Type = string(line[len(armorStart) : len(line)-len(armorEndOfLine)])
	p.Header = make(map[string]string)
	nextIsContinuation := false
	var lastKey string

	// Read headers
	for {
		isContinuation := nextIsContinuation
		line, nextIsContinuation, err = r.ReadLine()
		if err != nil {
			return nil, err
		}
		if isContinuation {
			p.Header[lastKey] += string(line)
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			break
		}

		i := bytes.Index(line, []byte(": "))
		if i == -1 {
			goto TryNextBlock
		}
		lastKey = string(line[:i])
		p.Header[lastKey] = string(line[i+2:])
	}

	p.lReader.in = r
	p.oReader.currentCRC = crc24Init
	p.oReader.lReader = &p.lReader
	p.oReader.b64Reader = base64.NewDecoder(base64.StdEncoding, &p.lReader)
	p.Body = &p.oReader

	return p, nil
}

var (
	armorHeaderSep    = []byte(": ")
	blockEnd          = []byte("\n=")
	newline           = []byte("\n")
	armorEndOfLineOut = []byte("-----\n")
)

// writeSlices writes its arguments to the given Writer.
func writeSlices(out io.Writer, slices ...[]byte) error {
	for _, s := range slices {
		if _, err := out.Write(s); err != nil {
			return err
		}
	}
	return nil
}

// lineBreaker breaks data across several lines, all of the same byte length
// (except possibly the last). Lines are broken with a single '\n'.
type lineBreaker struct {
	lineLength  int
	line        []byte
	used        int
	out         io.Writer
	haveWritten bool
}

func newLineBreaker(out io.Writer, lineLength int) *lineBreaker {
	return &lineBreaker{
		lineLength: lineLength,
		line:       make([]byte, lineLength),
		used:       0,
		out:        out,
	}
}

func (l *lineBreaker) Write(b []byte) (int, error) {
	n := len(b)

	if n == 0 {
		return 0, nil
	}

	if l.used == 0 && l.haveWritten {
		if _, err := l.out.Write([]byte{'\n'}); err != nil {
			return n, err
		}
	}

	if l.used+len(b) < l.lineLength {
		l.used += copy(l.line[l.used:], b)
		return n, nil
	}

	l.haveWritten = true
	if _, err := l.out.Write(l.line[0:l.used]); err != nil {
		return n, err
	}
	excess := l.lineLength - l.used
	l.used = 0

	if _, err := l.out.Write(b[0:excess]); err != nil {
		return n, err
	}

	_, err := l.Write(b[excess:])
	return n, err
}

func (l *lineBreaker) Close() error {
	if l.used > 0 {
		if _, err := l.out.Write(l.line[0:l.used]); err != nil {
			return err
		}
	}

	return nil
}

// armorEncoding keeps track of a running CRC24 over the data which has been written
// to it and outputs a OpenPGP checksum when closed, followed by an armor
// trailer.
//
// It's built into a stack of io.Writers:
//
//	armorEncoding -> base64 encoder -> lineBreaker -> out
type armorEncoding struct {
	out       io.Writer
	breaker   *lineBreaker
	b64       io.WriteCloser
	crc       uint32
	blockType []byte
}

func (e *armorEncoding) Write(data []byte) (int, error) {
	e.crc = crc24(e.crc, data)
	return e.b64.Write(data)
}

func (e *armorEncoding) Close() error {
	if err := e.b64.Close(); err != nil {
		return err
	}
	if err := e.breaker.Close(); err != nil {
		return err
	}

	var checksumBytes [3]byte
	checksumBytes[0] = byte(e.crc >> 16)
	checksumBytes[1] = byte(e.crc >> 8)
	checksumBytes[2] = byte(e.crc)

	var b64ChecksumBytes [4]byte
	base64.StdEncoding.Encode(b64ChecksumBytes[:], checksumBytes[:])

	return writeSlices(e.out, blockEnd, b64ChecksumBytes[:], newline, armorEnd, e.blockType, armorEndOfLine)
}

// encodeArmorBlock returns a WriteCloser which will encode the data written to it in
// OpenPGP armor.
func encodeArmorBlock(out io.Writer, blockType string, headers map[string]string) (io.WriteCloser, error) {
	bType := []byte(blockType)
	if err := writeSlices(out, armorStart, bType, armorEndOfLineOut); err != nil {
		return nil, err
	}

	for k, v := range headers {
		if err := writeSlices(out, []byte(k), armorHeaderSep, []byte(v), newline); err != nil {
			return nil, err
		}
	}

	if _, err := out.Write(newline); err != nil {
		return nil, err
	}

	e := &armorEncoding{
		out:       out,
		breaker:   newLineBreaker(out, 64),
		crc:       crc24Init,
		blockType: bType,
	}
	e.b64 = base64.NewEncoder(base64.StdEncoding, e.breaker)
	return e, nil
}
