package utils

import "sync"

// for now I keep the port to host map locally
type Domain struct {
	ProjectID   int64
	ContainerID string
	Port        int
}

type PortMap struct {
	MinPort        int
	MaxPort        int
	mu             sync.Mutex
	AvailablePorts map[int]bool
	PortMapping    map[string]Domain // map of hostname to domain info
}

func NewPortMap(minPort int, maxPort int) *PortMap {
	portMap := PortMap{MinPort: minPort, MaxPort: maxPort}
	portMap.AvailablePorts = make(map[int]bool)
	for j := minPort; j <= maxPort; j++ {
		portMap.AvailablePorts[j] = true
	}
	portMap.PortMapping = make(map[string]Domain)
	return &portMap
}

func (p *PortMap) GetPort() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	for port, avlbl := range p.AvailablePorts {
		if avlbl {
			p.AvailablePorts[port] = false
			return port
		}
	}
	return -1
}

func (p *PortMap) AssignProjectIDToDomain(projectID int64, hostname string, containerID string, port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.PortMapping[hostname] = Domain{
		ProjectID:   projectID,
		ContainerID: containerID,
		Port:        port,
	}
}

func (p *PortMap) ReleasePort(hostname string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.PortMapping[hostname] = Domain{}
}

func (p *PortMap) GetPortDomain(hostname string) Domain {
	return p.PortMapping[hostname]
}
