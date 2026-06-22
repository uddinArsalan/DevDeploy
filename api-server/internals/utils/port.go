package utils

import "sync"

// for now I keep the port to host map locally
type Domain struct {
	ProjectID    string
	ContainerID string
	Port        string
}

type PortMap struct {
	MinPort        int64
	MaxPort        int64
	mu             sync.Mutex
	AvailablePorts map[int64]bool
	PortMapping    map[string]Domain // map of hostname to domain info
}

func NewPortMap(minPort int64, maxPort int64) *PortMap {
	portMap := PortMap{MinPort: minPort, MaxPort: maxPort}
	portMap.AvailablePorts = make(map[int64]bool)
	for j := minPort; j <= maxPort; j++ {
		portMap.AvailablePorts[j] = true
	}
	portMap.PortMapping = make(map[string]Domain)
	return &portMap
}

func (p *PortMap) GetPort() int64 {
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

func (p *PortMap) AssignProjectIDToDomain(projectID string, hostname string, containerID string, port string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.PortMapping[hostname] = Domain{
		ProjectID:    projectID,
		ContainerID: containerID,
		Port:        port,
	}
}

func (p *PortMap) ReleasePort(hostname string){
	p.mu.Lock()
	defer p.mu.Unlock()
	p.PortMapping[hostname] = Domain{}
}

func (p *PortMap) GetPortDomain(hostname string) Domain {
	return p.PortMapping[hostname]
}
