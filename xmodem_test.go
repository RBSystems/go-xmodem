package xmodem

import (
	"encoding/binary"
	"io/ioutil"
	"testing"
)

func TestCRC(t *testing.T) {

	//read in the binary file
	bytes, err := ioutil.ReadFile("./XModemTestPacket1.bin")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	crc := calcCRC(bytes[3 : len(bytes)-2])

	if crc != binary.BigEndian.Uint16(bytes[len(bytes)-2:]) {
		t.Error("CRC calculation incorrect.")
		t.FailNow()
	}

	correct, err := checkCRC(bytes[3:])

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !correct {
		t.Error("CRC comparison incorrect.")
		t.FailNow()
	}

}
