//go:build base || cluster

package _cluster

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type ClusterListMembersCommand struct{}

func (mc *ClusterListMembersCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(0, 0)
	help := "List the members of the cluster"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("list-members [flags]")
	return nil
}

func (mc *ClusterListMembersCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	infos, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		cn := ci.ClusterService().FailoverService().Current().ClusterName
		sp.SetText(fmt.Sprintf("Getting member list for cluster: %s", cn))
		return memberInfos(ctx, ci)
	})
	if err != nil {
		return err
	}
	inf := infos.(map[types.UUID]*memberData)
	rows := []output.Row{}
	for uuid, info := range inf {
		row := output.Row{
			output.Column{
				Name:  "Order",
				Value: info.Order,
				Type:  serialization.TypeInt64,
			},
			output.Column{
				Name:  "UUID",
				Value: uuid,
				Type:  serialization.TypeUUID,
			},
			output.Column{
				Name:  "Public Address",
				Value: info.PublicAddress,
				Type:  serialization.TypeString,
			},
			output.Column{
				Name:  "Private Address",
				Value: info.PrivateAddress,
				Type:  serialization.TypeString,
			},
			output.Column{
				Name:  "Hazelcast Version",
				Value: info.Version,
				Type:  serialization.TypeString,
			},
			output.Column{
				Name:  "IsMaster",
				Value: info.Master,
				Type:  serialization.TypeBool,
			},
			output.Column{
				Name:  "IsLite",
				Value: info.LiteMember,
				Type:  serialization.TypeBool,
			},
		}
		if ec.Props().GetBool(clc.PropertyVerbose) {
			row = append(row,
				output.Column{
					Name:  "Member State",
					Value: info.MemberState,
					Type:  serialization.TypeString,
				},
				output.Column{
					Name:  "Name",
					Value: info.Name,
					Type:  serialization.TypeString,
				},
			)
		}
		rows = append(rows, row)
	}

	ec.AddOutputRows(ctx, rows...)
	if err != nil {
		return err
	}
	stop()
	return nil
}

type memberData struct {
	Order          int64
	PrivateAddress string
	PublicAddress  string
	UUID           string
	Version        string
	LiteMember     bool
	MemberState    string
	Master         bool
	Name           string
}

func newMemberData(order int64, m cluster.MemberInfo, s control.TimedMemberState) *memberData {
	priv, pub := findMemberAddresses(m)
	return &memberData{
		Order:          order,
		PrivateAddress: priv,
		PublicAddress:  pub,
		UUID:           m.UUID.String(),
		Version:        fmt.Sprintf("%d.%d.%d", m.Version.Major, m.Version.Minor, m.Version.Patch),
		LiteMember:     m.LiteMember,
		Master:         s.Master,
		MemberState:    s.MemberState.NodeState.State,
		Name:           s.MemberState.Name,
	}
}

func findMemberAddresses(m cluster.MemberInfo) (string, string) {
	pub := ""
	priv := ""
	for key, val := range m.AddressMap {
		if key.Type != cluster.EndpointQualifierTypeClient {
			continue
		}
		switch key.Identifier {
		case "":
			priv = val.String()
		case "public":
			pub = val.String()
		}
	}
	return priv, pub
}

func memberInfos(ctx context.Context, ci *hazelcast.ClientInternal) (map[types.UUID]*memberData, error) {
	activeMemberList := ci.OrderedMembers()
	activeMembers := make(map[types.UUID]*memberData, len(activeMemberList))
	for i, memberInfo := range activeMemberList {
		state, err := fetchTimedMemberState(ctx, ci, memberInfo.UUID)
		if err != nil {
			return nil, err
		}
		this := memberInfo.Address
		returned := state.TimedMemberState.MemberState.Address
		if string(this) != string(returned) {
			return nil, fmt.Errorf("Timed member state returned info for wrong member, this: %s, returned: %s", this, returned)
		}
		activeMembers[memberInfo.UUID] = newMemberData(int64(i), memberInfo, state.TimedMemberState)
	}
	return activeMembers, nil
}

func fetchTimedMemberState(ctx context.Context, ci *hazelcast.ClientInternal, uuid types.UUID) (*control.TimedMemberStateWrapper, error) {
	req := codec.EncodeMCGetTimedMemberStateRequest()
	resp, err := ci.InvokeOnMember(ctx, req, uuid, nil)
	if err != nil {
		return nil, err
	}
	jsonState := codec.DecodeMCGetTimedMemberStateResponse(resp)
	state, err := codec.DecodeTimedMemberStateJsonString(jsonState)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("cluster:list-members", &ClusterListMembersCommand{}))
}
