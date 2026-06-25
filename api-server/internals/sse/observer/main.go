package observer

import "github.com/uddinArsalan/devdeploy/internals/domain"

type Observer interface{
	Notify(deployID int64,event domain.LogEvent) 
}