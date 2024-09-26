package features

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

type AddrType uint8

const (
	AddrIPv4   AddrType = 1
	AddrDomain AddrType = 3
	AddrIPv6   AddrType = 4
)

// AddrFeature is a relay feature,
//
// Protocol spec:
//
//	+------+----------+----------+
//	| ATYP |   ADDR   |   PORT   |
//	+------+----------+----------+
//	|  1   | Variable |    2     |
//	+------+----------+----------+
//
//	ATYP - address type, 0x01 - IPv4, 0x03 - domain name, 0x04 - IPv6. 1 byte.
//	ADDR - host address, IPv4 (4 bytes), IPV6 (16 bytes) or doman name based on ATYP. For domain name, the first byte is the length of the domain name.
//	PORT - port number, 2 bytes.
type AddrFeature struct {
	AType AddrType
	Host  string
	Port  uint16
}

func (f *AddrFeature) Type() FeatureType {
	return FeatureAddr
}

func (f *AddrFeature) ParseFrom(address string) error {
	host, sport, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	port, err := strconv.Atoi(sport)
	if err != nil {
		port = 0
	}

	f.Host = host
	f.Port = uint16(port)
	f.AType = AddrDomain
	if ip := net.ParseIP(f.Host); ip != nil {
		if ip.To4() != nil {
			f.AType = AddrIPv4
		} else {
			f.AType = AddrIPv6
		}
	}

	return nil
}

func (f *AddrFeature) Encode() ([]byte, error) {
	var buf bytes.Buffer

	switch f.AType {
	case AddrIPv4:
		buf.WriteByte(byte(f.AType))
		ip4 := net.ParseIP(f.Host).To4()
		if ip4 == nil {
			ip4 = net.IPv4zero.To4()
		}
		buf.Write(ip4)
	case AddrDomain:
		buf.WriteByte(byte(f.AType))
		if len(f.Host) > 0xFF {
			return nil, errors.New("addr maximum length exceeded")
		}
		buf.WriteByte(uint8(len(f.Host)))
		buf.WriteString(f.Host)
	case AddrIPv6:
		buf.WriteByte(byte(f.AType))
		ip6 := net.ParseIP(f.Host).To16()
		if ip6 == nil {
			ip6 = net.IPv6zero.To16()
		}
		buf.Write(ip6)
	default:
		buf.WriteByte(byte(AddrIPv4))
		buf.Write(net.IPv4zero.To4())
	}

	var bp [2]byte
	binary.BigEndian.PutUint16(bp[:], f.Port)
	buf.Write(bp[:])

	return buf.Bytes(), nil
}

func (f *AddrFeature) Decode(b []byte) error {
	if len(b) < 4 {
		return ErrShortBuffer
	}

	f.AType = AddrType(b[0])
	pos := 1
	switch f.AType {
	case AddrIPv4:
		if len(b) < 3+net.IPv4len {
			return ErrShortBuffer
		}
		f.Host = net.IP(b[pos : pos+net.IPv4len]).String()
		pos += net.IPv4len
	case AddrIPv6:
		if len(b) < 3+net.IPv6len {
			return ErrShortBuffer
		}
		f.Host = net.IP(b[pos : pos+net.IPv6len]).String()
		pos += net.IPv6len
	case AddrDomain:
		alen := int(b[pos])
		if len(b) < 4+alen {
			return ErrShortBuffer
		}
		pos++
		f.Host = string(b[pos : pos+alen])
		pos += alen
	default:
		return ErrBadAddrType
	}

	f.Port = binary.BigEndian.Uint16(b[pos:])

	return nil
}
