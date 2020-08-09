package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

var finish = make(chan struct{})
var nmap = make(map[string]datausage)
var (
	cidr *net.IPNet
	svtf *string
)

func main() {
	log.SetFlags(0)
	ipxr := flag.String("net", "192.168.1.0/24", "network to capture on <ipv4/cidr>")
	infc := flag.String("inf", "lo", "network interface to capture on <interface_name>")
	dbfi := flag.String("svd", "ipfm.db", "database to save data <database_name>")
	ftim := flag.String("ttf", "3", "time in second to flush data into the database <integer>")
	svtf = flag.String("txt", "false", "also save data to file <filename|false>")
	flag.Parse()
	if *dbfi == *svtf {
		log.Fatal("database name and filename can not be the same.")
	}
	sectime, err := strconv.Atoi(*ftim)
	if err != nil {
		log.Fatal(err)
	}
	_, cidr, err = net.ParseCIDR(*ipxr)
	if err != nil {
		log.Fatal(err)
	}
	handle, err := pcapgo.NewEthernetHandle(*infc)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	log.Printf("starting go-ipfm on %v interface, network %v, database %v", *infc, *ipxr, *dbfi)
	flush := time.Tick(time.Duration(sectime) * time.Second)
	packetSource := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ip := ipLayer.(*layers.IPv4)
			select {
			case <-flush:
				saveTOdatabases(*dbfi)
				accFrom(ip)
			default:
				accFrom(ip)
			}
		}
	}
}

func accFrom(ip *layers.IPv4) {
	if issrc := cidr.Contains(net.ParseIP(ip.SrcIP.String())); issrc {
		if val, ok := nmap[ip.SrcIP.String()]; !ok {
			nmap[ip.SrcIP.String()] = datausage{ip: ip.SrcIP.String(), tx: uint(ip.Length)}
		} else {
			nmap[ip.SrcIP.String()] = datausage{ip: ip.SrcIP.String(), tx: uint(ip.Length) + val.tx, rx: val.rx}

		}
	} else if isdst := cidr.Contains(net.ParseIP(ip.DstIP.String())); isdst {
		if val, ok := nmap[ip.DstIP.String()]; !ok {
			nmap[ip.DstIP.String()] = datausage{ip: ip.DstIP.String(), rx: uint(ip.Length)}
		} else {
			nmap[ip.DstIP.String()] = datausage{ip: ip.DstIP.String(), rx: uint(ip.Length) + val.rx, tx: val.tx}
		}
	}
}
