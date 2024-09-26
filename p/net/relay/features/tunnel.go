package features

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

type TunnelFlag uint32

const (
	TunnelPrivate TunnelFlag = 0x80000000
)

// TunnelID is an identification for tunnel.
//
//	+------------------+
//	|   ID   |  FLAG   |
//	+------------------+
//	|   16   |    4    |
//	+------------------+
//
//	ID: 16-byte tunnel ID value, should be a valid UUID.
//	FLAG: 4-byte flag, 0x80000000 for private tunnel.
type TunnelID [20]byte

var zeroTunnelID TunnelID

const tunnelIDLen = 16

func NewTunnelID(v []byte) (tid TunnelID) {
	copy(tid[:tunnelIDLen], v[:])
	return
}

func NewPrivateTunnelID(v []byte) (tid TunnelID) {
	copy(tid[:], v[:])
	binary.BigEndian.PutUint32(tid[tunnelIDLen:], uint32(TunnelPrivate))
	return
}

func (tid TunnelID) ID() (id [connectorIDLen]byte) {
	copy(id[:], tid[:tunnelIDLen])
	return
}

func (tid TunnelID) IsZero() bool {
	return bytes.Equal(tid[:tunnelIDLen], zeroTunnelID[:tunnelIDLen])
}

func (tid TunnelID) IsPrivate() bool {
	return binary.BigEndian.Uint32(tid[tunnelIDLen:])&uint32(TunnelPrivate) > 0
}

func (tid TunnelID) Equal(x TunnelID) bool {
	return bytes.Equal(tid[:tunnelIDLen], x[:tunnelIDLen])
}

func (tid TunnelID) String() string {
	var buf [36]byte
	encodeHex(buf[:], tid[:tunnelIDLen])
	return string(buf[:])
}

func encodeHex(dst []byte, v []byte) {
	hex.Encode(dst, v[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], v[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], v[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], v[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], v[10:])
}

type ConnectorFlag uint32

const (
	ConnectorUDP = 0x01
)

// ConnectorID is an identification for tunnel connection.
//
//	+------------------+
//	|   ID   |  FLAG   |
//	+------------------+
//	|   16   |    4    |
//	+------------------+
//
//	ID: 16-byte connector ID value, should be a valid UUID.
//	FLAG: 4-byte flag, 0x1 for UDP connector.
type ConnectorID [20]byte

const connectorIDLen = 16

var zeroConnectorID ConnectorID

func NewConnectorID(v []byte) (cid ConnectorID) {
	copy(cid[:connectorIDLen], v[:])
	return
}

func NewUDPConnectorID(v []byte) (cid ConnectorID) {
	copy(cid[:], v[:])
	binary.BigEndian.PutUint32(cid[connectorIDLen:], uint32(ConnectorUDP))
	return
}

func (cid ConnectorID) ID() (id [connectorIDLen]byte) {
	copy(id[:], cid[:connectorIDLen])
	return
}

func (cid ConnectorID) IsZero() bool {
	return bytes.Equal(cid[:connectorIDLen], zeroConnectorID[:connectorIDLen])
}

func (cid ConnectorID) IsUDP() bool {
	return binary.BigEndian.Uint32(cid[connectorIDLen:])&uint32(ConnectorUDP) > 0
}

func (cid ConnectorID) Equal(x ConnectorID) bool {
	return bytes.Equal(cid[:connectorIDLen], x[:connectorIDLen])
}

func (cid ConnectorID) String() string {
	var buf [36]byte
	encodeHex(buf[:], cid[:connectorIDLen])
	return string(buf[:])
}

// TunnelFeature is a relay feature,
//
// Protocol spec:
//
//	+---------------------+
//	| TUNNEL/CONNECTOR ID |
//	+---------------------+
//	|          16         |
//	+---------------------+
//
//	ID - 16-byte tunnel ID for request or connector ID for response.
type TunnelFeature struct {
	ID [tunnelIDLen]byte
}

func (f *TunnelFeature) Type() FeatureType {
	return FeatureTunnel
}

func (f *TunnelFeature) Encode() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(f.ID[:])
	return buf.Bytes(), nil
}

func (f *TunnelFeature) Decode(b []byte) error {
	if len(b) < tunnelIDLen {
		return ErrShortBuffer
	}
	copy(f.ID[:], b)
	return nil
}
