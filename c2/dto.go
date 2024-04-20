package main

import (
	"encoding/json"
	"log"
)

type TransferPacket struct {
	Header  string
	Payload string
}

func decodeTransferPacket(inboundData []byte) TransferPacket {
	transferPacket := TransferPacket{}
	err := json.Unmarshal(inboundData, &transferPacket)
	if err != nil {
		log.Printf("err decodeTransferPacket: %s", err)
	}
	return transferPacket
}

func encodeTransferPacket(header string, payload string) []byte {
	transferPacket := TransferPacket{
		Header:  header,
		Payload: payload,
	}
	transferPayloadEncoded, err := json.Marshal(transferPacket)
	if err != nil {
		log.Print("err encodeTransferPacket:", err)
	}
	return transferPayloadEncoded
}
