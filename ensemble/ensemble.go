package ensemble

import "github.com/dacort/choirmaster/choir"
import "time"

type Servicer interface {
	Configure(config interface{})
	Run(conductor chan *choir.Note)
	PrintUpdatesSince(time.Time)
}

var services = map[string]Servicer{}

func RegisterService(name string, service Servicer) {
	services[name] = service
}

func FindService(name string) (s Servicer, ok bool) {
	s, ok = services[name]
	return
}
