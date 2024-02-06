package gate

import (
	"fmt"
	"net"
)

type Option struct {
	Name     string
	TcpPorts []uint16
	UdpPorts []uint16
}

func (o *Option) Clone() *Option {
	n := &Option{Name: o.Name}
	n.TcpPorts = append(n.TcpPorts, o.TcpPorts...)
	n.UdpPorts = append(n.UdpPorts, o.UdpPorts...)
	return n
}

type Gate struct {
	*Option
}

func (gate *Gate) Name4() string {
	return fmt.Sprintf("%v4", gate.Name)
}

func (gate *Gate) Name6() string {
	return fmt.Sprintf("%v6", gate.Name)
}

func NewGate(option *Option) *Gate {
	return &Gate{Option: option.Clone()}
}

func (gate *Gate) Add(addr net.IP) error {
	if addr.To4() == nil {
		return RunCmd("ipset", "add", gate.Name6(), addr.String(), "-exist")
	}
	return RunCmd("ipset", "add", gate.Name4(), addr.String(), "-exist")
}

func (gate *Gate) Delete(addr net.IP) error {
	if addr.To4() == nil {
		return RunCmd("ipset", "del", gate.Name6(), addr.String(), "-exist")
	}
	return RunCmd("ipset", "del", gate.Name4(), addr.String(), "-exist")
}

func (gate *Gate) Init() *Gate {
	_ = RunCmd("iptables", "-D", "INPUT", "-m", "set", "--match-set", gate.Name4(), "src", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-D", "INPUT", "-m", "set", "--match-set", gate.Name6(), "src", "-j", "ACCEPT")
	_ = RunCmd("ipset", "destroy", gate.Name4())
	_ = RunCmd("ipset", "destroy", gate.Name6(), "family", "inet6")

	_ = RunCmd("iptables", "-D", "INPUT", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-D", "INPUT", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = RunCmd("iptables", "-D", "INPUT", "-i", "lo", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-D", "INPUT", "-i", "lo", "-j", "ACCEPT")

	for _, v := range gate.TcpPorts {
		_ = RunCmd("iptables", "-D", "INPUT", "-p", "tcp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
		_ = RunCmd("ip6tables", "-D", "INPUT", "-p", "tcp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
	}

	for _, v := range gate.UdpPorts {
		_ = RunCmd("iptables", "-D", "INPUT", "-p", "udp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
		_ = RunCmd("ip6tables", "-D", "INPUT", "-p", "udp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
	}

	_ = RunCmd("ipset", "create", gate.Name4(), "hash:ip")
	_ = RunCmd("ipset", "create", gate.Name6(), "hash:ip", "family", "inet6")
	_ = RunCmd("iptables", "-I", "INPUT", "1", "-m", "set", "--match-set", gate.Name4(), "src", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-I", "INPUT", "1", "-m", "set", "--match-set", gate.Name6(), "src", "-j", "ACCEPT")
	_ = RunCmd("iptables", "-A", "INPUT", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-A", "INPUT", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	_ = RunCmd("iptables", "-A", "INPUT", "-i", "lo", "-j", "ACCEPT")
	_ = RunCmd("ip6tables", "-A", "INPUT", "-i", "lo", "-j", "ACCEPT")

	for _, v := range gate.TcpPorts {
		_ = RunCmd("iptables", "-A", "INPUT", "-p", "tcp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
		_ = RunCmd("ip6tables", "-A", "INPUT", "-p", "tcp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
	}

	for _, v := range gate.UdpPorts {
		_ = RunCmd("iptables", "-A", "INPUT", "-p", "udp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
		_ = RunCmd("ip6tables", "-A", "INPUT", "-p", "udp", "--dport", fmt.Sprintf("%d", v), "-j", "DROP")
	}

	return gate
}
