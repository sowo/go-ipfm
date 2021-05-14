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
	"github.com/google/gopacket/pcap"
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
var captureMask = "2006-01-02"
var (
	cidr                        *net.IPNet
	svtf, infc, srtx, srfm, trm *string
	dir, filt                   *string
	cHz                         uint
	lips                        []net.IP
)

func main() {
	log.SetFlags(0)
	infc = flag.String("inf", "lo", "network interface to capture on <interface_name>")
	dir = flag.String("dir", ".", "directory to save DB and file")
	dbfi := flag.String("svd", "ipfm.db", "database to save data <database_name>")
	svtf = flag.String("txt", "false", "also save data to file <filename|false>")
	ftim := flag.String("ttf", "60", "time in second to flush data into the database <integer>")
	trm = flag.String("hbm", "MB", "show (in txt file) data usage in <KB|MB|GB|TB>")
	srtx = flag.String("srt", "RX", "sort data in txt file based on <TX|RX>")
	srfm = flag.String("srf", "descending", "sort data in txt file based on <descending|ascending>")
	filt = flag.String("flt", "", "BPF filter string, e.g. \"tcp and port 80\"")

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
	iface, err := net.InterfaceByName(*infc)
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range addrs {
		switch v := item.(type) {
		case *net.IPNet:
			lips = append(lips, v.IP)
		case *net.IPAddr:
			lips = append(lips, v.IP)
		}
	}
	handle, err := pcap.OpenLive(*infc, 128, true, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	if *filt != "" {
		err = handle.SetBPFFilter(*filt)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer handle.Close()
	log.Printf("starting go-ipfm on %v interface, database %v, filter \"%v\"", *infc, *dbfi, *filt)
	flush := time.Tick(time.Duration(sectime) * time.Second)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
			ip := ipLayer.(*layers.IPv4)
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				select {
				case <-flush:
					saveTOdatabases(*dir + "/" + *dbfi)
					accFrom(ip, tcp)
				default:
					accFrom(ip, tcp)
				}
			}
		}
	}
}

func FindIP(addrs []net.IP, addr net.IP) (int, bool) {
	for i, item := range addrs {
		if item.String() == addr.String() {
			return i, true
		}
	}
	return -1, false
}

func accFrom(ip *layers.IPv4, tcp *layers.TCP) {
	if _, issrc := FindIP(lips, ip.SrcIP); !issrc {
		// RX OR TX ?
		if val, ok := nmap[ip.SrcIP.String()]; !ok {
			nmap[ip.SrcIP.String()] = datausage{ip: ip.SrcIP.String(), rx: uint(ip.Length)}
		} else {
			nmap[ip.SrcIP.String()] = datausage{ip: ip.SrcIP.String(), rx: uint(ip.Length) + val.rx, tx: val.tx}

		}
	} else if _, isdst := FindIP(lips, ip.DstIP); !isdst {
		if val, ok := nmap[ip.DstIP.String()]; !ok {
			nmap[ip.DstIP.String()] = datausage{ip: ip.DstIP.String(), tx: uint(ip.Length)}
		} else {
			nmap[ip.DstIP.String()] = datausage{ip: ip.DstIP.String(), tx: uint(ip.Length) + val.tx, rx: val.rx}
		}
	}
}
