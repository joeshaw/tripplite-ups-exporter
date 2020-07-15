# tripplite-ups-exporter

This is a simple Prometheus exporter for Tripp Lite UPS devices that expose their properties as a USB HID device.

As of today this has only been tested on a SMART1500LCD with USB vendor ID 09AE and product ID 2012.  It has only been tested on Linux (and specifically a Raspberry Pi) though it should also work on Mac and Windows thanks to the [Gopher Interface Devices](https://github.com/karalabe/hid) HID library it uses.

## Installing

The tool can be installed with:

    go get -u github.com/joeshaw/tripplite-ups-exporter

Because this program needs to access USB devices you may need elevated permissions to access the devices.  One way to do this is to run as root, but on Linux you can also set up a `udev` rule to give access to a unix group.  On the Raspberry Pi the default `pi` user is in the `dialout` group, which is meant to give access to serial devices.  So I just use that and create a file in `/etc/udev/rules.d`:

    echo 'SUBSYSTEM=="usb", ATTRS{idVendor}=="09ae", ATTRS{idProduct}=="2012", GROUP="dialout"' > /etc/udev/rules.d/55-tripplite-ups.rules

And then run:

    udevadm control --reload-rules

to load the new rule in.

Then you can run the service:

    tripplite-ups-exporter

By default the exporter will listen on port 9528.  This can be changed with the `-addr` flag.

## Contributing

Issues and pull requests are welcome.  When filing a PR, please make sure the code has been run through `gofmt`.

## License

Copyright 2020 Joe Shaw

`tripplite-ups-exporter` is licensed under the MIT License.  See the LICENSE file for details.


