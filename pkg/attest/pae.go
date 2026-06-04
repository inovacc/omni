package attest

import "strconv"

// PAE computes the DSSE Pre-Authentication Encoding (v1.0.2):
//
//	"DSSEv1" SP LEN(type) SP type SP LEN(body) SP body
//
// where SP is a single ASCII space and LEN is the base-10 byte length with no
// leading zeros. The result is the exact byte sequence to sign/verify.
func PAE(payloadType string, body []byte) []byte {
	out := make([]byte, 0, 6+len(payloadType)+len(body)+24)
	out = append(out, "DSSEv1"...)
	out = append(out, ' ')
	out = append(out, strconv.Itoa(len(payloadType))...)
	out = append(out, ' ')
	out = append(out, payloadType...)
	out = append(out, ' ')
	out = append(out, strconv.Itoa(len(body))...)
	out = append(out, ' ')
	out = append(out, body...)
	return out
}
