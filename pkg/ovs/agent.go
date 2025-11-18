package ovs

import (
	"context"
	"log"

	ovsModels "github.com/hicompute/histack/pkg/ovs/models"
	"github.com/ovn-kubernetes/libovsdb/client"
	"github.com/ovn-kubernetes/libovsdb/model"
)

type OvsAgent struct {
	ovsClient client.Client
}

func CreateOVSagent() (*OvsAgent, error) {
	dbModel, err := model.NewClientDBModel("Open_vSwitch", map[string]model.Model{
		ovsModels.BridgeTable:    &ovsModels.Bridge{},
		ovsModels.PortTable:      &ovsModels.Port{},
		ovsModels.InterfaceTable: &ovsModels.Interface{},
	})
	dbModel.SetIndexes(map[string][]model.ClientIndex{
		ovsModels.PortTable: {{Columns: []model.ColumnKey{
			{Column: "external_ids", Key: "iface-id"},
		}}},
	})
	if err != nil {
		log.Printf("failed to create DB model: %v", err)
		return nil, err
	}

	ovsClient, err := client.NewOVSDBClient(
		dbModel,
		client.WithEndpoint("unix:/var/run/openvswitch/db.sock"),
	)
	if err != nil {
		log.Printf("failed to create OVS client: %v", err)
		return nil, err
	}

	ctx := context.Background()
	if err := ovsClient.Connect(ctx); err != nil {
		log.Printf("failed to connect to OVSDB: %v", err)
		return nil, err
	}
	_, err = ovsClient.MonitorAll(ctx)
	if err != nil {
		log.Printf("failed to monitor OVSDB: %v", err)
		return nil, err
	}
	return &OvsAgent{ovsClient: ovsClient}, nil
}

func (oa *OvsAgent) Close() {
	oa.ovsClient.Disconnect()
}
