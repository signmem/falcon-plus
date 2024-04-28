package g

import (
        "bufio"
        "bytes"
        "errors"
        "io/ioutil"
        "net"
        "os"
        "regexp"
        "strings"
)

const (
        routerfile  = "/proc/net/route"
)

func parseLinuxProcNetRoute(f []byte) (string, error) {
        const (
                sep   = "\t" // field separator
                field = 2    // field containing hex gateway address
                flag = 3     // field containing hex gateway flag
                name = 0     // field containing interface name
        )

        scanner := bufio.NewScanner(bytes.NewReader(f))
        for scanner.Scan() {
                tokens := strings.Split(scanner.Text(), sep)
                if tokens[flag] == "0003" {
                        interfaceName := tokens[name]
                        return  interfaceName, nil
                }
        }

        return "", errors.New("Can not get gateway record")
}


func getGateInterFaceIPAddress(gatewayInterface string) (string, error) {
        interfaces, _ := net.Interfaces()

        for _, i := range interfaces {
                if i.Name == gatewayInterface {
                        byNameInterface, _ := net.InterfaceByName(i.Name)
                        addresses, _ := byNameInterface.Addrs()
                        if len(addresses) > 0 {
                                for _,addr := range addresses {
                                        ipv4AddrMask := addr.String()
                                        if strings.Count(ipv4AddrMask,":") < 2 {
                                                ipv4Addr, _, err := net.ParseCIDR(ipv4AddrMask)
                                                if err != nil {
                                                        return "", errors.New("Failed to change into IP")
                                                }
                                                return ipv4Addr.String(), nil
                                        }
                                }
                        }else{
                                return "", errors.New("Failed to get IP")
                        }
                }
        }
        return "", errors.New("Failed to get interface")
}

func getLinuxIPaddress()(string, error) {
        f, _:= os.Open(routerfile)
        defer f.Close()
        bytes, _ := ioutil.ReadAll(f)
        interfaceName, err := parseLinuxProcNetRoute(bytes)
        if err != nil {
		return "", err
        }
        addr, err := getGateInterFaceIPAddress(interfaceName)
	if err != nil {
		return "", err
	}
	return addr, err
}

func getSSHIpaddr() (string, error)   {
        sshConfig := "/etc/ssh/sshd_config"
        sshF, err := os.Open(sshConfig)

        defer sshF.Close()

        if err != nil {
            return "", errors.New("ssh config file not exists.")
        }

        filebytes, _ := ioutil.ReadAll(sshF)

        re := regexp.MustCompile("^ListenAddress")

        scanner := bufio.NewScanner(bytes.NewReader(filebytes))
        for scanner.Scan() {
            lines := strings.Split(scanner.Text(), "\n")
            for _, k := range lines {
                status := re.FindString(k)
                if status != "" {
                    sshAddrInfo := strings.Split(k," ")
                    infoLength := len(sshAddrInfo)
                    if infoLength > 1 {
                        ind := infoLength - 1
                        localIPaddr := sshAddrInfo[ind]
                        if  localIPaddr !=  "0.0.0.0" {
                            return localIPaddr, nil
                        } else {
                            return "", errors.New("ssh listen 0.0.0.0 error.")
                        }
                    }
                }
            }
        }
        return "", errors.New("can not get ip addr from sshd_config")

}
