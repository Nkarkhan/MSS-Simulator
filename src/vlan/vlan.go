package vlan

import (
       "fmt"
	"packet"
	"host"
       "intf"
       "trace"
)

type UpStream interface {
	RcvUp(p *packet.Packet, from *Vlan)
}

type vlan_intf struct {
     i *intf.Intf
     h map[string]string // Whats learnt on this intf
}

func (iv vlan_intf) String() string {
	if len(iv.h) > 0 {
		s := fmt.Sprintf(" Learnt on intf %s\n", iv.i.Name)
		for _,v := range iv.h {
			s += fmt.Sprintf(" Mac %s\n", v)
		}
		return s
	}
	return ""
}

type Vlan struct {
	Name string
	RcvChan chan intf.Pkt_Intf
	TxChan chan *packet.Packet
	UPStream UpStream
	intfs map[string]vlan_intf
	rxPkts int
	txPkts int
	dropPkts int
	floodPkts int
} 

func (v Vlan) String() string {
	s := fmt.Sprintf("Vlan Name %v\n", v.Name)
	for _,v := range v.intfs {
		s += fmt.Sprintf("%v\n", v.i)
	}
	if (v.rxPkts != 0 || v.txPkts !=0 || v.dropPkts!= 0 || v.floodPkts!= 0) {
		s += fmt.Sprintf("Vlan Name   - %v\n Receive Packets   - %v\n Transmit Packets - %v\n Drop Packets - %v\n Flood Packets - %v\n\n",
			v.Name, v.rxPkts, v.txPkts, v.dropPkts, v.floodPkts)
		return s
	}
	return s
}

func ( v *Vlan) AddIntf(i *intf.Intf)  {
	if ( v.intfs == nil ) {
		v.intfs = make(map[string]vlan_intf)
	}
	v.intfs[i.Name] = vlan_intf{i, make( map[string]string ) }
}

func (v *Vlan) GetHost(h string) (*host.Host, bool) {
	// Loop thru all vlan interfaces and find host
	for _,value := range v.intfs {
		host, ok := value.i.GetHost(h)
		if ok{
			return host, ok
		}
	}
	return nil, false
}

func (v *Vlan) DiscardLearnMac( mac string ) {
	for _,vi := range v.intfs {
		delete( vi.h, mac )
	}
}

func (v *Vlan) learnMac( i *intf.Intf, p *packet.Packet ) {
	// See about learning Src
	sender := p.SrcMac
	iv := v.intfs[i.Name]
	
	_, Learned := iv.h[sender]
	
	if !Learned {
		for _,vi := range v.intfs {
			_, move := vi.h[sender]
			if move {
				trace.T(fmt.Sprintf("Mac Move from %s to %s\n", vi.i.Name, iv.i.Name))
				p.History += fmt.Sprintf("Mac Move from %s to %s\n", vi.i.Name, iv.i.Name)
				delete(vi.h, sender)
			}
		}
		trace.T(fmt.Sprintf("Mac %s learnt on %s %s\n", sender, iv.i.Name, v.Name))
		p.History += fmt.Sprintf("Mac %s learnt on %s %s\n", sender, iv.i.Name, v.Name)
		iv.h[sender] = sender
	}
}

func (v *Vlan) sndDst( i *intf.Intf, p *packet.Packet ) {
	// See if we know where to send it
	destination := p.DstMac
	if i != nil {
		for k,iv := range v.intfs {
			_,know := iv.h[destination]
			if know {
				if k == i.Name {
					//Destination is on same intf
					return
				}
				trace.T(fmt.Sprintf("Packet forwarded to learnt interface %s\n", iv.i.Name))
				p.History += fmt.Sprintf("Packet forwarded to learnt interface %s\n", iv.i.Name)
				iv.i.TxChan <- p
				v.txPkts++
				return
			} else {
				fmt.Printf("Learnt mac lookup failed %s %v on %s\n", destination, iv.h, iv.i.Name)
			}
		}
	}
	// Time to flood
	// We are still here so flood it
	v.floodPkts++
	for k,iv := range v.intfs {
		if (i != nil) && (k == i.Name) {
			// Dont reflect packet back!
		} else {
			// Create a new packet
			newP := &packet.Packet{}
			*newP = *p
			trace.T(fmt.Sprintf("Packet forwarded to flood interface %s\n", iv.i.Name))
			newP.History += fmt.Sprintf("Packet forwarded to flood interface %s\n", iv.i.Name)
			iv.i.TxChan <- newP
		}
	}
}

func ( v *Vlan) RcvUp( p *packet.Packet, i *intf.Intf ) {
p.History += fmt.Sprintf( "Received from %s on %s\n", i.Name, v.Name )
	v.RcvChan <- intf.Pkt_Intf{p, i}
}

func (v *Vlan) Enable() {
	if (v.RcvChan == nil) {
		v.RcvChan = make(chan intf.Pkt_Intf, 100)
		v.TxChan = make(chan *packet.Packet, 100)
	}
	go func() {
		for {
			select {
			case p := <- v.RcvChan:
				v.rxPkts++
				// See about learning Src
				v.learnMac(p.I, p.P)
				// See about forwarding
				v.sndDst( p.I, p.P)
				if v.UPStream != nil {
					v.UPStream.RcvUp( p.P, v )
				} 	// Validate packet
			case p:= <- v.TxChan:
				// See about forwarding
				trace.T(fmt.Sprintf("Sending out on %s\n", v.Name))
				p.History += fmt.Sprintf("Sending out on %s\n", v.Name)
				v.sndDst( nil, p)
			}
		}
	}()
}
