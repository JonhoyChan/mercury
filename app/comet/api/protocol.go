package api

var ProtocolVersion int32 = 1

const (
	// MaxBodySize max proto body size
	MaxBodySize = uint32(1 << 12)
)

//const (
//	// size
//	_packageSize   = 4
//	_headerSize    = 2
//	_versionSize   = 2
//	_operationSize = 4
//	_rawHeaderSize = _packageSize + _headerSize + _versionSize + _operationSize
//	_maxPackSize   = MaxBodySize + uint32(_rawHeaderSize)
//	// offset
//	_packageOffset   = 0
//	_headerOffset    = _packageSize + _packageOffset
//	_versionOffset   = _headerSize + _headerOffset
//	_operationOffset = _versionSize + _versionOffset
//)
//
//var (
//	// ErrProtoPackLen proto packet len error
//	ErrProtoPackLen = ecode.NewError("default server codec pack length error")
//	// ErrProtoHeaderLen proto header len error
//	ErrProtoHeaderLen = ecode.NewError("default server codec header length error")
//	// ErrProtoVersion proto version error
//	ErrProtoVersion = ecode.NewError("default server codec protocol version error")
//)
//
//func Serialize(p *Proto) ([]byte, error) {
//	packLen := _rawHeaderSize + len(p.Body)
//
//	buf := new(bytes.Buffer)
//
//	var err error
//	if err = binary.Write(buf, binary.BigEndian, uint32(packLen)); err != nil {
//		return nil, err
//	}
//	if err = binary.Write(buf, binary.BigEndian, uint16(_rawHeaderSize)); err != nil {
//		return nil, err
//	}
//	if err = binary.Write(buf, binary.BigEndian, uint16(p.Version)); err != nil {
//		return nil, err
//	}
//	if err = binary.Write(buf, binary.BigEndian, uint32(p.Operation)); err != nil {
//		return nil, err
//	}
//	if p.Body != nil {
//		if err = binary.Write(buf, binary.BigEndian, p.Body); err != nil {
//			return nil, err
//		}
//	}
//
//	return buf.Bytes(), nil
//}
//
//func Deserialize(raw []byte) (*Proto, error) {
//	if len(raw) < _rawHeaderSize {
//		return nil, ErrProtoPackLen
//	}
//
//	packLen := binary.BigEndian.Uint32(raw[_packageOffset:_headerOffset])
//	if packLen < 0 || packLen > _maxPackSize {
//		return nil, ErrProtoPackLen
//	}
//	headerLen := binary.BigEndian.Uint16(raw[_headerOffset:_versionOffset])
//	if headerLen != _rawHeaderSize {
//		return nil, ErrProtoHeaderLen
//	}
//
//	p := &Proto{
//		Version:   int32(binary.BigEndian.Uint16(raw[_versionOffset:_operationOffset])),
//		Operation: int32(binary.BigEndian.Uint32(raw[_operationOffset:])),
//		Body:      nil,
//	}
//
//	if p.Version != ProtocolVersion {
//		return nil, ErrProtoVersion
//	}
//
//	if bodyLen := int(packLen - uint32(headerLen)); bodyLen > 0 {
//		p.Body = raw[headerLen:packLen]
//	}
//
//	return p, nil
//}
