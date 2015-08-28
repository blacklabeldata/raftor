package main

import "hash/fnv"

// Constants for FNV1A and derivatives
const (
	_OFF32 = 2166136261
	_P32   = 16777619
	_YP32  = 709607
)

// Constants for multiples of sizeof(WORD)
const (
	_WSZ    = 4         // 4
	_DWSZ   = _WSZ << 1 // 8
	_DDWSZ  = _WSZ << 2 // 16
	_DDDWSZ = _WSZ << 3 // 32
)

// Jesteress derivative of FNV1A from [http://www.sanmayce.com/Fastest_Hash/]
func Jesteress(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

// func Jesteress(data []byte) uint32 {
// 	h32 := uint32(_OFF32)
// 	i, dlen := 0, len(data)

// 	for ; dlen >= _DDWSZ; dlen -= _DDWSZ {
// 		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
// 		k2 := *(*uint64)(unsafe.Pointer(&data[i+4]))
// 		h32 = uint32((uint64(h32) ^ ((k1<<5 | k1>>27) ^ k2)) * _YP32)
// 		i += _DDWSZ
// 	}

// 	// Cases: 0,1,2,3,4,5,6,7
// 	if (dlen & _DWSZ) > 0 {
// 		k1 := *(*uint64)(unsafe.Pointer(&data[i]))
// 		h32 = uint32(uint64(h32)^k1) * _YP32
// 		i += _DWSZ
// 	}
// 	if (dlen & _WSZ) > 0 {
// 		k1 := *(*uint32)(unsafe.Pointer(&data[i]))
// 		h32 = (h32 ^ k1) * _YP32
// 		i += _WSZ
// 	}
// 	if (dlen & 1) > 0 {
// 		h32 = (h32 ^ uint32(data[i])) * _YP32
// 	}
// 	return h32 ^ (h32 >> 16)
// }
