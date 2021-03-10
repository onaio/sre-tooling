package audit

import "testing"

func TestComparePorts(t *testing.T) {
	userPorts := []string{"tcp/22", "tcp/80", "tcp/443"}
	nmapPorts := []string{"tcp/443", "tcp/80", "tcp/22"}
	isSimilar := comparePorts(userPorts, nmapPorts)
	if !isSimilar {
		t.Errorf(
			"comparePorts(%v, %v) = %t; want true",
			userPorts, nmapPorts, isSimilar,
		)
	}

	userPorts = []string{"tcp/22", "tcp/80", "tcp/443", "udp/123"}
	nmapPorts = []string{"tcp/443", "tcp/80", "tcp/22"}
	isSimilar = comparePorts(userPorts, nmapPorts)
	if isSimilar {
		t.Errorf(
			"comparePorts(%v, %v) = %t; want false",
			userPorts, nmapPorts, isSimilar,
		)
	}

	userPorts = []string{"tcp/22", "tcp/80", "tcp/443"}
	nmapPorts = []string{"tcp/443", "tcp/80", "tcp/22", "udp/123"}
	isSimilar = comparePorts(userPorts, nmapPorts)
	if isSimilar {
		t.Errorf(
			"comparePorts(%v, %v) = %t; want false",
			userPorts, nmapPorts, isSimilar,
		)
	}

	userPorts = []string{"tcp/22", "tcp/80", "tcp/443", "udp/6000-6002"}
	nmapPorts = []string{"tcp/443", "tcp/80", "tcp/22", "udp/6000", "udp/6001", "udp/6002"}
	isSimilar = comparePorts(userPorts, nmapPorts)
	if !isSimilar {
		t.Errorf(
			"comparePorts(%v, %v) = %t; want true",
			userPorts, nmapPorts, isSimilar,
		)
	}

	userPorts = []string{"tcp/22", "tcp/80", "tcp/443", "udp/6000-6002"}
	nmapPorts = []string{"tcp/443", "tcp/80", "tcp/22", "udp/6000", "udp/6002"}
	isSimilar = comparePorts(userPorts, nmapPorts)
	if isSimilar {
		t.Errorf(
			"comparePorts(%v, %v) = %t; want false",
			userPorts, nmapPorts, isSimilar,
		)
	}
}
