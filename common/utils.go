package common

import "crypto/sha256"

func GenerateCertChainHash(rawCerts [][]byte) (chainHash []byte) {
	for _, cert := range rawCerts {
		certHash := sha256.Sum256(cert)
		if chainHash == nil {
			chainHash = certHash[:]
		} else {
			newHash := sha256.Sum256(append(chainHash, certHash[:]...))
			chainHash = newHash[:]
		}
	}
	return chainHash
}

func Deduplicate[T comparable](list []T) []T {
	if list == nil {
		return nil
	}
	res := make([]T, 0, len(list))
	m := make(map[T]struct{})
	for _, v := range list {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		res = append(res, v)
	}
	return res
}
