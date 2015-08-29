package murmur

import "unsafe"

// Constants defined by the Murmur3 algorithm
const (
	_C1 = uint32(0xcc9e2d51)
	_C2 = uint32(0x1b873593)
	_F1 = uint32(0x85ebca6b)
	_F2 = uint32(0xc2b2ae35)
)

// A default seed for Murmur3
const M3Seed = uint32(0x9747b28c)

// Generates a Murmur3 Hash [http://code.google.com/p/smhasher/wiki/MurmurHash3]
// Does not generate intermediate objects.
func Murmur3(data []byte, seed uint32) uint32 {
	h1 := seed
	ldata := len(data)
	end := ldata - (ldata % 4)
	i := 0

	// Inner
	for ; i < end; i += 4 {
		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
		k1 *= _C1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= _C2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*5 + 0xe6546b64
	}

	// Tail
	var k1 uint32
	switch ldata - i {
	case 3:
		k1 |= uint32(data[i+2]) << 16
		fallthrough
	case 2:
		k1 |= uint32(data[i+1]) << 8
		fallthrough
	case 1:
		k1 |= uint32(data[i])
		k1 *= _C1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= _C2
		h1 ^= k1
	}

	// Finalization
	h1 ^= uint32(ldata)
	h1 ^= (h1 >> 16)
	h1 *= _F1
	h1 ^= (h1 >> 13)
	h1 *= _F2
	h1 ^= (h1 >> 16)

	return h1
}
