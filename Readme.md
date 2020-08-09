# Go IPFM
Go-ipfm (IP Flow Meter) is an standalone bandwidth analysis tool written in [Golang](https://golang.org). (no need to libpcap)  
Core features:  
Measure download and upload usage per ip  
Save data to SQL database and text file  
No dependence on libpcap  
Runing on Windows, Linux and Unix based systems  

### Installation
Prerequisites: [Golang](https://golang.org) + [Git](https://git-scm.com)
Installing for windows - linux - freebsd and macos. (linux is recommended)  
Clone the code form Github or [Snix](https://slc.snix.ir) servers.
```
# git clone https://slc.snix.ir/snix/go-ipfm.git          # Snix
# git clone https://github.com/Sina-Ghaderi/go-ipfm.git   # Github  
# cd go-ipfm
# go get -v
# go build
# ./go-ipfm -inf eth0 -net 10.10.10.0/24 -svd database.db -ttf 5 -txt filename.txt
starting go-ipfm on eth0 interface, network 10.10.10.0/24, database database.db
...
```
### Usage and Options
```
#./go-ipfm -h
Usage of ./go-ipfm:
  -inf string
    	network interface to capture on <interface_name> (default "lo")
  -net string
    	network to capture on <ipv4/cidr> (default "192.168.1.0/24")
  -svd string
    	database to save data <database_name> (default "ipfm.db")
  -ttf string
    	time in second to flush data into the database <integer> (default "3")
  -txt string
    	also save data to file <filename|false> (default "false")
```
### Examples
So im runing go-ipfm on KVM host to measure vm's bandwidth usage  
```
# ./go-ipfm -inf virbr0 -net 192.168.122.0/24 -txt data.txt
starting go-ipfm on virbr0 interface, network 192.168.122.0/24, database ipfm.db
...
```
Open data.txt file by using `watch` and `cat` command in linux:
```
# watch cat data.txt

SrcIP Address		RX		TX
192.168.122.110		16624		18304
192.168.122.1		0		4176
192.168.122.104		3612311		113727

```

### Support and Social Media
So if you interested to learn [Golang](https://golang.org) follow my [Instagram Account](https://instagram.com/Gonoobies), Thanks. 
