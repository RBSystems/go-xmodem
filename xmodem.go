package xmodem

import (
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/howeyc/crc16"
)

/*
Recieve takes a connection that has been already established and is
waiting for an XMODEM ACK (or 'C') command to begin XMODEM transmission.
Will return the byte array that is recieved via XMODEM. Uses XMODEM-CRC
*/
func Recieve(connection net.TCPConn) ([]byte, error) {

	//Do the first packet outside of regular recpetion loop, since it's recieval is
	//part of the transmission request
	firstPacket, err := requestTransmissionStart(connection)
	if err != nil {
		return []byte{}, err
	}

	//check the data for the valid crc - the first three bytes  are header information,
	//we'll check them next
	ok, err := checkCRC(firstPacket[4:])
	if err != nil {
		return []byte{}, err
	}
	if !ok {
		//we need to send a NAK
	}
	return []byte{}, nil
}

/*
checkCRC takes the last two bytes of data as the crc, calculates the crc on the preceding
bits and returns if they're equal
(i.e. )
*/
func checkCRC(data []byte) (bool, error) {

	if len(data) < 128 {
		return false, errors.New("Datablock is too small")
	}
	//get the last 2 bytes as the crc
	crcToCheck := data[len(data)-3:]

	sum := crc16.ChecksumCCITT(data[:len(data)-3])

	//convert to bytes for checking
	crcInt := binary.BigEndian.Uint16(crcToCheck)

	if sum == crcInt {
		return true, nil
	}

	return false, nil
}

/*
requestTransmissionStart sends the start character, with a timeout
until a packaet is recieved, the first packet is returned.
*/
func requestTransmissionStart(connection net.TCPConn) ([]byte, error) {
	maxAttempts := 10

	//Set the timeout to 4 seconds
	err := connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return []byte{}, err
	}

	written := 0
	attempts := 0
	//keep trying until either we write the bytes
	for written != 1 {
		//Ask for the CRC protocol
		written, err = connection.Write([]byte("C"))
		if err != nil {

			//type assertion to check if the error was a timeout
			if err, ok := err.(net.Error); ok && err.Timeout() {
				if attempts < maxAttempts {
					attempts++
					continue
				} else {
					return []byte{}, errors.New("Exceeded timeout attempts")
				}
			} else {
				return []byte{}, err
			}
		}
	}

	//Set the timeout to 4 seconds
	err = connection.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return []byte{}, err
	}

	//read the 3 header bytes plus the, 1024 bytes from the device, plus the 2 bit crc code.
	data := make([]byte, 1030)
	numRead, err := connection.Read(data)

	return data[:numRead], err
}
