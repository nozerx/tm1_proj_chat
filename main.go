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
	fmt.Println(len(p2pHost.Host.Network().Peers()))
	// fmt.Println(p2pHost.Host.Network().Peers())
	chagrp, err := src.JoinGroup(p2pHost, "lobby")
	if err != nil {
		fmt.Fprintln(src.File, "Error while creating a group")
		panic(err)
	}
	time.Sleep(5 * time.Second)
	fmt.Print(chagrp)
	ui := src.NewUI(chagrp)
	ui.Run()

	defer src.File.Close()

}
