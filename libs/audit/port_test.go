package audit

import "testing"

func TestPortsEqual(t *testing.T) {
	equalPortsTest1 := portsEqual([]uint16{22, 80, 443}, []uint16{80, 443, 22})
	if !equalPortsTest1 {
		t.Errorf(
			"portsEqual(([]uint16{22, 80, 443}, []uint16{80, 443, 22}) = %t; want true",
			equalPortsTest1,
		)
	}

	equalPortsTest2 := portsEqual([]uint16{22}, []uint16{80})
	if equalPortsTest2 {
		t.Errorf(
			"portsEqual([]uint16{22}, []uint16{80}) = %t; want false",
			equalPortsTest2,
		)
	}

	equalPortsTest3 := portsEqual([]uint16{22}, []uint16{80, 22})
	if equalPortsTest3 {
		t.Errorf(
			"portsEqual([]uint16{22}, []uint16{80, 22}) = %t; want false",
			equalPortsTest3,
		)
	}

	equalPortsTest4 := portsEqual([]uint16{}, []uint16{})
	if !equalPortsTest4 {
		t.Errorf(
			"portsEqual(([]uint16{}, []uint16{}) = %t; want true",
			equalPortsTest4,
		)
	}
}
