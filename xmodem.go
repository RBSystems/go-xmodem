package xmodem

import (
	"errors"
	"net"
	"time"
)

/*
Recieve takes a connection that has been already established and is
waiting for an XMODEM ACK (or 'C') command to begin XMODEM transmission.
Will return the byte array that is recieved via XMODEM. Uses XMODEM-CRC
*/
func Recieve(connection net.TCPConn) ([]byte, error) {
	firstPacket, numRead, err := requestTransmissionStart(connection)
	if err != nil {
		return []byte{}, err
	}

	return []byte{}, nil
}

/*
requestTransmissionStart sends the start character, with a timeout
until a packaet is recieved, the first packet is returned.
*/
func requestTransmissionStart(connection net.TCPConn) ([]byte, int, error) {
	maxAttempts := 10

	//Set the timeout to 4 seconds
	err := connection.SetWriteDeadline(time.Now().Add(4 * time.Second))
	if err != nil {
		return []byte{}, 0, err
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
					return []byte{}, 0, errors.New("Exceeded timeout attempts")
				}
			} else {
				return []byte{}, 0, err
			}
		}
	}

	//Set the timeout to 4 seconds
	err = connection.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return []byte{}, 0, err
	}

	//read the 3 header bytes plus the, 1024 bytes from the device, plus the 2 bit crc code.
	data := make([]byte, 1030)
	numRead, err := connection.Read(data)

	return data, numRead, err
}
