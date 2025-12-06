package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sewiti/rfid-app/rfid"
)

var silent bool

func emitOK(d *rfid.Device)    { beepWithLed(d, 50*time.Millisecond, rfid.LedGreen) }
func emitError(d *rfid.Device) { beepWithLed(d, 200*time.Millisecond, rfid.LedRed) }

func beepWithLed(d *rfid.Device, t time.Duration, color rfid.LedMode) {
	if silent {
		return
	}
	if err := d.ChangeLed(color); err != nil {
		log.Print(err)
		return
	}
	if err := d.Beep(t); err != nil {
		log.Print(err)
		return
	}
	if err := d.ChangeLed(rfid.LedOff); err != nil {
		log.Print(err)
		return
	}
}

func readOnce(d *rfid.Device) (id []byte, err error) {
	for {
		id, err = d.ReadTag()
		if err != rfid.ErrNoTag {
			break
		}
	}
	return id, err
}

func writeOnce(d *rfid.Device, id []byte) (err error) {
	for {
		err = d.WriteTag(id)
		if err != rfid.ErrNoTag {
			return err
		}
	}
}

func infoMode(d *rfid.Device) {
	model, err := d.Info()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(model)
}

func readMode(d *rfid.Device) {
	id, err := readOnce(d)
	if err != nil {
		emitError(d)
		log.Fatal(err)
	}
	fprintID(os.Stdout, id)
	emitOK(d)
}

func readLoopMode(d *rfid.Device) {
	var lastID []byte
	for {
		id, err := readOnce(d)
		if err != nil {
			// suppress read errors in loop mode
			continue
		}
		if bytes.Equal(id, lastID) {
			continue
		}
		if lastID != nil {
			fmt.Println()
		}
		fprintID(os.Stdout, id)
		emitOK(d)
		lastID = id
	}
}

func writeMode(d *rfid.Device, payload []byte) {
	err := writeOnce(d, payload)
	if err != nil {
		emitError(d)
		log.Fatal(err)
	}
	fmt.Println("write successful")
	emitOK(d)
	time.Sleep(50 * time.Millisecond)
	emitOK(d)
}

func fprintID(w io.Writer, id []byte) {
	fmt.Fprintf(w, "hex      : %x\n", id)
	if len(id) == 5 {
		id40 := make([]byte, 8)
		id24 := make([]byte, 4)
		copy(id40[3:], id)
		copy(id24[1:], id[2:])
		fmt.Fprintf(w, "dec(40)  : %013d\n", binary.BigEndian.Uint64(id40))
		fmt.Fprintf(w, "dec(32)  : %010d\n", binary.BigEndian.Uint32(id[1:]))
		fmt.Fprintf(w, "dec(24)  : %08d\n", binary.BigEndian.Uint32(id24))
		fmt.Fprintf(w, "dec(8+16): %03d,%05d\n", id[2], binary.BigEndian.Uint16(id[3:]))
	}
}

func main() {
	var dev string
	var mode string
	var payload []byte

	flag.StringVar(&dev, "dev", "/dev/ttyUSB0", "RFID read/writer serial interface device")
	flag.StringVar(&mode, "mode", "read", "Application mode, one of: read, read-loop, write, info")
	flag.BoolVar(&silent, "silent", false, "Skip beeps and LED flashes, reduces number of commands sent to the reader")
	flag.Func("payload", "Hex payload to write", func(s string) error {
		var err error
		payload, err = hex.DecodeString(s)
		return err
	})
	flag.Parse()

	d, err := rfid.OpenDevice(dev, false)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	if !silent {
		if err := d.ChangeLed(rfid.LedOff); err != nil {
			log.Fatal("switching LED off", err)
		}
	}

	switch mode {
	case "info":
		infoMode(d)
	case "read":
		readMode(d)
	case "read-loop":
		readLoopMode(d)
	case "write":
		writeMode(d, payload)
	default:
		flag.PrintDefaults()
		os.Exit(2)
	}
}
