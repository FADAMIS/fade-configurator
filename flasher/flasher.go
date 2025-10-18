package flasher

import (
	"encoding/binary"
	"fmt"
	"io"
)

// protocol implementation: https://www.st.com/resource/en/application_note/an3155-usart-protocol-used-in-the-stm32-bootloader-stmicroelectronics.pdf
const (
	CMD_GET            = 0x00
	CMD_GET_VER        = 0x01
	CMD_GET_ID         = 0x02
	CMD_READ_MEM       = 0x11
	CMD_GO             = 0x21
	CMD_WRITE_MEM      = 0x31
	CMD_EXTENDED_ERASE = 0x44

	ACK  = 0x79
	NACK = 0x1F

	FLASH_ADDR_START = 0x08000000
)

func sendCmd(port io.ReadWriter, cmd byte) error {
	pkt := []byte{cmd, cmd ^ 0xFF}
	_, err := port.Write(pkt)
	return err
}

func waitAck(port io.Reader) error {
	buf := make([]byte, 1)

	port.Read(buf)

	if buf[0] != ACK {
		return fmt.Errorf("ACK expected, got 0x%X", buf[0])
	}

	return nil
}

func eraseMemory(port io.ReadWriter) error {
	sendCmd(port, CMD_EXTENDED_ERASE)

	err := waitAck(port)
	if err != nil {
		return fmt.Errorf("ACK expected, got 0x%X", err)
	}

	// 0xFFFF erases whole flash
	eraseBytes := []byte{0xFF, 0xFF}
	checksum := eraseBytes[0] ^ eraseBytes[1]

	port.Write(append(eraseBytes, checksum))

	err = waitAck(port)
	if err != nil {
		return fmt.Errorf("ACK expected, got 0x%X", err)
	}

	return nil
}

func writeMemory(port io.ReadWriter, addr uint32, data []byte) error {
	sendCmd(port, CMD_WRITE_MEM)

	err := waitAck(port)
	if err != nil {
		return fmt.Errorf("ACK expected, got 0x%X", err)
	}

	addrBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(addrBytes, addr)
	checksum := addrBytes[0] ^ addrBytes[1] ^ addrBytes[2] ^ addrBytes[3]

	port.Write(append(addrBytes, checksum))
	err = waitAck(port)
	if err != nil {
		return fmt.Errorf("ACK expected, got 0x%X", err)
	}

	n := byte(len(data) - 1)
	checksum = n
	for _, dataByte := range data {
		checksum ^= byte(dataByte)
	}

	pkt := append([]byte{n}, data...)
	port.Write(append(pkt, checksum))
	err = waitAck(port)
	if err != nil {
		return fmt.Errorf("ACK expected, got 0x%X", err)
	}

	return nil
}

func flashFirmware(port io.ReadWriter, data []byte) error {
	const chunkSize = 256

	for offset := 0; offset < len(data); offset += chunkSize {
		dataEnd := offset + chunkSize
		if dataEnd > len(data) {
			dataEnd = len(data)
		}

		chunk := data[offset:dataEnd]

		addr := FLASH_ADDR_START + offset

		err := writeMemory(port, uint32(addr), chunk)
		if err != nil {
			return fmt.Errorf("error at %X: %v", addr, err)
		}
	}

	return nil
}
