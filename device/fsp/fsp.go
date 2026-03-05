package fsp

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"time"

	"go.bug.st/serial"
)

// Fade Serial Protocol for updating variables inside the device's flash
// Inspired by the AN3155 protocol

// Each key may have a different datatype as written in the FSP documentation

const (
	CMD_SET  = 0x67
	CMD_GET  = 0x69
	CMD_DFU  = 0xDF
	CMD_SAVE = 0xFE

	ACK  = 0x0A
	NACK = 0xC4
)

// List of available keys to set
const (
	KEY_P_VALUE = uint16(iota)
	KEY_I_VALUE
	KEY_D_VALUE
	KEY_LEN
)

type SerialDevice struct {
	port serial.Port
}

func (p *SerialDevice) Close() {
	if p == nil {
		return
	}
	if p.port != nil {
		p.port.Close()
	}
}

func NewSerialDevice(portName string) (*SerialDevice, error) {
	port, err := serial.Open(portName, &serial.Mode{
		BaudRate: 115200,
	})

	if err != nil {
		return nil, err
	}

	port.SetReadTimeout(time.Second * 10)

	serDe := &SerialDevice{port}

	return serDe, nil
}

func (p *SerialDevice) sendCmd(cmd byte) error {
	if p == nil {
		slog.Error("Device not found")
		return fmt.Errorf("Device not found")
	}

	packet := []byte{cmd, cmd ^ 0xFF}
	_, err := p.port.Write(packet)

	return err
}

func (p *SerialDevice) waitForAck() error {
	if p == nil {
		slog.Error("Device not found")
		return fmt.Errorf("Device not found")
	}

	buffer := make([]byte, 1)

	err := p.readFull(buffer)
	if err != nil {
		slog.Error("Could not read port: %s", err.Error())
		return fmt.Errorf("Could not read port: %s", err.Error())
	}

	if buffer[0] != ACK {
		slog.Error("Expected ACK, got: 0x%X", buffer[0])
		return fmt.Errorf("Expected ACK, got: 0x%X", buffer[0])
	}

	return nil
}

func (p *SerialDevice) readFull(buf []byte) error {
	total := 0
	for total < len(buf) {
		n, err := p.port.Read(buf[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (p *SerialDevice) SetValue(key uint16, value float32) error {
	if p == nil {
		slog.Error("Device not found")
		return fmt.Errorf("Device not found")
	}

	if key >= KEY_LEN {
		return fmt.Errorf("Invalid key")
	}

	err := p.sendCmd(CMD_SET)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = p.waitForAck()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	keyBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(keyBytes, key)
	checksum := keyBytes[0] ^ keyBytes[1]

	packet := append(keyBytes, checksum)
	_, err = p.port.Write(packet)
	if err != nil {
		slog.Error("Could not send key: %s", err.Error())
		return fmt.Errorf("Could not send key: %s", err.Error())
	}

	err = p.waitForAck()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	valueBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(valueBytes, math.Float32bits(value))
	checksum = valueBytes[0] ^ valueBytes[1] ^ valueBytes[2] ^ valueBytes[3]

	packet = append(valueBytes, checksum)
	_, err = p.port.Write(packet)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("Could not send value: %s", err.Error())
	}

	err = p.sendCmd(CMD_SAVE)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = p.waitForAck()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return p.waitForAck()
}

func (p *SerialDevice) GetValue(key uint16) (float32, error) {
	if p == nil {
		slog.Error("Device not found")
		return 0.0, fmt.Errorf("Device not found")
	}

	if key >= KEY_LEN {
		return 0.0, fmt.Errorf("Invalid key")
	}

	err := p.sendCmd(CMD_GET)
	if err != nil {
		slog.Error(err.Error())
		return 0.0, err
	}

	err = p.waitForAck()
	if err != nil {
		slog.Error(err.Error())
		return 0.0, err
	}

	keyBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(keyBytes, key)
	checksum := keyBytes[0] ^ keyBytes[1]

	packet := append(keyBytes, checksum)
	_, err = p.port.Write(packet)
	if err != nil {
		slog.Error("Could not send key: %s", err.Error())
		return 0.0, fmt.Errorf("Could not send key: %s", err.Error())
	}

	err = p.waitForAck()
	if err != nil {
		slog.Error(err.Error())
		return 0.0, err
	}

	buffer := make([]byte, 5)
	err = p.readFull(buffer)
	if err != nil {
		slog.Error("Could not read port: %s", err.Error())
		return 0.0, fmt.Errorf("Could not read port: %s", err.Error())
	}

	if !checksumValid(buffer) {
		slog.Error("Invalid checksum")
		return 0.0, fmt.Errorf("Invalid checksum")
	}

	valUint32Format := binary.BigEndian.Uint32(buffer)

	return math.Float32frombits(valUint32Format), nil
}

func checksumValid(data []byte) bool {
	checksum := data[len(data)-1]

	var verify uint8 = 0x00
	for b := 0; b < len(data); b++ {
		verify = verify ^ data[b]
	}

	return verify == checksum
}
