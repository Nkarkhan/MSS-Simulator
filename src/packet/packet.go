package packet

import "fmt"

type Packet struct {
	Name   string
	Vlan   string
	DstMac string
	SrcMac string
	DstIp  string
	SrcIp  string
	Type   string
	OldVlan string
	History string
	Loop int
}

func (p Packet) String() string {
	return fmt.Sprintf("\nPacket\nName   - %v\nVlan   - %v\nDstMac - %v\nSrcMac - %v\nDstIp  - %v\nSrcIp  - %v\nType   - %v\nOldV   - %v\n Loop %v\nHistory \n[%s]\n",
		p.Name, p.Vlan, p.DstMac, p.SrcMac, p.DstIp, p.SrcIp, p.Type, p.OldVlan, p.Loop, p.History)
}

// Change vlan
func (p *Packet) ActOuterVlan(v string) {
     p.OldVlan = p.Vlan
     p.Vlan = v
}


