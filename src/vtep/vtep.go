package vtep

import (
	"fmt"
	"trace"
	"strings"
	"packet"
	"intf"
	"vlan"
	"host"
)

type UpStream interface {
	RcvUp(p *packet.Packet, from *Vtep, vLearn map[string]string, svLearn map[string]string)
}

type Pkt_Vlan struct {
	P *packet.Packet
	V *vlan.Vlan
}

func ( vt *Vtep) RcvUp( p *packet.Packet, v *vlan.Vlan ) {
	trace.T(fmt.Sprintf( "Received from %s on %s\n", v.Name, vt.Name ))
	p.History += fmt.Sprintf( "Received from %s on %s\n", v.Name, vt.Name )
	vt.RcvChan <- Pkt_Vlan{p, v}
}

type Vtep struct {
	Name string
	v *vlan.Vlan
	vLearn map[string]string
	serviceVlan *vlan.Vlan
	serviceVlanLearn map[string]string
	vlanVniMap map[string]string // Just 1 mapping function for both directions
	RcvChan chan Pkt_Vlan
	TxChan chan *packet.Packet
	UPStream UpStream
	vfpRules DFVfp
	pc packet.PktCounts
}

type DFVfp struct {
	Name string
	Vlan string
	New_Vlan string
	Mac string
	v []*vlan.Vlan
}

func (d DFVfp) String() string {
	s := fmt.Sprintf("Match on Vlan %v and Mac %v, Set Vlan to %s\n",
		d.Vlan, d.Mac, d.New_Vlan)
	return s
}

func ( r *DFVfp) AddVlan( v *vlan.Vlan ) {
	if r.v == nil {
		r.v = make([]*vlan.Vlan, 1)
		r.v[0] = v
		return
	}
	r.v = append(r.v, v )
}

func ( r *DFVfp) RcvUp( p *packet.Packet, i *intf.Intf ) {
	// Look at destination of packet and send it to right interface
	d := p.Vlan
	if ( (d == r.Vlan) && (p.SrcMac == r.Mac)) || ( r.Mac == "" ) {
		p.OldVlan = p.Vlan
		p.Vlan = r.New_Vlan
		p.History += fmt.Sprintf("VFP-Vlan changed from %s to %s\n", p.OldVlan, p.Vlan)
		for _,v := range r.v {
			if v.Name == r.New_Vlan {
				p.History += fmt.Sprintf("VFP-Fwd packet to to %s\n", v.Name)
				v.RcvUp( p, i)
			}
		}
	}
}

func (v Vtep) String() string {
	s:= fmt.Sprintf("Vtep Name   - %v\n%v\n%v%v\n%v",
		v.Name, v.v, v.vLearn, v.serviceVlan, v.serviceVlanLearn  )
	s += "\n"
	return s
}

func (vt *Vtep) GetHost(h string) (*host.Host, bool) {
	// Loop thru all vlans
	host, ok := vt.v.GetHost(h)
	if ok{
		return host, ok
	}
	host,ok = vt.serviceVlan.GetHost(h)
	if ok{
		return host, ok
	}

	return nil, false
}

func (vt *Vtep) learnMac( v *vlan.Vlan, p *packet.Packet ) {
	// See about learning Src
	var vLearn *map[string]string
	if v.Name == vt.serviceVlan.Name {
		vLearn = &vt.serviceVlanLearn
	} else {
		vLearn = &vt.vLearn
	}
	sender := p.SrcMac
	_, Learned := (*vLearn)[sender]
	if !Learned {
		p.History += fmt.Sprintf("Mac %s learnt on %s on %s\n", sender, v.Name, vt.Name)
		(*vLearn)[sender] = sender
	}
}

