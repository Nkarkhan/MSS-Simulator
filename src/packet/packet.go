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


type PktCounts struct {
	rx int
	tx int
	drop int
	flood int
}

func (pc PktCounts) String() string {
	s:=""
	if (pc.rx != 0 || pc.tx !=0 || pc.drop!= 0) {
		s += fmt.Sprintf("Receive Packets   - %v\n Transmit Packets - %v\n Drop Packets - %v\n",
			pc.rx, pc.tx, pc.drop)
	}
	return s
}

func (pc PktCounts) PktCounter() bool {
	return (pc.rx != 0 || pc.tx !=0 || pc.drop!= 0)
}

func (pc *PktCounts) RxIncr() {
	pc.rx++
}

func (pc *PktCounts) RxGet() int{
	return pc.rx
}

func (pc *PktCounts) TxIncr() {
	pc.tx++
}

func (pc *PktCounts) TxGet() int{
	return pc.tx
}

func (pc *PktCounts) DropIncr() {
	pc.drop++
}

func (pc *PktCounts) DropGet() int{
	return pc.drop
}

func (pc *PktCounts) FloodIncr() {
	pc.flood++
}

func (pc *PktCounts) FloodGet() int{
	return pc.flood
}

func (pc *PktCounts) Reset(){
	pc.rx = 0
	pc.tx = 0
	pc.drop = 0
	pc.flood = 0
}

