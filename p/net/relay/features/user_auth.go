package features

import (
	"bytes"
	"errors"
)

// UserAuthFeature is a relay feature,
// it contains the username and password for user authentication on server side.
//
// Protocol spec:
//
//	+------+----------+------+----------+
//	| ULEN |  UNAME   | PLEN |  PASSWD  |
//	+------+----------+------+----------+
//	|  1   | 0 to 255 |  1   | 1 to 255 |
//	+------+----------+------+----------+
//
//	ULEN - length of username field, 1 byte.
//	UNAME - username, variable length, 0 to 255 bytes, 0 means no username.
//	PLEN - length of password field, 1 byte.
//	PASSWD - password, variable length, 0 to 255 bytes, 0 means no password.
type UserAuthFeature struct {
	Username string
	Password string
}

func (f *UserAuthFeature) Type() FeatureType {
	return FeatureUserAuth
}

func (f *UserAuthFeature) Encode() ([]byte, error) {
	var buf bytes.Buffer

	ulen := len(f.Username)
	if ulen > 0xFF {
		return nil, errors.New("username maximum length exceeded")
	}
	buf.WriteByte(uint8(ulen))
	buf.WriteString(f.Username)

	plen := len(f.Password)
	if plen > 0xFF {
		return nil, errors.New("password maximum length exceeded")
	}
	buf.WriteByte(uint8(plen))
	buf.WriteString(f.Password)

	return buf.Bytes(), nil
}

func (f *UserAuthFeature) Decode(b []byte) error {
	if len(b) < 2 {
		return ErrShortBuffer
	}

	pos := 0
	ulen := int(b[pos])

	pos++
	if len(b) < pos+ulen+1 {
		return ErrShortBuffer
	}
	f.Username = string(b[pos : pos+ulen])

	pos += ulen
	plen := int(b[pos])

	pos++
	if len(b) < pos+plen {
		return ErrShortBuffer
	}
	f.Password = string(b[pos : pos+plen])

	return nil
}
