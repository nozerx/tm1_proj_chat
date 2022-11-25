package main

import (
	"documents/GitHub/tm1_proj_chat/src"
	"fmt"
	"time"
)

func main() {
	fmt.Println("The app is starting")
	fmt.Println("This may take some time")
	p2pHost := src.EstablishP2P()
	fmt.Println(len(p2pHost.Host.Network().Peers()))
	fmt.Println(p2pHost.Host.Network().Peers())
	p2pHost.AdvertiseConnect()
	fmt.Fprintln(src.File, len(p2pHost.Host.Network().Peers()), "connections to service nodes established")
	// fmt.Println(p2pHost.Host.Network().Peers())
	chagrp, err := src.JoinGroup(p2pHost, "lobby")
	if err != nil {
		fmt.Fprintln(src.File, "Error while creating a group")
		panic(err)
	}
	time.Sleep(5 * time.Second)
	src.PrintAllActiveConnections(p2pHost)
	fmt.Print(chagrp)
	ui := src.NewUI(chagrp)
	ui.Run()

	defer src.File.Close()

}
