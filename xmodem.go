package xmodem

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// crctab calculated by Mark G. Mendel, Network Systems Corporation
var crctable = []uint16{0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
	0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
	0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
	0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
	0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
	0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
	0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
	0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
	0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
	0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
	0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
	0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
	0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
	0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
	0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
	0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
	0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
	0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
	0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
	0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
	0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
	0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
	0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
	0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
	0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
	0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
	0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
	0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
	0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
	0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0,
}

//STX is the XMODEM-1k header start code
var STX = byte(0x02)

//NAK is the non-acknowledge code
var NAK = byte(0x15)

//ACK is the acknowledge code
var ACK = byte(0x06)

//EOT is the Ent of Transmission code
var EOT = byte(0x04)

//ETB is the End of transmission Block code
var ETB = byte(0x17)

//CAN is the cancel code
var CAN = byte(0x18)

/*
Receive takes a connection that has been already established and is
waiting for an XMODEM ACK (or 'C') command to begin XMODEM transmission.
Will return the byte array that is recieved via XMODEM. Uses XMODEM-CRC
*/
func Receive(connection net.Conn) ([]byte, error) {
	log.Printf("Starting XModem Receive.")

	//Do the first packet outside of regular recpetion loop, since it's reciept is
	//part of the transmission request
	log.Printf("Requesting transmission start.")
	firstPacket, err := requestTransmissionStart(connection)
	if err != nil {
		return []byte{}, err
	}

	transmitting := true
	message := []byte{}
	curBlock := firstPacket
	blockCount := 1
	log.Printf("First Packet Received. %s", firstPacket[:15])

	for transmitting == true {
		log.Printf("Reading Packet.")
		//check the data for the valid crc - the first three bytes  are header information,
		//we'll check them next
		ok, err := checkCRC(curBlock[3:])
		if err != nil {
			return []byte{}, err
		}
		curCount := curBlock[2]
		if !ok || curCount != byte(blockCount) {
			log.Printf("Not enough bytes, sending NAK")

			//send a NAK code
			_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_, err = connection.Write([]byte{NAK})

			if err != nil {
				log.Printf("Not enough bytes, sending NAK")
				return []byte{}, err
			}
		} else {
			message = append(message, curBlock[3:len(curBlock)-2]...)
			_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))

			_, err = connection.Write([]byte{ACK})
			if err != nil {
				return []byte{}, err
			}
		}

		_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))

		//read the next bytes.
		err = connection.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			return []byte{}, err
		}

		_, err = connection.Read(curBlock)
		if err != nil {
			log.Printf("Read timeout")
			return []byte{}, err
		}
		switch curBlock[0] {
		case STX:
			blockCount++
			continue
		case EOT:
			transmitting = false
			break
		}
	}

	_, err = connection.Write([]byte{ACK})
	if err != nil {
		log.Printf("Write timeout.")
		return []byte{}, err
	}
	_, err = connection.Read(curBlock)
	if err != nil {
		log.Printf("Read timeout.")
		return []byte{}, err
	}
	if curBlock[0] == ETB || curBlock[0] == EOT {

		_, err = connection.Write([]byte{ACK})

		if err != nil {
			log.Printf("Write timeout.")
			return []byte{}, err
		}

		return message, nil
	}

	return message, nil
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
	crcToCheck := data[len(data)-2:]
	//fmt.Printf("%v\n", len(data)-2)

	sum := calcCRC(data[:len(data)-2])

	//fmt.Printf("%b\n", sum)
	//fmt.Printf("%b\n", crcToCheck)
	//convert to bytes for checking
	crcInt := binary.BigEndian.Uint16(crcToCheck)

	if sum == crcInt {
		return true, nil
	}

	return false, nil
}

func calcCRC(data []byte) uint16 {
	var crc uint16
	crc = 0
	b := make([]byte, 2)

	for _, char := range data {

		binary.BigEndian.PutUint16(b, crc)
		//fmt.Printf("crc: %x\n", b)
		//fmt.Printf("b[0]: %x\n", b[0])
		//fmt.Printf("char: %x\n", char)
		crctblidx := (b[0] ^ char) & 0xff
		//fmt.Printf("indx: %x\n", crctblidx)
		crc = ((crc << 8) ^ crctable[crctblidx]) & 0xffff
	}
	return crc & 0xffff
}

/*
requestTransmissionStart sends the start character, with a timeout
until a packaet is recieved, the first packet is returned.
*/
func requestTransmissionStart(connection net.Conn) ([]byte, error) {
	maxAttempts := 10

	//Set the timeout to 10 seconds
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

	//Set the timeout to 10 seconds
	err = connection.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return []byte{}, err
	}

	//read the 3 header bytes plus the, 1024 bytes from the device, plus the 2 byte crc code.
	data := make([]byte, 1030)
	numRead, err := connection.Read(data)
	if err != nil {
		fmt.Printf("%+v", err.Error())

		return []byte{}, err
	}

	return data[:numRead], nil
}
