package document

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	// CodecMagic is the first byte identifying a SearchForge serialized document.
	CodecMagic byte = 0x53 // ASCII 'S'
	// CodecVersion is the current binary format version.
	CodecVersion uint8 = 1
)

var (
	ErrCorruptData       = errors.New("document: corrupt binary data")
	ErrUnsupportedVersion = errors.New("document: unsupported codec version")
	ErrUnexpectedEOF     = errors.New("document: unexpected end of buffer")
)

// Flag bitmasks for Field options
const (
	flagStored  uint8 = 1 << 0
	flagIndexed uint8 = 1 << 1
)

// MarshalBinary serializes a Document into a deterministic, compact binary format.
// It implements the standard encoding.BinaryMarshaler interface.
func (d *Document) MarshalBinary() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid document: %w", err)
	}

	// Calculate exact buffer size upfront to perform a single heap allocation
	size := 1 + 1 // Magic + Version
	size += 2 + len(d.ID) // ID length (uint16) + ID bytes
	size += 2             // Field count (uint16)

	for _, f := range d.fields {
		size += 2 + len(f.Name) // Name length (uint16) + Name bytes
		size += 1               // Type (uint8)
		size += 1               // Flags (uint8)

		switch f.Type {
		case TypeText, TypeKeyword:
			strVal := f.Value.(string)
			size += 4 + len(strVal) // Value length (uint32) + UTF-8 bytes
		case TypeNumeric:
			size += 8 // float64 (uint64 bits)
		case TypeTimestamp:
			size += 8 // int64 (Unix nanoseconds)
		case TypeBoolean:
			size += 1 // uint8 (0 or 1)
		default:
			return nil, fmt.Errorf("%w: unknown field type %v", ErrInvalidFieldType, f.Type)
		}
	}

	buf := make([]byte, size)
	offset := 0

	// 1. Magic and Version
	buf[offset] = CodecMagic
	offset++
	buf[offset] = CodecVersion
	offset++

	// 2. Document ID
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(d.ID)))
	offset += 2
	copy(buf[offset:], d.ID)
	offset += len(d.ID)

	// 3. Field Count
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(d.fields)))
	offset += 2

	// 4. Fields
	for _, f := range d.fields {
		// Field Name
		binary.BigEndian.PutUint16(buf[offset:], uint16(len(f.Name)))
		offset += 2
		copy(buf[offset:], f.Name)
		offset += len(f.Name)

		// Field Type
		buf[offset] = uint8(f.Type)
		offset++

		// Flags
		var flags uint8
		if f.Stored {
			flags |= flagStored
		}
		if f.Indexed {
			flags |= flagIndexed
		}
		buf[offset] = flags
		offset++

		// Field Value
		switch f.Type {
		case TypeText, TypeKeyword:
			strVal := f.Value.(string)
			binary.BigEndian.PutUint32(buf[offset:], uint32(len(strVal)))
			offset += 4
			copy(buf[offset:], strVal)
			offset += len(strVal)

		case TypeNumeric:
			numVal, err := castToFloat64(f.Value)
			if err != nil {
				return nil, err
			}
			bits := math.Float64bits(numVal)
			binary.BigEndian.PutUint64(buf[offset:], bits)
			offset += 8

		case TypeTimestamp:
			tVal := f.Value.(time.Time).UTC()
			binary.BigEndian.PutUint64(buf[offset:], uint64(tVal.UnixNano()))
			offset += 8

		case TypeBoolean:
			if f.Value.(bool) {
				buf[offset] = 1
			} else {
				buf[offset] = 0
			}
			offset++
		}
	}

	return buf, nil
}

// UnmarshalBinary reconstructs a Document from its binary representation.
// It implements the standard encoding.BinaryUnmarshaler interface.
func (d *Document) UnmarshalBinary(data []byte) error {
	if len(data) < 6 { // Min: Magic(1) + Version(1) + IDLen(2) + FieldCount(2)
		return ErrUnexpectedEOF
	}

	offset := 0

	// 1. Validate Magic
	if data[offset] != CodecMagic {
		return fmt.Errorf("%w: invalid magic byte 0x%x", ErrCorruptData, data[offset])
	}
	offset++

	// 2. Validate Version
	version := data[offset]
	if version != CodecVersion {
		return fmt.Errorf("%w: version %d", ErrUnsupportedVersion, version)
	}
	offset++

	// 3. Document ID
	if offset+2 > len(data) {
		return ErrUnexpectedEOF
	}
	idLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2

	if offset+idLen > len(data) {
		return ErrUnexpectedEOF
	}
	docID := string(data[offset : offset+idLen])
	offset += idLen

	// 4. Field Count
	if offset+2 > len(data) {
		return ErrUnexpectedEOF
	}
	fieldCount := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2

	// Reinitialize Document struct
	d.ID = docID
	d.fields = make([]Field, 0, fieldCount)
	d.index = make(map[string]int, fieldCount)

	// 5. Decode Fields
	for i := 0; i < fieldCount; i++ {
		// Field Name
		if offset+2 > len(data) {
			return ErrUnexpectedEOF
		}
		nameLen := int(binary.BigEndian.Uint16(data[offset:]))
		offset += 2

		if offset+nameLen > len(data) {
			return ErrUnexpectedEOF
		}
		fieldName := string(data[offset : offset+nameLen])
		offset += nameLen

		// Field Type & Flags
		if offset+2 > len(data) {
			return ErrUnexpectedEOF
		}
		fType := FieldType(data[offset])
		offset++
		flags := data[offset]
		offset++

		stored := (flags & flagStored) != 0
		indexed := (flags & flagIndexed) != 0

		var val any
		switch fType {
		case TypeText, TypeKeyword:
			if offset+4 > len(data) {
				return ErrUnexpectedEOF
			}
			valLen := int(binary.BigEndian.Uint32(data[offset:]))
			offset += 4
			if offset+valLen > len(data) {
				return ErrUnexpectedEOF
			}
			val = string(data[offset : offset+valLen])
			offset += valLen

		case TypeNumeric:
			if offset+8 > len(data) {
				return ErrUnexpectedEOF
			}
			bits := binary.BigEndian.Uint64(data[offset:])
			offset += 8
			val = math.Float64frombits(bits)

		case TypeTimestamp:
			if offset+8 > len(data) {
				return ErrUnexpectedEOF
			}
			nanos := int64(binary.BigEndian.Uint64(data[offset:]))
			offset += 8
			val = time.Unix(0, nanos).UTC()

		case TypeBoolean:
			if offset+1 > len(data) {
				return ErrUnexpectedEOF
			}
			val = data[offset] == 1
			offset++

		default:
			return fmt.Errorf("%w: unknown field type id %d", ErrCorruptData, fType)
		}

		field := Field{
			Name:    fieldName,
			Type:    fType,
			Value:   val,
			Stored:  stored,
			Indexed: indexed,
		}

		if err := d.AddField(field); err != nil {
			return fmt.Errorf("%w: cannot add field during unmarshal: %v", ErrCorruptData, err)
		}
	}

	return nil
}

// EncodeDocument is a package helper returning the serialized byte slice of a document.
func EncodeDocument(doc *Document) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("document: cannot encode nil document")
	}
	return doc.MarshalBinary()
}

// DecodeDocument is a package helper deserializing raw bytes into a new Document.
func DecodeDocument(data []byte) (*Document, error) {
	doc := &Document{}
	if err := doc.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return doc, nil
}

func castToFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("%w: cannot cast %T to float64", ErrInvalidFieldValue, v)
	}
}

