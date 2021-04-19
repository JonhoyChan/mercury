package service

import (
	"bytes"
	"encoding/binary"
	"mercury/x/ecode"
	"mercury/x/types"
)

var ProtocolVersion uint8 = 1

const (
	// MaxBodySize max proto body size
	MaxBodySize = uint32(1 << 12)
)

const (
	// size
	_packageSize   = 4
	_headerSize    = 2
	_versionSize   = 2
	_operationSize = 4
	_rawHeaderSize = _packageSize + _headerSize + _versionSize + _operationSize
	_maxPackSize   = MaxBodySize + uint32(_rawHeaderSize)
	// offset
	_packageOffset   = 0
	_headerOffset    = _packageSize + _packageOffset
	_versionOffset   = _headerSize + _headerOffset
	_operationOffset = _versionSize + _versionOffset
)

var (
	// ErrProtoPackLen proto packet len error
	ErrProtoPackLen = ecode.NewError("default server codec pack length error")
	// ErrProtoHeaderLen proto header len error
	ErrProtoHeaderLen = ecode.NewError("default server codec header length error")
	// ErrProtoVersion proto version error
	ErrProtoVersion = ecode.NewError("default server codec protocol version error")
)

type Protocol struct {
	// operation for request
	Operation types.Operation `json:"operation" validate:"required"`
	// binary body bytes
	Body []byte `json:"body" validate:"required"`
}

func (p *Protocol) Validate() bool {
	if err := validate.Struct(p); err != nil {
		return false
	}
	return true
}

func (p *Protocol) Marshal() ([]byte, error) {
	return Serialize(p)
}

func (p *Protocol) Unmarshal(data []byte) error {
	return Deserialize(data, p)
}

func Serialize(p *Protocol) ([]byte, error) {
	packLen := _rawHeaderSize + len(p.Body)

	buf := new(bytes.Buffer)

	var err error
	if err = binary.Write(buf, binary.BigEndian, uint32(packLen)); err != nil {
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, uint16(_rawHeaderSize)); err != nil {
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, uint16(ProtocolVersion)); err != nil {
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, uint32(p.Operation)); err != nil {
		return nil, err
	}
	if p.Body != nil {
		if err = binary.Write(buf, binary.BigEndian, p.Body); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func Deserialize(raw []byte, p *Protocol) error {
	if len(raw) < _rawHeaderSize {
		return ErrProtoPackLen
	}

	packLen := binary.BigEndian.Uint32(raw[_packageOffset:_headerOffset])
	if packLen < 0 || packLen > _maxPackSize {
		return ErrProtoPackLen
	}
	headerLen := binary.BigEndian.Uint16(raw[_headerOffset:_versionOffset])
	if headerLen != _rawHeaderSize {
		return ErrProtoHeaderLen
	}
	version := uint8(binary.BigEndian.Uint16(raw[_versionOffset:_operationOffset]))

	if p == nil {
		p = &Protocol{}
	}
	p.Operation = types.Operation(binary.BigEndian.Uint32(raw[_operationOffset:]))

	if version != ProtocolVersion {
		return ErrProtoVersion
	}

	if bodyLen := int(packLen - uint32(headerLen)); bodyLen > 0 {
		p.Body = raw[headerLen:packLen]
	}

	return nil
}
