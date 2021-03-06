package main

import (
	"fmt"
	"trace"
	"time"
	"packet"
	"vtep"
	"sync"
	"strings"
//	"reflect"
)
type vtep_learns struct {
	m sync.Mutex
	vtep *vtep.Vtep
	vlearn map[string]string
	svlearn map[string]string
}
var glb_vtep_learn map[string]vtep_learns

func ( vl *vtep_learns)RcvUp(p *packet.Packet, from *vtep.Vtep, vLearn map[string]string, svLearn map[string]string) {
	// Update learn map and see where packet can be forwarded
	vtep_learn := glb_vtep_learn[from.Name]
	vtep_learn.m.Lock()
	defer vtep_learn.m.Unlock()
	// See if any mac learnt has duplicate entries
	for _, key := range vLearn {
		for _, vx := range  glb_vtep_learn {
			if vx.vtep.Name == from.Name {
			} else {
				_, ok := vx.vlearn[key]
				if ok {
					fmt.Printf("Duplicate mac learnt %s and %s\n", from.Name, vx.vtep.Name )
				}
				_, ok1 := vx.svlearn[key]
				if ok1 {
					fmt.Printf("Duplicate mac learnt(service) %s and %s\n", from.Name, vx.vtep.Name )
				}
			}
		}
	}
	vtep_learn.vlearn = vLearn
	vtep_learn.svlearn = svLearn
	//See where packet can be forwarded
	for k,v := range glb_vtep_learn {
		var learn map[string]string
		if strings.Contains(p.Vlan, "Service") {
			learn = v.svlearn
		} else {
			learn = v.vlearn
		}
		_,know := learn[p.DstMac]
		if know {
			if k == from.Name {
				// Dont reflect
			} else {
				// Send packet to vtep
				p.History += fmt.Sprintf("Packet forwarded to learnt %s\n", v.vtep.Name);
				v.vtep.TxChan <- p
				return
			}
		}
	}
	// Time to flood the packet
	for _,v := range glb_vtep_learn {
		if v.vtep.Name == from.Name {
		} else {
			//Create new packet and flood
			newP := &packet.Packet{}
			*newP = *p
			trace.T( fmt.Sprintf("Packet forwarded to flood %s\n", v.vtep.Name))
			newP.History += fmt.Sprintf("Packet forwarded to flood %s\n", v.vtep.Name)
			v.vtep.TxChan <- newP
		}
	}	
}

var vtepA, vtepB, vtepService vtep.Vtep
var vteps []*vtep.Vtep

func main() {

	vteps = make([]*vtep.Vtep, 1)
	vtepA = vtep.Vtep{Name:"VtepA"}
	vtepB = vtep.Vtep{Name:"VtepB"}
	vtepService= vtep.Vtep{Name:"VtepService"}
	
//	vtepA.Create( "Vlan10", "Vlan10Service", []string {"H1_A", "H2_A"}, []string {"H3_A", "H4_A"})
	vtepA.Create( "Vlan10", "Vlan10Service", []string {"H1_A"}, []string {"H3_A", "H4_A"})
	vteps[0] = &vtepA
	vtepB.Create( "Vlan10", "Vlan10Service", []string {"H1_B", "H2_B"}, []string {"H3_B", "H4_B"})
	vteps = append(vteps, &vtepB)
	vtepService.CreateService( "Vlan10", "Vlan10Service" )
	vteps = append(vteps, &vtepService)
	
	glb_vtep_learn = make(map[string]vtep_learns)
	vl := vtep_learns{vtep:&vtepA, vlearn:make(map[string]string), svlearn:make(map[string]string)}
	glb_vtep_learn[vtepA.Name] = vl
	vtepA.UPStream = &vl
//	glb_vtep_learn[vtepB.Name] = vtep_learns{vtep:&vtepB, vlearn:make(map[string]string), svlearn:make(map[string]string)}
	vtepB.UPStream = &vl
	glb_vtep_learn[vtepService.Name] = vtep_learns{vtep:&vtepService, vlearn:make(map[string]string), svlearn:make(map[string]string)}
	vtepService.UPStream = &vl

	h1_A,_ := vtepA.GetHost("H1_A")
	h1_A.Send("H1_A-H3_A in different Zones", "H3_A", 1 * time.Second, 1)
//	h1_A.Send("H1-H2 in same BlueZone", "H2_A", 10, 1)
//	time.Sleep(1 * time.Second )
//	fmt.Printf("******Verify H1-H2 BlueZone \n")
//	h3,_ := vtepA.GetHost("H3")
//	h3.Send("H3-H4 in same RedZone", "H4", 10, 1)
	time.Sleep(1 * time.Second )
//	fmt.Printf("******Verify H3-H4 RedZone \n")
//	h1_A.Send("H1-H3 in different Zones", "H3_A", 10, 1)
//	time.Sleep(1 * time.Second )
//	fmt.Printf("******Verify H1-H3 DifferentZone \n")
	for {
		time.Sleep(10 * time.Second )
	}
}

