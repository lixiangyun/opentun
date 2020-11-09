package main

import (
	"flag"
	"fmt"
	"github.com/lixiangyun/opentun"
	"io"
	"log"
	"net"
	"time"
)


func OpenUdp(bindAddr string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", bindAddr)
	if err != nil {
		return nil, err
	}
	udpHander, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return udpHander, nil
}

func UdpWrite(conn *net.UDPConn, dstAddr *net.UDPAddr, body []byte ) error {
	cnt, err := conn.WriteToUDP(body, dstAddr)
	if err != nil {
		return fmt.Errorf("udp write fail, %s", err.Error())
	}
	if cnt != len(body) {
		return fmt.Errorf("udp send %d out of %d bytes", cnt, len(body))
	}
	return nil
}

var tunRead chan []byte
var tunWrite chan []byte

func TunRecv(tun opentun.TunApi)  {
	var buff [65535]byte
	for  {
		cnt, err := tun.Read(buff[:])
		if err != nil {
			log.Printf("tun read fail, %s\n", err.Error())
			continue
		}
		pkg := make([]byte, cnt)
		copy(pkg, buff[:cnt])
		tunRead <- pkg
	}
}

func TunSend(tun opentun.TunApi)  {
	for  {
		buff := <- tunWrite
		err := tun.Write(buff)
		if err != nil {
			log.Printf("tun write fail, %s\n", err.Error())
		}
	}
}

var remoteUdpAddr *net.UDPAddr

func UdpRecv(conn *net.UDPConn)  {
	var buff [65535]byte

	for  {
		cnt, addr, err := conn.ReadFromUDP(buff[:])
		if err != nil {
			if err != io.EOF {
				log.Println(err.Error())
			}
			return
		}

		if remote == "" {
			remoteUdpAddr = addr
		}

		buff2 := make([]byte, cnt)
		copy(buff2, buff[:cnt])

		tunWrite <- buff2
	}
}

func UdpSend(conn *net.UDPConn)  {
	for  {
		buff := <- tunRead
		_, err := conn.WriteToUDP(buff, remoteUdpAddr)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func UdpClientInit(addr string) error {
	conn, err := OpenUdp(addr)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	for i:=0;i<200;i++ {
		go UdpRecv(conn)
		go UdpSend(conn)
	}

	return nil
}

var (
	help   bool
	ipnet  string
	local  string
	remote string
	client bool
)

func init()  {
	flag.StringVar(&ipnet, "ipnet", "172.168.1.1/16", "ip + network")
	flag.StringVar(&remote, "remote", "", "bind or connect")
	flag.StringVar(&local, "local", "192.168.3.3:1020", "bind or connect")
	flag.BoolVar(&help, "help", false, "help usage")
}

func main()  {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	ip, ipnet, err := net.ParseCIDR(ipnet)
	if err != nil {
		log.Println(err.Error())
		return
	}

	tunWrite = make(chan []byte, 1024)
	tunRead  = make(chan []byte, 1024)

	var tun opentun.TunApi
	tun, err = opentun.OpenTun("eth0", ip, *ipnet)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if remote != "" {
		remoteUdpAddr, err = net.ResolveUDPAddr("udp", remote)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}

	for i:=0;i<10;i++ {
		go TunRecv(tun)
		go TunSend(tun)
	}

	UdpClientInit(local)

	log.Println("tun init success")

	for {
		time.Sleep(time.Hour)
	}
}