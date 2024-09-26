package features

import "encoding/binary"

type NetworkID uint16

func (p NetworkID) String() string {
	name := networkNames[p]
	if name == "" {
		name = networkNames[NetworkTCP]
	}
	return name
}

const (
	networkIDLen = 2
)

const (
	NetworkTCP    NetworkID = 0x0
	NetworkUDP    NetworkID = 0x1
	NetworkIP     NetworkID = 0x2
	NetworkUnix   NetworkID = 0x10
	NetworkSerial NetworkID = 0x11
)

var networkNames = map[NetworkID]string{
	NetworkTCP:    "tcp",
	NetworkUDP:    "udp",
	NetworkIP:     "ip",
	NetworkUnix:   "unix",
	NetworkSerial: "serial",
}

// NetworkFeature is a relay feature,
//
// Protocol spec:
//
//	+---------------------+
//	|       NETWORK       |
//	+---------------------+
//	|          2          |
//	+---------------------+
//
//	NETWORK - 2-byte network ID.
type NetworkFeature struct {
	Network NetworkID
}

func (f *NetworkFeature) Type() FeatureType {
	return FeatureNetwork
}

func (f *NetworkFeature) Encode() ([]byte, error) {
	var buf [networkIDLen]byte
	binary.BigEndian.PutUint16(buf[:], uint16(f.Network))
	return buf[:], nil
}

func (f *NetworkFeature) Decode(b []byte) error {
	if len(b) < networkIDLen {
		return ErrShortBuffer
	}
	f.Network = NetworkID(binary.BigEndian.Uint16(b))
	return nil
}
