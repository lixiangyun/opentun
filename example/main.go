package main

import (
	"fmt"
	"github.com/lixiangyun/opentun"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
)

func main()  {
	ip, ipnet, err := net.ParseCIDR("172.168.1.1/16")
	if err != nil {
		log.Println(err.Error())
		return
	}

	var tun opentun.TunApi
	tun, err = opentun.OpenTun("eth0", ip, *ipnet)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("tun init success")

	buff := make([]byte, 1500)
	for  {
		cnt, err := tun.Read(buff)
		if err != nil {
			log.Printf("tun read fail, %s\n", err.Error())
			continue
		}

		eth := gopacket.NewPacket(buff[:cnt], layers.LayerTypeEtherIP, gopacket.Lazy)
		fmt.Printf("eth: %s\n", eth.String())

		if layer := eth.Layer(layers.LayerTypeEtherIP); layer != nil {
			fmt.Printf("This is a IP packet! %v\n", layer.LayerType())
			continue
		}
	}
}