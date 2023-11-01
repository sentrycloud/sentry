package protocol

import (
	"encoding/binary"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"sync/atomic"
)

const (
	MagicNumber = 0x1F2E3C4D
	PduVersion  = 0x01
	PduHeadSize = 16

	PduTypeMetrics = 1
)

var (
	seqNum uint32 = 0
	Json          = jsoniter.ConfigFastest
)

type PduHeader struct {
	Magic         uint32
	Version       uint16
	Protocol      uint16
	Sequence      uint32
	PayloadLength uint32
}

func serializePduHeader(data []byte, protocol uint16, payloadLength uint32) {
	binary.BigEndian.PutUint32(data[0:4], MagicNumber)
	binary.BigEndian.PutUint16(data[4:6], PduVersion)
	binary.BigEndian.PutUint16(data[6:8], protocol)
	binary.BigEndian.PutUint32(data[8:12], atomic.AddUint32(&seqNum, 1))
	binary.BigEndian.PutUint32(data[12:16], payloadLength)
}

func SerializeMetricValues(values []MetricValue) ([]byte, error) {
	payload, err := Json.Marshal(values)
	if err != nil {
		return nil, err
	}

	payloadLength := len(payload)
	data := make([]byte, PduHeadSize+payloadLength)
	serializePduHeader(data, PduTypeMetrics, uint32(payloadLength))
	copy(data[PduHeadSize:], payload)
	return data, nil
}

func DeserializePduHeader(data []byte) (PduHeader, error) {
	var pduHeader = PduHeader{}
	if len(data) != PduHeadSize {
		return pduHeader, errors.New("pdu header length not match")
	}

	pduHeader.Magic = binary.BigEndian.Uint32(data[0:4])
	pduHeader.Version = binary.BigEndian.Uint16(data[4:6])
	pduHeader.Protocol = binary.BigEndian.Uint16(data[6:8])
	pduHeader.Sequence = binary.BigEndian.Uint32(data[8:12])
	pduHeader.PayloadLength = binary.BigEndian.Uint32(data[12:16])

	if pduHeader.Magic != MagicNumber {
		return pduHeader, errors.New("magic number not match")
	}
	return pduHeader, nil
}

func UnmarshalPayload(data []byte) ([]MetricValue, error) {
	var metrics []MetricValue
	err := Json.Unmarshal(data, &metrics)
	return metrics, err
}
