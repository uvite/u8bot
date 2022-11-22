package ta

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"hash"

	"github.com/dop251/goja"

	"github.com/uvite/u8/js/common"
	"github.com/uvite/u8/js/modules"
)

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// Ta represents an instance of the Ta module.
	Ta struct {
		vu modules.VU
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Ta{}
)

// New returns a pointer to a new RootModule instance.
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Ta{vu: vu}
}

// Exports returns the exports of the execution module.
func (c *Ta) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"atr": c.atr,
		},
	}
}

// randomBytes returns random data of the given size.
func (c *Ta) atr(size int) (*goja.ArrayBuffer, error) {
	if size < 1 {
		return nil, errors.New("invalid size")
	}
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	ab := c.vu.Runtime().NewArrayBuffer(bytes)
	return &ab, nil
}

// md4 returns the MD4 hash of input in the given encoding.
func (c *Ta) md4(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("md4")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// md5 returns the MD5 hash of input in the given encoding.
func (c *Ta) md5(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("md5")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// sha1 returns the SHA1 hash of input in the given encoding.
func (c *Ta) sha1(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha1")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// sha256 returns the SHA256 hash of input in the given encoding.
func (c *Ta) sha256(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha256")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// sha384 returns the SHA384 hash of input in the given encoding.
func (c *Ta) sha384(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha384")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// sha512 returns the SHA512 hash of input in the given encoding.
func (c *Ta) sha512(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha512")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// sha512_224 returns the SHA512/224 hash of input in the given encoding.
func (c *Ta) sha512_224(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha512_224")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// shA512_256 returns the SHA512/256 hash of input in the given encoding.
func (c *Ta) sha512_256(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("sha512_256")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// ripemd160 returns the RIPEMD160 hash of input in the given encoding.
func (c *Ta) ripemd160(input interface{}, outputEncoding string) (interface{}, error) {
	hasher := c.createHash("ripemd160")
	hasher.Update(input)
	return hasher.Digest(outputEncoding)
}

// createHash returns a Hasher instance that uses the given algorithm.
func (c *Ta) createHash(algorithm string) *Hasher {
	hashfn := c.parseHashFunc(algorithm)
	return &Hasher{
		runtime: c.vu.Runtime(),
		hash:    hashfn(),
	}
}

// hexEncode returns a string with the hex representation of the provided byte
// array or ArrayBuffer.
func (c *Ta) hexEncode(data interface{}) (string, error) {
	d, err := common.ToBytes(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(d), nil
}

// createHMAC returns a new HMAC hash using the given algorithm and key.
func (c *Ta) createHMAC(algorithm string, key interface{}) (*Hasher, error) {
	h := c.parseHashFunc(algorithm)
	if h == nil {
		return nil, fmt.Errorf("invalid algorithm: %s", algorithm)
	}

	kb, err := common.ToBytes(key)
	if err != nil {
		return nil, err
	}
	return &Hasher{runtime: c.vu.Runtime(), hash: hmac.New(h, kb)}, nil
}

// HMAC returns a new HMAC hash of input using the given algorithm and key
// in the given encoding.
func (c *Ta) hmac(algorithm string, key, input interface{}, outputEncoding string) (interface{}, error) {
	hasher, err := c.createHMAC(algorithm, key)
	if err != nil {
		return nil, err
	}
	err = hasher.Update(input)
	if err != nil {
		return nil, err
	}
	return hasher.Digest(outputEncoding)
}

func (c *Ta) parseHashFunc(a string) func() hash.Hash {
	var h func() hash.Hash
	switch a {
	case "md4":
		h = md4.New
	case "md5":
		h = md5.New
	case "sha1":
		h = sha1.New
	case "sha256":
		h = sha256.New
	case "sha384":
		h = sha512.New384
	case "sha512_224":
		h = sha512.New512_224
	case "sha512_256":
		h = sha512.New512_256
	case "sha512":
		h = sha512.New
	case "ripemd160":
		h = ripemd160.New
	}
	return h
}

// Hasher wraps an hash.Hash with goja.Runtime.
type Hasher struct {
	runtime *goja.Runtime
	hash    hash.Hash
}

// Update the hash with the input data.
func (hasher *Hasher) Update(input interface{}) error {
	d, err := common.ToBytes(input)
	if err != nil {
		return err
	}
	_, err = hasher.hash.Write(d)
	if err != nil {
		return err
	}
	return nil
}

// Digest returns the hash value in the given encoding.
func (hasher *Hasher) Digest(outputEncoding string) (interface{}, error) {
	sum := hasher.hash.Sum(nil)

	switch outputEncoding {
	case "base64":
		return base64.StdEncoding.EncodeToString(sum), nil

	case "base64url":
		return base64.URLEncoding.EncodeToString(sum), nil

	case "base64rawurl":
		return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum), nil

	case "hex":
		return hex.EncodeToString(sum), nil

	case "binary":
		ab := hasher.runtime.NewArrayBuffer(sum)
		return &ab, nil

	default:
		return nil, fmt.Errorf("invalid output encoding: %s", outputEncoding)
	}
}
