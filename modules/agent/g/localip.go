package g

import (
	"net"
	"regexp"
	"strings"
)

// GetLocalIP returns the private & non loopback local IP of the host
func GetLocalIP() string {
	//net.Interfaces returns a list of the system's network interfaces ([]Interface).
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// inf_pattern := "#eth0$#eth1$#eth2$#eth3$#bond0$#em1$#br0$#vlanbr0$"
	// reg := regexp.MustCompile(`^(eth|bond|em|br|vlanbr)[0-9]?\..*`)

	inf_pattern := "#eth0$#eth1$#eth2$#eth3$#bond0$#em1$#br0$#vlanbr0$#p1p1$#p1p2$"
	reg := regexp.MustCompile(`^(eth|bond|em|br|vlanbr|p1p)[0-9]?\..*`)

	//iterating []Interface
	for _, inf := range ifs {
		if strings.Contains(inf_pattern, "#"+inf.Name+"$") || reg.MatchString(inf.Name) {
			addrs, err := inf.Addrs()
			if err != nil {
				return ""
			}
			//iterating []Addr
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						//check address is not public
						if strings.HasPrefix(ipnet.IP.String(), "10.") ||
							strings.HasPrefix(ipnet.IP.String(), "192.168.") ||
							strings.HasPrefix(ipnet.IP.String(), "172.") {
							return ipnet.IP.String()
						}
					}
				}
			}
		}

	}

	return ""
}
