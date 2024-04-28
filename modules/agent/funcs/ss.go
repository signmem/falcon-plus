package funcs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type tcpinfo struct {
	RtoAlgorithm int
	RtoMin       int
	RtoMax       int
	MaxConn      int
	ActiveOpens  int
	PassiveOpens int
	AttemptFails int
	EstabResets  int
	CurrEstab    int
	InSegs       int
	OutSegs      int
	RetransSegs  int
	InErrs       int
	OutRsts      int
	InCsumErrors int
}

type ssummary struct {
	socks       int
	tcp_mem     int
	tcp_total   int
	tcp_orphans int
	tcp_tws     int
	tcp4_hashed int
	udp4        int
	raw4        int
	frag4       int
	frag4_mem   int
	tcp6_hashed int
	udp6        int
	raw6        int
	frag6       int
	frag6_mem   int
}

func get_tcp_info_from_snmp(filePath string) (u *tcpinfo, err error) {
	var i tcpinfo

	bytev, err := ioutil.ReadFile(filePath)

	if err != nil {
		return &i, err
	}

	if len(bytev) == 0 {
		errstrings := "get_tcp_info_from_snmp() error file read error "
		errstrings += filePath
		error := errors.New(errstrings)
		return &i, error
	}

	txt := strings.Split(string(bytev), "\n")
	for _, v := range txt {
		fmt.Sscanf(v, "Tcp: %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d", &i.RtoAlgorithm, &i.RtoMin, &i.RtoMax, &i.MaxConn, &i.ActiveOpens, &i.PassiveOpens, &i.AttemptFails, &i.EstabResets, &i.CurrEstab, &i.InSegs, &i.OutSegs, &i.RetransSegs, &i.InErrs, &i.OutRsts, &i.InCsumErrors)
	}
	return &i, nil
}


func get_sockstat_line(filePath string) (u *ssummary, err error) {

	var tcpDump ssummary

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &tcpDump, err
	}

	bytev, _ := ioutil.ReadFile(filePath)

	if len(bytev) == 0 {
		errstrings := "get_sockstat_line() error, nothings in "
		errstrings +=  filePath
		return  &tcpDump, errors.New( errstrings )
	}

	txt := strings.Split(string(bytev), "\n")
	for _, line := range txt {
		if strings.Contains(line, "sockets:") {
			fmt.Sscanf(line, "sockets: used %d", &tcpDump.socks)
		} else if strings.Contains(line, "UDP:")  {
			fmt.Sscanf(line, "UDP: inuse %d", &tcpDump.udp4)
		} else if strings.Contains(line, "UDP6:")  {
			fmt.Sscanf(line, "UDP6: inuse %d", &tcpDump.udp6)
		} else if strings.Contains(line, "RAW:") {
			fmt.Sscanf(line, "RAW: inuse %d", &tcpDump.raw4)
		} else if strings.Contains(line, "RAW6:") {
			fmt.Sscanf(line, "RAW6: inuse %d", &tcpDump.raw4)
		} else if strings.Contains(line, "FRAG:")  {
			fmt.Sscanf(line, "FRAG: inuse %d memory %d", &tcpDump.frag4, &tcpDump.frag4_mem)
		} else if strings.Contains(line, "FRAG6:")  {
			fmt.Sscanf(line, "FRAG6: inuse %d memory %d", &tcpDump.frag6, &tcpDump.frag6_mem)
		} else if strings.Contains(line, "TCP6:")  {
			fmt.Sscanf(line, "TCP6: inuse %d", &tcpDump.tcp6_hashed)
		} else if strings.Contains(line, "TCP:")  {
			fmt.Sscanf(line, "TCP: inuse %d orphan %d tw %d alloc %d mem %d", &tcpDump.tcp4_hashed, &tcpDump.tcp_orphans, &tcpDump.tcp_tws, &tcpDump.tcp_total, &tcpDump.tcp_mem)
		}
	}

	return &tcpDump, nil

}

func getSS() (m(map[string]uint64), err error) {

	// use to replace ss -s command
	// skip read /proc/slabinfo file
	// define slabinfo.timewait   synrecv to zero

	m = make(map[string]uint64)

	snmpFile := "/proc/net/snmp"
	ti, err := get_tcp_info_from_snmp(snmpFile)

	if err != nil {
		log.Println("getSS() error ",  err)
		return m, err
	}

	sockFile := "/proc/net/sockstat"
	sock6File := "/proc/net/sockstat6"

	tcpStat, err := get_sockstat_line(sockFile)

	if err !=nil {
		log.Println("tcpStat get_sockstat_line() error", err)
		return m, err
	}

	tcp6Stat, err := get_sockstat_line(sock6File)

	if err != nil {
		m["closed"] = uint64(tcpStat.tcp_total-(tcpStat.tcp4_hashed-tcpStat.tcp_tws))
	}else{
		m["closed"] = uint64(tcpStat.tcp_total-(tcpStat.tcp4_hashed+tcp6Stat.tcp6_hashed-tcpStat.tcp_tws))
	}


	m["estab"] = uint64(ti.CurrEstab)
	m["orphaned"] = uint64(tcpStat.tcp_orphans)
	m["slabinfo.timewait"] =  uint64(0)
	m["synrecv"] = uint64(0)
	m["timewait"] = uint64(tcpStat.tcp_tws)

	return m, nil
}
