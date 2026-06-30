package main

import (
	"context"
	"sync"

	"github.com/moby/moby/client"
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/repository"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

type Dispatcher struct {
	wg           *sync.WaitGroup
	ctx          context.Context
	queue        queue.Queue
	cache        cache.Cache
	client       *client.Client
	portMap      *utils.PortMap
	deployRepo   *repository.DeploymentRepository
	envRepo      *repository.EnvRepo
	numOfWorkers int
}

func NewDispatcher(
	wg *sync.WaitGroup,
	ctx context.Context,
	numOfWorkers int,
	client *client.Client,
	portMap *utils.PortMap,
	deployRepo *repository.DeploymentRepository,
	envRepo *repository.EnvRepo,
	queue queue.Queue,
	cache cache.Cache,
) *Dispatcher {

	return &Dispatcher{
		wg,
		ctx,
		queue,
		cache,
		client,
		portMap,
		deployRepo,
		envRepo,
		numOfWorkers,
	}
}

func (d *Dispatcher) Start() {
	for i := range d.numOfWorkers {
		d.wg.Add(1)
		w := DeployWorker{
			Id:         i,
			wg:         d.wg,
			client:     d.client,
			portMap:    d.portMap,
			deployRepo: d.deployRepo,
			queue:      d.queue,
			cache:      d.cache,
			envRepo:    d.envRepo,
		}
		go w.DeployBuildWorker(d.ctx)
	}
}
