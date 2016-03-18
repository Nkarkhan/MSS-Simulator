package intf

import (
	"fmt"
	"reflect"
	"strings"
	"trace"
	"packet"
	"host"
)

type UpStream interface {
	RcvUp(p *packet.Packet, from *Intf)
}

type Intf struct {
	Name string   
	rcvChan chan *packet.Packet
	TxChan chan *packet.Packet
	UPStream UpStream
	EgressVlan string
	rxPkts int
	txPkts int
	dropPkts int
	h map[string]*host.Host
	fwAttached *host.FireWall
	fwAttachedNear bool
} 

type Pkt_Intf struct {
	P *packet.Packet
	I *Intf
}

func (i Intf) String() string {
	s := fmt.Sprintf("Interface Name - %v, Hosts attached %v\n", i.Name, i.h )
	if (i.rxPkts != 0 || i.txPkts !=0 || i.dropPkts!= 0) {
		s += fmt.Sprintf("Interface Name   - %v\n Receive Packets   - %v\n Transmit Packets - %v\n Drop Packets - %v\n\n",
			i.Name, i.rxPkts, i.txPkts, i.dropPkts)
	}
	if i.UPStream != nil {
		// This can be a vlan or a rule
		if strings.Contains(fmt.Sprintf("%v",reflect.TypeOf(i.UPStream)), "DFVfp") {
			s += "Vfp Rule atached\n"
			s += fmt.Sprintf("%v", i.UPStream)
		}
	}
	return s
}

func ( i *Intf) AttachHost( h *host.Host ) {
	h.UPStream = i
	if i.h == nil {
		i.h = make(map [string]*host.Host)
	}
	i.h[fmt.Sprintf("%s-Mac", h.Name )] = h
}

func ( i *Intf) AttachFirewall( fw *host.FireWall, near bool ) {
	i.fwAttached = fw
	if near {
		fw.UPStreamNear = i
		i.fwAttachedNear = true
	} else {
		fw.UPStreamFar = i
		i.fwAttachedNear = false
	}
}

func (i *Intf) GetHost(h string) (*host.Host, bool) {
	host, ok := i.h[ fmt.Sprintf("%s-Mac", h ) ]
	if ok {
	} else {
	}
	return host, ok
}

func ( i *Intf) RcvUp( p *packet.Packet, h *host.Host ) {
	if (h == nil) {
		trace.T(fmt.Sprintf("rcv from Firewall"))
		p.History += fmt.Sprintf("rcv from Firewall on %s\n", i.Name)
		i.rxPkts++
		i.UPStream.RcvUp( p, i )
		return
	}
	// Validate packet
	trace.T(fmt.Sprintf("rcv on %s from %s", i.Name, h.Name))
	_,ok := i.h[fmt.Sprintf("%s-Mac", h.Name )]
	if ok {
		trace.T(fmt.Sprintf( " Received from %s on %s\n", h.Name, i.Name ) )
		p.History += fmt.Sprintf( "Received from %s on %s\n", h.Name, i.Name )
		if i.UPStream != nil {
			i.rxPkts++
			i.UPStream.RcvUp( p, i )
		} else {
			// Consume it
			i.rcvChan <- p
		}
	} else {
		fmt.Printf( "**Error - Received %v from %s on %s\n", p, h.Name, i.Name )
		fmt.Printf( "**Error - Hosts are %v on %v\n", i.h, i.Name )
		i.dropPkts++
	}
}

func ( i *Intf) Enable() {
	if i.rcvChan == nil {
		i.rcvChan = make(chan *packet.Packet, 100 )
		i.TxChan = make(chan *packet.Packet, 100 )
	}
     // Wait till packet received and then print it
     go func() {
     	for {
     	     select {
	     case  <- i.rcvChan:
		     trace.T("")
		     fmt.Printf("**********");
		     i.rxPkts++
	     case p:= <- i.TxChan:
		     trace.T(fmt.Sprintf("Sending out on %s\n", i.Name))
		     if i.EgressVlan == "" {
		     } else {
			     p.Vlan = i.EgressVlan
			     p.History += fmt.Sprintf("VXlate - Change egress vlan to %s\n", i.EgressVlan)
		     }
		     if i.fwAttached != nil {
			     trace.T(fmt.Sprintf("On to Firewall\n"))
			     if i.fwAttachedNear {
				     trace.T("Near intf")
				     i.fwAttached.RcvChanNear <- p
			     } else {
				     trace.T("Far intf")
				     i.fwAttached.RcvChanFar <- p
			     }
		     } else {
			     h,ok := i.h[p.DstMac]
			     if ok {
				     trace.T(fmt.Sprintf("Sending packet out to %s from %s\n", h.Name, i.Name))
				     p.History += fmt.Sprintf("Sending packet out to %s from %s\n", h.Name, i.Name)
				     h.RcvChan <- p
				     i.txPkts++
			     } else {
				     i.dropPkts++
			     }
		     }
	     }	 
     	} 
     }()
}
