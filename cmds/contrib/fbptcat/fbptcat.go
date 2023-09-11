// Copyright 2013-2023 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/u-root/u-root/pkg/acpi"
	"github.com/u-root/u-root/pkg/acpi/fbpt"
	"github.com/u-root/u-root/pkg/acpi/fpdt"
)

func main() {
	// Get FPDT table from ACPI
	var acpiFPDT acpi.Table = nil
	var err error
	if acpiFPDT, err = fpdt.ReadACPIFPDTTable(); err != nil {
		fmt.Println(err)
	}

	// Get FBPT Pointer from FPDT Table
	var FBPTAddr uint64
	if FBPTAddr, err = fpdt.FindFBPTTableAdrr(acpiFPDT); err != nil {
		log.Fatal(err)
	}

	var measurementRecords []fbpt.MEASUREMENT_RECORD
	if _, measurementRecords, err = fbpt.FindAllFBPTRecords(FBPTAddr); err != nil {
		log.Fatal(err)
	}

	for i, measurementRecord := range measurementRecords {
		if measurementRecord.Timestamp == 0 && len(measurementRecord.HookType) == 0 && len(measurementRecord.Description) == 0 {
			continue
		}
		fmt.Printf("Index: %d,Hook Type: %s, Processor Identifier/APIC ID: %d, Timestamp: %d, Guid: %s, Description: %s\n", i, measurementRecord.HookType, measurementRecord.ProcessorIdentifier, measurementRecord.Timestamp, measurementRecord.GUID.String(), measurementRecord.Description)
	}

}
