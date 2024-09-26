package features

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrShortBuffer = errors.New("short buffer")
	ErrBadAddrType = errors.New("bad address type")
)

const (
	featureHeaderLen = 3
)

type FeatureType uint8

const (
	FeatureUserAuth FeatureType = 0x01
	FeatureAddr     FeatureType = 0x02
	FeatureTunnel   FeatureType = 0x03
	FeatureNetwork  FeatureType = 0x04
)

// Feature represents a feature the client or server owned.
//
// Protocol spec:
//
//	+------+----------+------+
//	| TYPE |  LEN  | FEATURE |
//	+------+-------+---------+
//	|  1   |   2   |   VAR   |
//	+------+-------+---------+
//
//	TYPE - feature type, 1 byte.
//	LEN - length of feature data, 2 bytes.
//	FEATURE - feature data.
type Feature interface {
	Type() FeatureType
	Encode() ([]byte, error)
	Decode([]byte) error
}

func ReadFeature(r io.Reader) (Feature, error) {
	var header [featureHeaderLen]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	b := make([]byte, int(binary.BigEndian.Uint16(header[1:3])))
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return NewFeature(FeatureType(header[0]), b)
}

func NewFeature(t FeatureType, data []byte) (f Feature, err error) {
	switch t {
	case FeatureUserAuth:
		f = new(UserAuthFeature)
	case FeatureAddr:
		f = new(AddrFeature)
	case FeatureTunnel:
		f = new(TunnelFeature)
	case FeatureNetwork:
		f = new(NetworkFeature)
	default:
		return nil, errors.New("unknown feature")
	}
	err = f.Decode(data)
	return
}
