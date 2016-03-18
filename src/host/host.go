package host

import (
	"fmt"
	"time"
	"trace"
	"packet"
)

type UpStream interface {
	RcvUp(p *packet.Packet, from *Host)
}

type PktCounts struct {
	rxPkts int
	txPkts int
	dropPkts int
}

func (pc PktCounts) String() string {
	s:=""
	if (pc.rxPkts != 0 || pc.txPkts !=0 || pc.dropPkts!= 0) {
		s += fmt.Sprintf("Receive Packets   - %v\n Transmit Packets - %v\n Drop Packets - %v\n\n",
			pc.rxPkts, pc.txPkts, pc.dropPkts)
	}
	return s
}

func (pc PktCounts) PktCounter() bool {
	return (pc.rxPkts != 0 || pc.txPkts !=0 || pc.dropPkts!= 0)
}

type Host struct {
	Name string
	UPStream UpStream
	RcvChan chan *packet.Packet
	Vlan string
	pc PktCounts
}

func (h Host) String() string {
	s := ""
	if h.pc.PktCounter() {
		s += fmt.Sprintf("Host Name   - %v\n%v\n", h.Name, h.pc)
	}
	return s
}

func (h *Host) Send( PktType string, DstHost string, D time.Duration, N int ) {
	go func() {
		for i:=0; i < N; i++ {
			p := packet.Packet{ Name:fmt.Sprintf("Packet from %s", h.Name ),
				Vlan:h.Vlan,
				DstMac:fmt.Sprintf("%s-Mac", DstHost ),
				SrcMac:fmt.Sprintf("%s-Mac", h.Name ),
				DstIp:fmt.Sprintf("%s-IP", DstHost ),
				SrcIp:fmt.Sprintf("%s-IP", h.Name ),
				Type:PktType,
				Loop: 1 }
			p.History += fmt.Sprintf("Sent from %s\n", h.Name)
			h.UPStream.RcvUp( &p, h )
			h.pc.txPkts++
			time.Sleep(D)
		}
	}()
}

func (h *Host) Enable() {
	if h.RcvChan == nil {
		h.RcvChan = make( chan *packet.Packet, 100)
	}
	go func() {
		for {
			p := <-h.RcvChan
			trace.T( fmt.Sprintf("Host rcv %s", h.Name) )
			if p.DstMac != fmt.Sprintf("%s-Mac", h.Name) {
				h.pc.dropPkts++
			} else {
				trace.T( "My mac ")
				h.pc.rxPkts++
				if p.Loop != 0 {
					// Loop it back out
					newP := &packet.Packet{}
					*newP = *p
					newP.Loop = 0
					newP.Vlan = h.Vlan
					tmpStr  := newP.DstMac
					newP.DstMac = newP.SrcMac
					newP.SrcMac = tmpStr
					tmpStr  = newP.DstIp
					newP.DstIp = newP.SrcIp
					newP.SrcIp = tmpStr
					newP.History += fmt.Sprintf("Packet looped from %s to %s\n", h.Name, newP.DstMac );
					h.UPStream.RcvUp( newP, h )
					h.pc.txPkts++
				} else {
					p.History += fmt.Sprintf("EOF on %s", h.Name);
					fmt.Printf("Host %s rcv %v\n", h.Name, p)
				}
			}

		}
	}()
}

type FireWall struct {
	Name string
	UPStreamNear UpStream
	UPStreamFar UpStream
	RcvChanNear chan *packet.Packet
	RcvChanFar chan *packet.Packet
	nearPc PktCounts
	farPc PktCounts
}

func (fw FireWall) String() string {
	s := ""
	if fw.nearPc.PktCounter()  || fw.farPc.PktCounter() {
		s += fmt.Sprintf("FireWall Name   - %v\nNear Intf \n%v\nFar intf \n%v\n", fw.Name, fw.nearPc, fw.farPc)
	}
	return s
}

func (fw *FireWall) Enable() {
	if fw.RcvChanNear == nil {
		fw.RcvChanNear = make( chan *packet.Packet, 100)
		fw.RcvChanFar = make( chan *packet.Packet, 100)
	}
	go func() {
		for {
			select {
			case p := <-fw.RcvChanNear:
				fw.nearPc.rxPkts++
				fw.farPc.txPkts++
				trace.T(fmt.Sprintf("FireWall Near rcv\n"))
				p.History += fmt.Sprintf("FireWall Near rcv\n")
				fw.UPStreamFar.RcvUp( p, nil )
			case p := <-fw.RcvChanFar:
				fw.farPc.rxPkts++
				fw.nearPc.txPkts++
				trace.T(fmt.Sprintf("FireWall Far rcv\n"))
				p.History += fmt.Sprintf("FireWall Far rcv\n")
				fw.UPStreamNear.RcvUp( p, nil )
			}
		}
	}()
}
