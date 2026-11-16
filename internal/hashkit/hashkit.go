package hashkit

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
)

// Spec describes one supported algorithm.
type Spec struct {
	Name string
	New  func() hash.Hash
}

// specs is ordered the way algorithms are presented in the UI.
var specs = []Spec{
	{"MD5", func() hash.Hash { return md5.New() }},
	{"SHA1", func() hash.Hash { return sha1.New() }},
	{"SHA224", func() hash.Hash { return sha256.New224() }},
	{"SHA256", func() hash.Hash { return sha256.New() }},
	{"SHA384", func() hash.Hash { return sha512.New384() }},
	{"SHA512", func() hash.Hash { return sha512.New() }},
	{"SHA512_224", func() hash.Hash { return sha512.New512_224() }},
	{"SHA512_256", func() hash.Hash { return sha512.New512_256() }},
	{"SHA3_224", func() hash.Hash { return sha3.New224() }},
	{"SHA3_256", func() hash.Hash { return sha3.New256() }},
	{"SHA3_384", func() hash.Hash { return sha3.New384() }},
	{"SHA3_512", func() hash.Hash { return sha3.New512() }},
	{"BLAKE2b", func() hash.Hash {
		h, _ := blake2b.New512(nil)
		return h
	}},
	{"BLAKE2s", func() hash.Hash {
		h, _ := blake2s.New256(nil)
		return h
	}},
	{"SM3", func() hash.Hash { return NewSM3() }},
	{"CRC32", func() hash.Hash { return crc32.NewIEEE() }},
}

var byName = func() map[string]Spec {
	m := make(map[string]Spec, len(specs))
	for _, s := range specs {
		m[strings.ToUpper(s.Name)] = s
	}
	return m
}()

// Names returns every supported algorithm in display order.
func Names() []string {
	out := make([]string, len(specs))
	for i, s := range specs {
		out[i] = s.Name
	}
	return out
}

// Resolve maps requested names to canonical specs, dropping duplicates.
func Resolve(names []string) ([]Spec, error) {
	seen := make(map[string]bool, len(names))
	out := make([]Spec, 0, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		spec, ok := byName[strings.ToUpper(name)]
		if !ok {
			return nil, fmt.Errorf("unsupported algorithm %q", name)
		}
		if seen[spec.Name] {
			continue
		}
		seen[spec.Name] = true
		out = append(out, spec)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no algorithms selected")
	}
	return out, nil
}

// Hash reads r once and computes every requested digest in a single pass.
// It returns the lowercase hex digests keyed by canonical algorithm name.
func Hash(r io.Reader, names []string) (map[string]string, int64, error) {
	resolved, err := Resolve(names)
	if err != nil {
		return nil, 0, err
	}

	hashers := make([]hash.Hash, len(resolved))
	writers := make([]io.Writer, len(resolved))
	for i, spec := range resolved {
		hashers[i] = spec.New()
		writers[i] = hashers[i]
	}

	n, err := io.CopyBuffer(io.MultiWriter(writers...), r, make([]byte, 1<<20))
	if err != nil {
		return nil, n, err
	}

	out := make(map[string]string, len(resolved))
	for i, spec := range resolved {
		out[spec.Name] = hex.EncodeToString(hashers[i].Sum(nil))
	}
	return out, n, nil
}
