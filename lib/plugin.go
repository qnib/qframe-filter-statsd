package qframe_filter_statsd

import (
	"fmt"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
	"github.com/qnib/statsdaemon/lib"
)

const (
	version   = "0.1.1"
	pluginTyp = "filter"
	pluginPkg = "statsd"
)

type Plugin struct {
	qtypes.Plugin
	Statsd statsdaemon.StatsDaemon
}


func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	p := qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version)
	sdName := fmt.Sprintf("%s.%s", pluginTyp, name)
	sd := statsdaemon.NewNamedStatsdaemon(sdName, cfg, p.QChan)
	return Plugin{Plugin: p,Statsd: sd}, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start plugin v%s", p.Version))
	dc := p.QChan.Data.Join()
	go p.Statsd.Run()
	for {
		select {
		case val := <-dc.Read:
			switch val.(type) {
			case qtypes.Message:
				msg := val.(qtypes.Message)
				if p.StopProcessingMessage(msg, false) {
					continue
				}
				p.Statsd.ParseLine(msg.Message)
			case *qtypes.StatsdPacket:
				sd := val.(*qtypes.StatsdPacket)
				p.Statsd.HandlerStatsdPacket(sd)
			}
		}
	}
}
