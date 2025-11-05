package dfu

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

// Protocol implemented based on AN3156 and USB DFU specification

const (
	DFU_INTERFACE     = 0
	DFU_TRANSFER_SIZE = 2048

	DFU_DETACH    = 0x00
	DFU_DNLOAD    = 0x01
	DFU_UPLOAD    = 0x02
	DFU_GETSTATUS = 0x03
	DFU_CLRSTATUS = 0x04
	DFU_GETSTATE  = 0x05

	DFU_CMD_SET_ADDR_PTR = 0x21
	DFU_CMD_ERASE        = 0x41

	DFU_REQ_TYPE_HOST_TO_DEVICE = 0x21
	DFU_REQ_TYPE_DEVICE_TO_HOST = 0xa1

	DFU_FLASH_START_ADDRESS = 0x08000000

	dfuIDLE        = 0x02
	dfuDNLOAD_IDLE = 0x05
	dfuERROR       = 0x0a

	STM32_VID = 0x0483
	STM32_PID = 0xdf11
)

type DFUDevice struct {
	Context      *gousb.Context
	Device       *gousb.Device
	Interface    uint16
	TransferSize int
	Address      uint32
}

func NewDFUDevice() (*DFUDevice, error) {
	ctx := gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(STM32_VID, STM32_PID)
	if err != nil {
		return nil, fmt.Errorf("Failed to open device: %w", err)
	}
	if dev == nil {
		return nil, fmt.Errorf("Device not found")
	}

	device := &DFUDevice{
		Context:      ctx,
		Device:       dev,
		Interface:    DFU_INTERFACE,
		TransferSize: DFU_TRANSFER_SIZE,
		Address:      DFU_FLASH_START_ADDRESS,
	}

	return device, nil
}

func (d *DFUDevice) Close() {
	if d == nil {
		return
	}
	if d.Device != nil {
		d.Device.Close()
	}
	if d.Context != nil {
		d.Context.Close()
	}
}

func (d *DFUDevice) controlIn(bRequest uint8, wValue uint16, wLength int) ([]byte, error) {
	buf := make([]byte, wLength)
	n, err := d.Device.Control(
		DFU_REQ_TYPE_DEVICE_TO_HOST,
		bRequest,
		wValue,
		d.Interface,
		buf,
	)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func (d *DFUDevice) controlOut(bRequest uint8, wValue uint16, data []byte) error {
	_, err := d.Device.Control(
		DFU_REQ_TYPE_HOST_TO_DEVICE,
		bRequest,
		wValue,
		d.Interface,
		data,
	)

	return err
}

func (d *DFUDevice) GetStatus() ([]byte, time.Duration, error) {
	data, err := d.controlIn(DFU_GETSTATUS, 0, 6)
	if err != nil {
		return nil, 0, err
	}
	if len(data) < 6 {
		return nil, 0, fmt.Errorf("Invalid response")
	}

	timeout := uint32(data[1]) | uint32(data[2])<<8 | uint32(data[3])<<16
	return data, time.Duration(timeout) * time.Millisecond, nil
}

func (d *DFUDevice) WaitForIdle() error {
	for {
		data, timeout, err := d.GetStatus()
		if err != nil {
			return err
		}

		if data[0] != 0x00 {
			return fmt.Errorf("DFU Error")
		}

		if data[4] == dfuIDLE || data[4] == dfuDNLOAD_IDLE {
			return nil
		}

		time.Sleep(timeout)
	}
}

func (d *DFUDevice) GetState() (byte, error) {
	data, err := d.controlIn(DFU_GETSTATE, 0, 1)
	if err != nil {
		return 0, err
	}

	return data[0], nil
}

func (d *DFUDevice) ClearStatus() error {
	return d.controlOut(DFU_CLRSTATUS, 0, nil)
}

func (d *DFUDevice) EnsureReady() error {
	for {
		state, err := d.GetState()
		if err != nil {
			return err
		}

		switch state {
		case dfuIDLE, dfuDNLOAD_IDLE:
			return nil
		case dfuERROR:
			if err := d.ClearStatus(); err != nil {
				return err
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (d *DFUDevice) EraseFlash() error {
	if err := d.controlOut(DFU_DNLOAD, 0, []byte{DFU_CMD_ERASE}); err != nil {
		return err
	}

	err := d.WaitForIdle()
	return err
}

func (d *DFUDevice) LeaveDFU() error {
	err := d.controlOut(DFU_DNLOAD, 0, nil)
	return err
}

func (d *DFUDevice) FlashFirmware(firmware []byte, progress func(string)) error {
	err := d.EnsureReady()
	if err != nil {
		return err
	}

	progress("Erasing flash...")

	err = d.EraseFlash()
	if err != nil {
		return err
	}

	addrCmd := []byte{
		DFU_CMD_SET_ADDR_PTR,
		byte(d.Address >> 0),
		byte(d.Address >> 8),
		byte(d.Address >> 16),
		byte(d.Address >> 24),
	}

	if err := d.controlOut(DFU_DNLOAD, 0, addrCmd); err != nil {
		return err
	}

	if err := d.WaitForIdle(); err != nil {
		return err
	}

	progress("Flashing firmware...")

	block := 2
	for offset := 0; offset < len(firmware); offset += d.TransferSize {
		end := offset + d.TransferSize
		if end > len(firmware) {
			end = len(firmware)
		}

		chunk := firmware[offset:end]

		if err := d.controlOut(DFU_DNLOAD, uint16(block), chunk); err != nil {
			return err
		}

		if err := d.WaitForIdle(); err != nil {
			return err
		}

		block++
	}

	return d.WaitForIdle()
}
