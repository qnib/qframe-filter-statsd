package qframe_filter_statsd

import (
	"fmt"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
	"github.com/qnib/statsdaemon/lib"
	"time"
)

const (
	version   = "0.0.0"
	pluginTyp = "filter"
	pluginPkg = "statsd"
)

type Plugin struct {
	qtypes.Plugin
	Statsd statsdaemon.StatsDaemon
}


func New(qChan qtypes.QChan, cfg *config.Config, name string) (p Plugin, err error) {
	p = Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
		Statsd: statsdaemon.NewNamedStatsdaemon(p.Name, cfg, p.QChan),
	}
	return
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start plugin v%s", p.Version))
	dc := p.QChan.Data.Join()
	inputs := p.GetInputs()
	srcSuccess := p.CfgBoolOr("source-success", true)
	tickMs := p.CfgIntOr("send-metric-ms", 1000)
	ticker := time.NewTicker(time.Duration(tickMs)*time.Millisecond).C
	go p.Statsd.Run()
	for {
		select {
		case val := <-dc.Read:
			switch val.(type) {
			case qtypes.Message:
				msg := val.(qtypes.Message)
				if msg.IsLastSource(p.Name) {
					p.Log("debug", "IsLastSource() = true")
					continue
				}
				if len(inputs) != 0 && ! msg.InputsMatch(inputs) {
					p.Log("debug", fmt.Sprintf("InputsMatch(%v) = false", inputs))
					continue
				}
				if msg.SourceSuccess != srcSuccess {
					p.Log("debug", "qcs.SourceSuccess != srcSuccess")
					continue
				}

				p.Statsd.ParseLine(msg.Message)
			}
		case <-ticker:
			p.Statsd.FanOutMetrics()
		}
	}
}
