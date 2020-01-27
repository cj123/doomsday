package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/cj123/go-ipsw/api"
)

var (
	ipswClient = api.NewIPSWClient("https://api.ipsw.me/v4", nil)
)

func main() {
	devices, err := ipswClient.Devices(false)

	if err != nil {
		log.Fatalf("Unable to retrieve firmware information, err: %s", err)
	}

	var firmwares []api.Firmware

	for _, device := range devices {
		deviceInformation, err := ipswClient.DeviceInformation(device.Identifier)

		if err != nil {
			log.Printf("Could not get firmwares for device: %s, err: %s", device.Identifier, err)
		}

		for _, firmware := range deviceInformation.Firmwares {
			firmwares = append(firmwares, firmware)
		}
	}

	sort.Slice(firmwares, func(i, j int) bool {
		if firmwares[i].Identifier == firmwares[j].Identifier {
			return firmwares[i].BuildID < firmwares[j].BuildID
		} else {
			return firmwares[i].Identifier < firmwares[j].Identifier
		}
	})

	for i, firmware := range firmwares {
		fmt.Printf("\r%04d/%04d checking... (%04d failed so far)", i+1, len(firmwares), failed)
		checkFirmware(firmware)
	}

	fmt.Println()

	log.Printf("%d firmwares total.", len(firmwares))
	log.Printf("%d firmwares failed.", failed)
	log.Printf("%.3f%% dead.", float64(failed)/float64(len(firmwares))*100.0)

	for version, count := range failedByVersion {
		log.Printf("iOS %s: %d failed", version, count)
	}

	for device, count := range failedByDevice {
		log.Printf("Device %s: %d failed", device, count)
	}
}

var (
	failed          int
	failedByVersion = make(map[string]int)
	failedByDevice  = make(map[string]int)
)

func checkFirmware(fw api.Firmware) {
	resp, err := http.Get(fw.URL)

	if err != nil {
		log.Printf("http err: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		failed++

		if _, ok := failedByDevice[fw.Identifier]; !ok {
			failedByDevice[fw.Identifier] = 1
		} else {
			failedByDevice[fw.Identifier]++
		}

		if _, ok := failedByVersion[fw.Version]; !ok {
			failedByVersion[fw.Version] = 1
		} else {
			failedByVersion[fw.Version]++
		}

		return
	}

	buf := make([]byte, 1024)

	n, _ := io.ReadFull(resp.Body, buf)

	if bytes.Contains(buf[:n], []byte("AccessDenied")) {
		failed++

		if _, ok := failedByDevice[fw.Identifier]; !ok {
			failedByDevice[fw.Identifier] = 1
		} else {
			failedByDevice[fw.Identifier]++
		}

		if _, ok := failedByVersion[fw.Version]; !ok {
			failedByVersion[fw.Version] = 1
		} else {
			failedByVersion[fw.Version]++
		}
	}
}
