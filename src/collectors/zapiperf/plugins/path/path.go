package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/matrix"
	"strings"
)

type Path struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Path{AbstractPlugin: p}
}

func (me *Path) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		// no need to continue if labels are already parsed
		if instance.GetLabel("target_wwpn") != "" {
			break
		}

		name := instance.GetLabel("path")

		// example name = 1a.2100001086a45d80
		// hostadapter  => 1a
		// targetwwpn   => 2100001086a45d80

		if split := strings.Split(name, "."); len(split) == 2 {
			instance.SetLabel("hostadapter", split[0])
			instance.SetLabel("target_wwpn", split[1])
		} else if split := strings.Split(name, "_"); len(split) == 2 {
			instance.SetLabel("hostadapter", split[0])
			instance.SetLabel("target_wwpn", split[1])
		}
	}

	return nil, nil
}
