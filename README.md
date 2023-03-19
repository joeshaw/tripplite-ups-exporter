# tripplite-ups-exporter

This is a simple Prometheus exporter for Tripp Lite UPS devices that expose their properties as a USB HID device.

As of today this has only been tested on Tripp Lite devices with USB vendor ID 09AE, including a SMART1500LCD with product ID 2012, an AVR900U and an AVR650UM, both with product ID 3024. It has only been tested on Linux (a Raspberry Pi, an Intel NUC (native build), a Linksys WRT3200ACM (native build), and a VoCore2 (cross-compiled)) though it should also work on Mac and Windows thanks to the [Gopher Interface Devices](https://github.com/karalabe/hid) HID library it uses.

## Cross Compiling

As this is written in Go, cross-compilation is logically simple, and the preferred method for building when the target OS is OpenWrt Linux. However, the C package used to access USB devices has to be compiled for the target platform, otherwise the UPS will never be detected. For OpenWrt, set up a toolchain as described in https://openwrt.org/docs/guide-developer/toolchain/use-buildsystem up to making the kernel_menuconfig. This will generate the staging_dir with the needed gcc compiler.

The OpenWrt `git clone` directory as used below is `$HOME/Source/External/openwrt`, so update the export line as needed to point the directory used and the toolchain version. The below:

    export STAGING_DIR=$HOME/Source/External/openwrt/staging_dir/toolchain-mipsel_24kc_gcc-11.2.0_musl
    CGO_ENABLED=1 CC=$STAGING_DIR/bin/mipsel-openwrt-linux-musl-gcc GOOS=linux GOARCH=mipsle GOMIPS=softfloat go get -a -ldflags '-w' github.com/karalabe/hid
    CGO_ENABLED=1 CC=$STAGING_DIR/bin/mipsel-openwrt-linux-musl-gcc GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -a -ldflags '-w -extldflags -static' -o tripplite-ups-exporter-mipsel main.go

will generate a static executable suitable for running on a VoCore2 running OpenWRT 22.03. The below generates a static executable suitable for running on a Linksys WRT3200ACM, also running running OpenWRT.

    export STAGING_DIR=$HOME/Source/External/openwrt/staging_dir/toolchain-arm_cortex-a9+vfpv3-d16_gcc-11.2.0_musl_eabi
    CGO_ENABLED=1 CC=$STAGING_DIR/bin/arm-openwrt-linux-muslgnueabi-gcc GOOS=linux GOARCH=arm go get -a -ldflags '-w' github.com/karalabe/hid
    CGO_ENABLED=1 CC=$STAGING_DIR/bin/arm-openwrt-linux-muslgnueabi-gcc GOOS=linux GOARCH=arm go build -a -ldflags '-w -extldflags -static' -o tripplite-ups-exporter-arm7 main.go

See https://go.dev/doc/install/source#environment for the list of valid combinations and additional CPU specific options (like GOMIPS as above).

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

### OpenWRT init script

The prometheus-tripplite-ups-exporter script goes in /etc/init.d on OpenWRT to enable automatic startup on boot. The script expects the appropriate binary at /usr/bin/prometheus-tripplite-ups-exporter.

## Contributing

Issues and pull requests are welcome.  When filing a PR, please make sure the code has been run through `gofmt`.

## License

Copyright 2020 Joe Shaw

`tripplite-ups-exporter` is licensed under the MIT License.  See the LICENSE file for details.


