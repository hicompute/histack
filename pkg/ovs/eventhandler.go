package ovs

import (
	ovsModels "github.com/hicompute/histack/pkg/ovs/models"
	"github.com/ovn-kubernetes/libovsdb/model"
)

type ovsEventHandler struct {
	agent *OvsAgent
}

func (h *ovsEventHandler) OnAdd(table string, m model.Model) { h.handle(table, m) }
func (h *ovsEventHandler) OnUpdate(table string, old, new model.Model) {
	h.handle(table, new)
}
func (h *ovsEventHandler) OnDelete(table string, m model.Model) {}

func (h *ovsEventHandler) handle(table string, m model.Model) {
	switch obj := m.(type) {

	case *ovsModels.Interface:
		h.agent.updateInterfaceStats(obj)

	case *ovsModels.Port:
		for _, ifaceUUID := range obj.Interfaces {
			h.agent.ifaceToPort[ifaceUUID] = obj.UUID
		}

	case *ovsModels.Bridge:
		for _, portUUID := range obj.Ports {
			h.agent.portToBridge[portUUID] = obj.Name
		}
	}
}
