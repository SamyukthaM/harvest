package main

/*  Some postprocessing on counter data "nic_common"
    Converts link_speed to numeric MBs
    Adds custom metrics:
        - "rc_percent":    receive data utilization percent
        - "tx_percent":    sent data utilization percent
        - "util_percent":  max utilization percent
        - "nic_state":     0 if port is up, 1 otherwise
*/

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/errors"
	"math"
	"strconv"
	"strings"
)

type Nic struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Nic{AbstractPlugin: p}
}

func (me *Nic) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var read, write, rx, tx, util, nic_state matrix.Metric
	var err error

	if read = data.GetMetric("rx_bytes"); read == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "rx_bytes")
	}

	if write = data.GetMetric("tx_bytes"); write == nil {
		return nil, errors.New(errors.ERR_NO_METRIC, "tx_bytes")
	}

	if rx = data.GetMetric("rx_percent"); rx == nil {
		if rx, err = data.AddMetricFloat64("rx_percent"); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, err
		}

	}
	if tx = data.GetMetric("tx_percent"); tx == nil {
		if tx, err = data.AddMetricFloat64("tx_percent"); err == nil {
			tx.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	if util = data.GetMetric("util_percent"); util == nil {
		if util, err = data.AddMetricFloat64("util_percent"); err == nil {
			util.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	if nic_state = data.GetMetric("state"); nic_state == nil {
		if nic_state, err = data.AddMetricUint8("state"); err == nil {
			nic_state.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		var speed, base int
		var s string
		var err error

		if s = instance.GetLabel("speed"); strings.HasSuffix(s, "M") {
			base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
			if err != nil {
				logger.Warn(me.Prefix, "convert speed [%s]", s)
			} else {
				speed = base * 125000
				instance.SetLabel("speed", strconv.Itoa(speed))
				logger.Debug(me.Prefix, "converted speed (%s) to numeric (%d)", s, speed)
			}
		} else if speed, err = strconv.Atoi(s); err != nil {
			logger.Warn(me.Prefix, "convert speed [%s]", s)
		}

		if speed != 0 {

			var rx_bytes, tx_bytes, rx_percent, tx_percent float64
			var rx_ok, tx_ok bool

			if rx_bytes, rx_ok = read.GetValueFloat64(instance); rx_ok {
				rx_percent = rx_bytes / float64(speed)
				rx.SetValueFloat64(instance, rx_percent)
			}

			if tx_bytes, tx_ok = write.GetValueFloat64(instance); tx_ok {
				tx_percent = tx_bytes / float64(speed)
				tx.SetValueFloat64(instance, tx_percent)
			}

			if rx_ok || tx_ok {
				util.SetValueFloat64(instance, math.Max(rx_percent, tx_percent))
			}
		}

		if instance.GetLabel("state")== "up" {
			nic_state.SetValueUint8(instance, 0)
		} else {
			nic_state.SetValueUint8(instance, 1)
		}

		// truncate redundant prefix in nic type
		if t := instance.GetLabel("type"); strings.HasPrefix(t, "nic_") {
			instance.SetLabel("type", strings.TrimPrefix(t, "nic_"))
		}

	}

	return nil, nil
}
