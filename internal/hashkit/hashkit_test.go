package hashkit

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
)

type vector struct {
	algo string
	size int
	want string
}

func testData(size int) []byte {
	switch size {
	case 0:
		return nil
	case 3:
		return []byte("abc")
	default:
		data := make([]byte, 0, size)
		for len(data) < 1048576 {
			for i := 0; i < 256; i++ {
				data = append(data, byte(i))
			}
		}
		return append(data, []byte("prismatic")...)
	}
}

// Reference digests produced independently with Python hashlib/OpenSSL.
var referenceVectors = []vector{
	// empty input
	{"MD5", 0, "d41d8cd98f00b204e9800998ecf8427e"},
	{"SHA1", 0, "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
	{"SHA224", 0, "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f"},
	{"SHA256", 0, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	{"SHA384", 0, "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"},
	{"SHA512", 0, "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
	{"SHA512_224", 0, "6ed0dd02806fa89e25de060c19d3ac86cabb87d6a0ddd05c333b84f4"},
	{"SHA512_256", 0, "c672b8d1ef56ed28ab87c3622c5114069bdd3ad7b8f9737498d0c01ecef0967a"},
	{"SHA3_224", 0, "6b4e03423667dbb73b6e15454f0eb1abd4597f9a1b078e3f5b5a6bc7"},
	{"SHA3_256", 0, "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"},
	{"SHA3_384", 0, "0c63a75b845e4f7d01107d852e4c2485c51a50aaaa94fc61995e71bbee983a2ac3713831264adb47fb6bd1e058d5f004"},
	{"SHA3_512", 0, "a69f73cca23a9ac5c8b567dc185a756e97c982164fe25859e0d1dcc1475c80a615b2123af1f5f94c11e3e9402c3ac558f500199d95b6d3e301758586281dcd26"},
	{"BLAKE2b", 0, "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce"},
	{"BLAKE2s", 0, "69217a3079908094e11121d042354a7c1f55b6482ca1a51e1b250dfd1ed0eef9"},
	{"SM3", 0, "1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b"},
	{"CRC32", 0, "00000000"},
	// "abc"
	{"MD5", 3, "900150983cd24fb0d6963f7d28e17f72"},
	{"SHA1", 3, "a9993e364706816aba3e25717850c26c9cd0d89d"},
	{"SHA224", 3, "23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7"},
	{"SHA256", 3, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	{"SHA384", 3, "cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7"},
	{"SHA512", 3, "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
	{"SHA512_224", 3, "4634270f707b6a54daae7530460842e20e37ed265ceee9a43e8924aa"},
	{"SHA512_256", 3, "53048e2681941ef99b2e29b76b4c7dabe4c2d0c634fc6d46e0e2f13107e7af23"},
	{"SHA3_224", 3, "e642824c3f8cf24ad09234ee7d3c766fc9a3a5168d0c94ad73b46fdf"},
	{"SHA3_256", 3, "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532"},
	{"SHA3_384", 3, "ec01498288516fc926459f58e2c6ad8df9b473cb0fc08c2596da7cf0e49be4b298d88cea927ac7f539f1edf228376d25"},
	{"SHA3_512", 3, "b751850b1a57168a5693cd924b6b096e08f621827444f70d884f5d0240d2712e10e116e9192af3c91a7ec57647e3934057340b4cf408d5a56592f8274eec53f0"},
	{"BLAKE2b", 3, "ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923"},
	{"BLAKE2s", 3, "508c5e8c327c14e2e1a72ba34eeb452f37458b209ed63a294d999b4c86675982"},
	{"SM3", 3, "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"},
	{"CRC32", 3, "352441c2"},
	// multi-block input, 1 MiB + 9 bytes
	{"MD5", 1048585, "be5e08fc163a5443ba29550bf274d8bf"},
	{"SHA1", 1048585, "f1feab23f0f3c3915eba0f349acb0bb0a3d0585a"},
	{"SHA224", 1048585, "86511a92a41276f98ab05b0cfaef35d0e171bb386c41540551d43f99"},
	{"SHA256", 1048585, "d1ab13c433533f7923d7341680ae92cf82c875eef8d3a4949102d29a11975913"},
	{"SHA384", 1048585, "80fccf0bbe66abc07fa4269d4a47352088061de6080ba860ac7de5c4eaf6d8e63dabe3c676f570bd94f403086b4e9c39"},
	{"SHA512", 1048585, "dc371eefc0503352683df7838ef0ccff07c77afa6f481a98beefa50e851ffde32876ba9c0a679a6e1ce6b12cbc92b6ea2708d49a690ff02ed9a1db1986b6553b"},
	{"SHA512_224", 1048585, "5d8c32d94a3a63d56171e52e37a6864f7b2a1b468962b15927a3a324"},
	{"SHA512_256", 1048585, "49c277f6ccd4b8c02638d231519ac49f9817e3f074a04af86262103c32e73188"},
	{"SHA3_224", 1048585, "a9ca2743ac6a6ae6a2863a3689d5b9c9216de2124a9192da55a3a20a"},
	{"SHA3_256", 1048585, "787e0110a2dbfd31792823da27082d83cbfb76abf2cbdb9293409e3bddc3ead5"},
	{"SHA3_384", 1048585, "71c3b236a7020d78b18dfb1cc4dac91590bee0d0ca440ef095eaa91c3195af6760a8e77f5bf82712d06d326426ac6dfb"},
	{"SHA3_512", 1048585, "fad40b7aa9a8d81138cbd5cae31fc5266f3956c2248883bd399fb798e9e479bf438ee858957a24d591de24910eee6ccf50b47431441b372746d2e4981c22312e"},
	{"BLAKE2b", 1048585, "5dc0cb999fc144bfb8d25174304cc094bc843e0ea6813e6dbfec4e0974bb35439a6e365e1096dafc535c25e2c805844fa5dcf017798fa154a19e8f8d30604563"},
	{"BLAKE2s", 1048585, "7806570657c87db5a25bba2c132c5662a3132132043bcf8a5cfd5a484040651e"},
	{"SM3", 1048585, "d23d8b821da2790d269f70042846b6afd95e258d8fa4f805fa619478c3b3fa02"},
	{"CRC32", 1048585, "e96ddb80"},
}

func TestReferenceVectors(t *testing.T) {
	for _, tc := range referenceVectors {
		t.Run(tc.algo+"/"+strconv.Itoa(tc.size), func(t *testing.T) {
			got, n, err := Hash(bytes.NewReader(testData(tc.size)), []string{tc.algo})
			if err != nil {
				t.Fatalf("Hash: %v", err)
			}
			if n != int64(tc.size) {
				t.Fatalf("read %d bytes, want %d", n, tc.size)
			}
			if got[tc.algo] != tc.want {
				t.Fatalf("%s = %s, want %s", tc.algo, got[tc.algo], tc.want)
			}
		})
	}
}

func TestStreamingMatchesOneShot(t *testing.T) {
	data := testData(1048585)
	chunkSizes := []int{1, 7, 63, 64, 65, 255, 4096}
	for _, algo := range Names() {
		want, _, err := Hash(bytes.NewReader(data), []string{algo})
		if err != nil {
			t.Fatal(err)
		}
		for _, chunk := range chunkSizes {
			spec := byName[strings.ToUpper(algo)]
			h := spec.New()
			for off := 0; off < len(data); off += chunk {
				end := off + chunk
				if end > len(data) {
					end = len(data)
				}
				if _, err := h.Write(data[off:end]); err != nil {
					t.Fatal(err)
				}
			}
			got := hex.EncodeToString(h.Sum(nil))
			if got != want[algo] {
				t.Fatalf("%s with chunk %d = %s, want %s", algo, chunk, got, want[algo])
			}
		}
	}
}

func TestHashMultipleAlgorithmsSinglePass(t *testing.T) {
	data := testData(1048585)
	all, n, err := Hash(bytes.NewReader(data), Names())
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(data)) {
		t.Fatalf("read %d, want %d", n, len(data))
	}
	if len(all) != len(Names()) {
		t.Fatalf("got %d digests, want %d", len(all), len(Names()))
	}
	for _, algo := range Names() {
		one, _, err := Hash(bytes.NewReader(data), []string{algo})
		if err != nil {
			t.Fatal(err)
		}
		if all[algo] != one[algo] {
			t.Fatalf("%s differs between multi and single pass", algo)
		}
	}
}

func TestResolve(t *testing.T) {
	specs, err := Resolve([]string{" sha256 ", "SHA3_256", "sha256", "blake2b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 {
		t.Fatalf("got %d specs, want 3", len(specs))
	}
	if specs[0].Name != "SHA256" || specs[1].Name != "SHA3_256" || specs[2].Name != "BLAKE2b" {
		t.Fatalf("unexpected canonical names: %+v", specs)
	}
	if _, err := Resolve([]string{"NOPE"}); err == nil {
		t.Fatal("expected error for unknown algorithm")
	}
	if _, err := Resolve([]string{"  ", ""}); err == nil {
		t.Fatal("expected error for empty algorithm list")
	}
}

func TestSM3SumDoesNotMutateState(t *testing.T) {
	h := NewSM3()
	_, _ = h.Write([]byte("abc"))
	first := hex.EncodeToString(h.Sum(nil))
	second := hex.EncodeToString(h.Sum(nil))
	if first != second {
		t.Fatalf("Sum is not stable: %s != %s", first, second)
	}
	_, _ = h.Write([]byte("def"))
	got := hex.EncodeToString(h.Sum(nil))

	want, _, err := Hash(strings.NewReader("abcdef"), []string{"SM3"})
	if err != nil {
		t.Fatal(err)
	}
	if got != want["SM3"] {
		t.Fatalf("continued digest = %s, want %s", got, want["SM3"])
	}
}
