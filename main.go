package main

import (
	"flag"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

const (
	_  = iota
	kB = 1 << (10 * iota)
	mB
	gB
	tB
)

var finish = make(chan struct{})
var nmap = make(map[string]datausage)
var (
	cidr                        *net.IPNet
	svtf, infc, srtx, srfm, trm *string
	sprt, dprt                  *int
	cHz                         uint
)

func main() {
	log.SetFlags(0)
	ipxr := flag.String("net", "192.168.1.0/24", "network to capture on <ipv4/cidr>")
	infc = flag.String("inf", "lo", "network interface to capture on <interface_name>")
	dbfi := flag.String("svd", "ipfm.db", "database to save data <database_name>")
	ftim := flag.String("ttf", "3", "time in second to flush data into the database <integer>")
	svtf = flag.String("txt", "false", "also save data to file <filename|false>")
	trm = flag.String("hbm", "MB", "show (in txt file) data usage in <KB|MB|GB|TB>")
	srtx = flag.String("srt", "RX", "sort data in txt file based on <TX|RX>")
	srfm = flag.String("srf", "descending", "sort data in txt file based on <descending|ascending>")
	sprt = flag.Int("sprt", -1, "port to capture <0-65535>")
	dprt = flag.Int("dprt", -1, "port to capture <0-65535>")

	flag.Parse()
	if *dbfi == *svtf {
		log.Fatal("database name and filename can not be the same.")
	}
	if match, _ := regexp.MatchString("^(RX|TX)$", *srtx); !match {
		log.Fatal("regexp.MatchString: syntax err in srt flag ", *srtx)
	}
	if match, _ := regexp.MatchString("^(descending|ascending)$", *srfm); !match {
		log.Fatal("regexp.MatchString: syntax err in srf flag ", *srfm)
	}
	if match, _ := regexp.MatchString("^(KB|MB|GB|TB)$", *trm); !match {
		log.Fatal("regexp.MatchString: syntax err in hbm flag ", *trm)
	} else {
		switch *trm {
		case "KB":
			cHz = kB
		case "MB":
			cHz = mB
		case "GB":
			cHz = gB
		case "TB":
			cHz = tB
		}
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
	log.Printf("starting go-ipfm on %v interface, network %v, database %v, source port %v, dest port %v", *infc, *ipxr, *dbfi, *sprt, *dprt)
	flush := time.Tick(time.Duration(sectime) * time.Second)
	packetSource := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ip := ipLayer.(*layers.IPv4)
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				if (*sprt == -1 || tcp.SrcPort == layers.TCPPort(*sprt)) && (*dprt == -1 || tcp.DstPort == layers.TCPPort(*dprt)) {
					select {
					case <-flush:
						// saveTOdatabases(*dbfi)
						accFrom(ip, tcp)
					default:
						accFrom(ip, tcp)
					}
				}
			}
		}
	}
}

func accFrom(ip *layers.IPv4, tcp *layers.TCP) {
	if issrc := cidr.Contains(net.ParseIP(ip.SrcIP.String())); issrc {
		// RX OR TX ?
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
