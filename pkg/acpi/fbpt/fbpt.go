// Package fbpt reads Firmware Basic Performance Table within ACPI FPDT Table.

package fbpt

import (
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/u-root/u-root/pkg/acpi/fpdt"
	"github.com/u-root/u-root/pkg/uefivars"
)

const (
	FBPTStructureSig = "FBPT"
	memDevice        = "/dev/mem"

	// see ACPI Table Spec: https://uefi.org/sites/default/files/resources/ACPI%206_2_A_Sept29.pdf (page 208/page 212)
	EFI_ACPI_5_0_FPDT_PERFORMANCE_RECORD_HEADER_SIZE = 4
	EFI_ACPI_5_0_FBPT_HEADER_SIZE                    = 8

	// maximum number of FBPTPerfRecords to return in 'FindAllFBPTRecords'
	maxNumberOfFBPTPerfRecords = 2000

	FPDT_DYNAMIC_STRING_EVENT_RECORD_IDENTIFIER = 0x1011

	MODULE_START_ID            = 0x01
	MODULE_END_ID              = 0x02
	MODULE_LOADIMAGE_START_ID  = 0x03
	MODULE_LOADIMAGE_END_ID    = 0x04
	MODULE_DB_START_ID         = 0x05
	MODULE_DB_END_ID           = 0x06
	MODULE_DB_SUPPORT_START_ID = 0x07
	MODULE_DB_SUPPORT_END_ID   = 0x08
	MODULE_DB_STOP_START_ID    = 0x09
	MODULE_DB_STOP_END_ID      = 0x0A

	PERF_EVENTSIGNAL_START_ID = 0x10
	PERF_EVENTSIGNAL_END_ID   = 0x11
	PERF_CALLBACK_START_ID    = 0x20
	PERF_CALLBACK_END_ID      = 0x21
	PERF_FUNCTION_START_ID    = 0x30
	PERF_FUNCTION_END_ID      = 0x31
	PERF_INMODULE_START_ID    = 0x40
	PERF_INMODULE_END_ID      = 0x41
	PERF_CROSSMODULE_START_ID = 0x50
	PERF_CROSSMODULE_END_ID   = 0x51
)

var eventTypeMap = map[uint16]string{
	MODULE_START_ID:            "MODULE_START_ID",
	MODULE_END_ID:              "MODULE_END_ID",
	MODULE_LOADIMAGE_START_ID:  "MODULE_LOADIMAGE_START_ID",
	MODULE_LOADIMAGE_END_ID:    "MODULE_LOADIMAGE_END_ID",
	MODULE_DB_START_ID:         "MODULE_DB_START_ID",
	MODULE_DB_END_ID:           "MODULE_DB_END_ID",
	MODULE_DB_SUPPORT_START_ID: "MODULE_DB_SUPPORT_START_ID",
	MODULE_DB_SUPPORT_END_ID:   "MODULE_DB_SUPPORT_END_ID",
	MODULE_DB_STOP_START_ID:    "MODULE_DB_STOP_START_ID",
	MODULE_DB_STOP_END_ID:      "MODULE_DB_STOP_END_ID",

	PERF_EVENTSIGNAL_START_ID: "PERF_EVENTSIGNAL_START_ID",
	PERF_EVENTSIGNAL_END_ID:   "PERF_EVENTSIGNAL_END_ID",
	PERF_CALLBACK_START_ID:    "PERF_CALLBACK_START_ID",
	PERF_CALLBACK_END_ID:      "PERF_CALLBACK_END_ID",
	PERF_FUNCTION_START_ID:    "PERF_FUNCTION_START_ID",
	PERF_FUNCTION_END_ID:      "PERF_FUNCTION_END_ID",
	PERF_INMODULE_START_ID:    "PERF_INMODULE_START_ID",
	PERF_INMODULE_END_ID:      "PERF_INMODULE_END_ID",
	PERF_CROSSMODULE_START_ID: "PERF_CROSSMODULE_START_ID",
	PERF_CROSSMODULE_END_ID:   "PERF_CROSSMODULE_END_ID",
}

// based on struct definition found in edk2: /MdePkg/Include/IndustryStandard/Acpi50.h
type EFI_ACPI_5_0_FPDT_PERFORMANCE_RECORD_HEADER struct {
	Type     uint16
	Length   uint8
	Revision uint8
}