func (vt *Vtep) CreateService( v string, serviceVlan string ) {
	// ! Firewall and 2 interfaces is all we need
	vt.v = new(vlan.Vlan)
	vt.v.Name = v
	vt.v.UPStream = vt
	vt.vLearn = make( map[string]string)
	vt.vlanVniMap = make(map[string]string)
	vt.vlanVniMap[fmt.Sprintf("%s-Vni", v)] = v
	vt.vlanVniMap[v] = fmt.Sprintf("%s-Vni", v)
	vt.vlanVniMap[fmt.Sprintf("%s-Vni", serviceVlan)] = serviceVlan
	vt.vlanVniMap[serviceVlan] = fmt.Sprintf("%s-Vni", serviceVlan)

	vt.serviceVlan = new(vlan.Vlan)
	vt.serviceVlan.Name = serviceVlan
	vt.serviceVlan.UPStream = vt
	vt.serviceVlanLearn = make( map[string]string)

	intfNear := intf.Intf{Name:fmt.Sprintf("IntfNear"), EgressVlan:""}
	intfNear.UPStream = vt.v
	vt.v.AddIntf(&intfNear)

	intfFar := intf.Intf{Name:fmt.Sprintf("IntfFar"), EgressVlan:v}
	intfFar.UPStream = vt.serviceVlan
	vt.serviceVlan.AddIntf(&intfFar)
	var vfpRule *DFVfp
	vfpRule = &DFVfp{Name:"VFP Rule",
		Vlan:v,
		New_Vlan:serviceVlan,
		Mac:"" }
	vfpRule.AddVlan( vt.v )
	vfpRule.AddVlan( vt.serviceVlan )
	// Add vfp Rule to interface having this host
	intfFar.UPStream = vfpRule
	
	fw := host.FireWall{Name:"FireWall"}
	intfNear.AttachFirewall( &fw, true )
	intfFar.AttachFirewall( &fw, false )

	
	intfNear.Enable()
	intfFar.Enable()
	fw.Enable()
	vt.v.Enable()
	vt.serviceVlan.Enable()
	vt.Enable()
}

func (vt *Vtep) Create( v string, serviceVlan string, blueZoneH []string, redZoneH []string) {
	vt.v = new(vlan.Vlan)
	vt.v.Name = v
	vt.v.UPStream = vt
	vt.vLearn = make( map[string]string)
	vt.vlanVniMap = make(map[string]string)
	vt.vlanVniMap[fmt.Sprintf("%s-Vni", v)] = v
	vt.vlanVniMap[v] = fmt.Sprintf("%s-Vni", v)
	vt.vlanVniMap[fmt.Sprintf("%s-Vni", serviceVlan)] = serviceVlan
	vt.vlanVniMap[serviceVlan] = fmt.Sprintf("%s-Vni", serviceVlan)
	for _,h := range blueZoneH {
		host := host.Host{Name:h, Vlan:v}
		intf := intf.Intf{Name:fmt.Sprintf("Intf-%s", h), EgressVlan:""}
		intf.AttachHost(&host)
		intf.UPStream = vt.v
		vt.v.AddIntf(&intf)
		intf.Enable()
		host.Enable()
	}
	vt.serviceVlan = new(vlan.Vlan)
	vt.serviceVlan.Name = serviceVlan
	vt.serviceVlan.UPStream = vt
	vt.serviceVlanLearn = make( map[string]string)
	for _,h := range redZoneH {
		host := host.Host{Name:h, Vlan:v}
		intf := intf.Intf{Name:fmt.Sprintf("Intf-%s", h), EgressVlan:v} // Setup Egress Rule too
		intf.AttachHost(&host)
		vt.serviceVlan.AddIntf(&intf)
//		vt.v.AddIntf(&intf)
		intf.Enable()
		host.Enable()
		// For Red Zone create vfp Rules
		var vfpRule *DFVfp
		vfpRule = &DFVfp{Name:"VFP Rule",
			Vlan:v,
			New_Vlan:serviceVlan,
			Mac:fmt.Sprintf("%s-Mac",host.Name) }
		vfpRule.AddVlan( vt.v )
		vfpRule.AddVlan( vt.serviceVlan )
		// Add vfp Rule to interface having this host
		intf.UPStream = vfpRule
	}
	vt.v.Enable()
	vt.serviceVlan.Enable()
	vt.Enable()
}

func (vt *Vtep) Enable() {
	if (vt.RcvChan == nil) {
		vt.RcvChan = make(chan Pkt_Vlan, 100)
		vt.TxChan = make(chan *packet.Packet, 100)
	}
	go func() {
		for {
			select {
			case p := <- vt.RcvChan:
				vt.pc.RxIncr()
				// See about learning Src
				vt.learnMac(p.V, p.P)
				// See about forwarding
				mapVlan, _ := vt.vlanVniMap[p.P.Vlan]
				p.P.Vlan = mapVlan
				vt.UPStream.RcvUp(p.P, vt, vt.vLearn, vt.serviceVlanLearn)
			case p:= <- vt.TxChan:
				vt.pc.TxIncr()
				// See about forwarding
				mapVlan, _ := vt.vlanVniMap[p.Vlan]
				p.Vlan = mapVlan
				trace.T(fmt.Sprintf("Sending out from %s to %s\n", vt.Name, p.Vlan))
				p.History += fmt.Sprintf("Sending out from %s to %s\n", vt.Name, p.Vlan)
				if strings.Contains(p.Vlan, "Service") {
					vt.serviceVlan.TxChan <- p
				} else {
					vt.v.TxChan <- p
				}
			}
		}
	}()
}
