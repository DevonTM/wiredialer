package wiredialer

import (
	"io"
	"log"
	"os"

	"net"

	"context"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"

	"github.com/botanica-consulting/wiredialer/internal/config"
)

type WireDialer struct {
	tun    tun.Device
	tnet   *netstack.Net
	device *device.Device
}

func (d *WireDialer) Dial(network, address string) (net.Conn, error) {
	return d.tnet.Dial(network, address)
}

func (d *WireDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.tnet.DialContext(ctx, network, address)
}

func (d *WireDialer) DialPing(lAddress, rAddress *netstack.PingAddr) (*netstack.PingConn, error) {
	return d.tnet.DialPing(lAddress, rAddress)
}

func (d *WireDialer) ListenPing(address *netstack.PingAddr) (*netstack.PingConn, error) {
	return d.tnet.ListenPing(address)
}

func (d *WireDialer) ListenTCP(address *net.TCPAddr) (*gonet.TCPListener, error) {
	return d.tnet.ListenTCP(address)
}

func (d *WireDialer) ListenUDP(address *net.UDPAddr) (*gonet.UDPConn, error) {
	return d.tnet.ListenUDP(address)
}

func (d *WireDialer) LookupHost(host string) ([]string, error) {
	return d.tnet.LookupHost(host)
}

func (d *WireDialer) LookupContextHost(ctx context.Context, host string) ([]string, error) {
	return d.tnet.LookupContextHost(ctx, host)
}

func NewDialerFromFile(path string) (*WireDialer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return NewDialerFromConfiguration(file)
}

func NewDialerFromConfiguration(config_reader io.Reader) (*WireDialer, error) {
	iface_addresses, dns_addresses, mtu, ipcConfig, err := config.ParseConfig(config_reader)
	if err != nil {
		return nil, err
	}

	tun, tnet, err := netstack.CreateNetTUN(
		iface_addresses,
		dns_addresses,
		mtu)
	if err != nil {
		log.Panic(err)
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, ""))
	err = dev.IpcSet(ipcConfig)
	if err != nil {
		log.Panic(err)
	}
	err = dev.Up()
	if err != nil {
		log.Panic(err)
	}

	return &WireDialer{
		tun:    tun,
		tnet:   tnet,
		device: dev,
	}, nil
}
