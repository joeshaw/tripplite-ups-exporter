package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/karalabe/hid"
)

const (
	vendorID  = 0x09ae
	productID = 0x2012
)

type formatFunc func(data []byte, d *hid.Device, name string, w io.Writer)

type feature struct {
	name     string
	reportID byte
	length   int
	format   formatFunc
}

func intFormatter(data []byte, d *hid.Device, name string, w io.Writer) {
	var v uint64
	for i, b := range data {
		v |= uint64(b) << (8 * i)
	}

	fmt.Fprintf(w, "# TYPE %s gauge\n", name)
	fmt.Fprintf(w, "%s{path=\"%s\",product=\"%s\"} %d\n", name, d.Path, d.Product, v)
}

func oneTenthFloatFormatter(data []byte, d *hid.Device, name string, w io.Writer) {
	var v uint64
	for i, b := range data {
		v |= uint64(b) << (8 * i)
	}

	fmt.Fprintf(w, "# TYPE %s gauge\n", name)
	fmt.Fprintf(w, "%s{path=\"%s\",product=\"%s\"} %f\n", name, d.Path, d.Product, float64(v)/10.0)

}

func statusFormatter(data []byte, d *hid.Device, name string, w io.Writer) {
	bitOn := func(b byte, i int) int {
		if b&(1<<i) != 0 {
			return 1
		} else {
			return 0
		}
	}

	fields := []string{
		"shutdown_imminent",
		"ac_present",
		"charging",
		"discharging",
		"needs_replacement",
		"below_remaining_capacity",
		"fully_charged",
		"fully_discharged",
	}

	for i, f := range fields {
		n := name + "_" + f

		fmt.Fprintf(w, "# TYPE %s gauge\n", n)
		fmt.Fprintf(w, "%s{path=\"%s\",product=\"%s\"} %d\n", n, d.Path, d.Product, bitOn(data[0], i))
	}
}

var features = []feature{
	{"tripplite_config_voltage", 48, 1, intFormatter},
	{"tripplite_config_frequency_hz", 2, 1, intFormatter},
	{"tripplite_config_power_watts", 3, 2, intFormatter},
	{"tripplite_input_voltage", 24, 2, oneTenthFloatFormatter},
	{"tripplite_input_frequency_hz", 25, 2, oneTenthFloatFormatter},
	{"tripplite_output_voltage", 27, 2, oneTenthFloatFormatter},
	{"tripplite_output_power_watts", 71, 2, intFormatter},
	{"tripplite_current_charge_pct", 52, 1, intFormatter},
	{"tripplite_run_time_to_empty_minutes", 53, 2, intFormatter},
	{"tripplite_status", 50, 1, statusFormatter},
}

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":9528", "Prometheus exporter listen address")
	flag.Parse()

	var dinfos []hid.DeviceInfo
	for len(dinfos) == 0 {
		dinfos = hid.Enumerate(vendorID, productID)
		if len(dinfos) == 0 {
			fmt.Println("No devices found, waiting 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}

	var devices []*hid.Device
	for _, di := range dinfos {
		d, err := di.Open()
		if err != nil {
			fmt.Println(err)
			continue
		}

		devices = append(devices, d)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		for _, d := range devices {
			for _, f := range features {
				data := make([]byte, f.length+1)
				data[0] = f.reportID

				_, err := d.GetFeatureReport(data)
				if err != nil {
					fmt.Println(err)
					continue
				}

				f.format(data[1:], d, f.name, w)
			}
		}
	})

	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Println("Starting Prometheus exporter on", addr)
	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("Unable to run HTTP server: %v", err)
		os.Exit(1)
	}
}