// based on struct definition found in edk2: /MdeModulePkg/Include/Guid/ExtendedFirmwarePerformance.h
type EFI_ACPI_6_5_FPDT_FIRMWARE_BASIC_BOOT_RECORD struct {
	PerformanceRecordHeader EFI_ACPI_5_0_FPDT_PERFORMANCE_RECORD_HEADER
	ResetEnd                uint64
	OSLoaderLoadImageStart  uint64
	OSLoaderStartImageStart uint64
	ExitBootServicesEntry   uint64
	ExitBootServicesExit    uint64
}

type MEASUREMENT_RECORD struct {
	HookType            string
	ProcessorIdentifier uint32
	Timestamp           uint64
	GUID                uefivars.MixedGUID
	Description         string
}

func verifyFBPTSignature(mem io.ReadSeeker, fbptAddr uint64) (uint32, error) {

	// Read & confirm FBPT struct signature
	if _, err := mem.Seek(int64(fbptAddr), io.SeekStart); err != nil {
		return 0, err
	}
	// Read as slices
	var fbptSig [4]byte
	if _, err := io.ReadFull(mem, fbptSig[:]); err != nil {
		return 0, err
	}

	if string(fbptSig[:]) != FBPTStructureSig {
		return 0, errors.New("FBPT structure signature check failed. Expected: FBPT, Got: " + string(fbptSig[:]))
	}

	var fbptLength [4]byte
	if _, err := io.ReadFull(mem, fbptLength[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(fbptLength[:]), nil
}

func FindAllFBPTRecords(FBPTAddr uint64) (int, []MEASUREMENT_RECORD, error) {

	var f *os.File
	var err error
	if f, err = os.OpenFile(memDevice, os.O_RDONLY, 0); err != nil {
		return 0, nil, err
	}
	defer f.Close()

	var tablelength uint32
	if tablelength, err = verifyFBPTSignature(f, FBPTAddr); err != nil {
		return 0, nil, err
	}

	// iterate through FBPT table
	var measurementRecords = make([]MEASUREMENT_RECORD, maxNumberOfFBPTPerfRecords)
	var index int
	var tableBytesRead uint32
	var HeaderInfo EFI_ACPI_5_0_FPDT_PERFORMANCE_RECORD_HEADER
	for tableBytesRead < (tablelength - EFI_ACPI_5_0_FBPT_HEADER_SIZE) && index < maxNumberOfFBPTPerfRecords{
		if HeaderInfo.Type, HeaderInfo.Length, _, err = fpdt.ReadFPDTRecordHeader(f); err != nil {
			return index, nil, err
		}
		if HeaderInfo.Type == FPDT_DYNAMIC_STRING_EVENT_RECORD_IDENTIFIER {
			if measurementRecords[index], err = readFirmwarePerformanceDataTableDynamicRecord(f, HeaderInfo.Length); err != nil {
				return index, nil, err
			}
			index++
		} else {
			if _, err := f.Seek(int64(HeaderInfo.Length-EFI_ACPI_5_0_FPDT_PERFORMANCE_RECORD_HEADER_SIZE), io.SeekCurrent); err != nil {
				return index, nil, err
			}
		}
		tableBytesRead += uint32(HeaderInfo.Length)
	}

	return index, measurementRecords, nil
}

func readFirmwarePerformanceDataTableDynamicRecord(mem io.ReadSeeker, recordLength uint8) (MEASUREMENT_RECORD, error) {
	var measurementRecord MEASUREMENT_RECORD
	var HookType [2]byte
	if _, err := io.ReadFull(mem, HookType[:]); err != nil {
		return measurementRecord, err
	}

	var ProcessorIdentifier [4]byte
	if _, err := io.ReadFull(mem, ProcessorIdentifier[:]); err != nil {
		return measurementRecord, err
	}

	var Timestamp [8]byte
	if _, err := io.ReadFull(mem, Timestamp[:]); err != nil {
		return measurementRecord, err
	}

	var Guid [16]byte
	if _, err := io.ReadFull(mem, Guid[:]); err != nil {
		return measurementRecord, err
	}

	String := make([]byte, recordLength-34)
	if _, err := io.ReadFull(mem, String[:]); err != nil {
		return measurementRecord, err
	}

	measurementRecord.HookType = eventTypeMap[binary.LittleEndian.Uint16(HookType[:])]
	measurementRecord.ProcessorIdentifier = binary.LittleEndian.Uint32(ProcessorIdentifier[:])
	measurementRecord.Timestamp = binary.LittleEndian.Uint64(Timestamp[:])
	measurementRecord.GUID = uefivars.MixedGUID(Guid)
	measurementRecord.Description = string(String[:])

	return measurementRecord, nil
}
