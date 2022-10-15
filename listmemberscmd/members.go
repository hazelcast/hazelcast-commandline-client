package listmemberscmd

import (
	"fmt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/connection"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
)

func New(config *hazelcast.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "list-members [distributed object type]",
		Short: "List distributed objects in the cluster",
		Long:  `List type and name of the distributed objects present in the cluster`,
		Example: `  list
  list --type map
  list --type fencedlock`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, _ := connection.ConnectToCluster(ctx, config)
			members := hazelcast.NewClientInternal(c).OrderedMembers()
			cmd.Println(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", "Order", "Private_IP", "Public_IP", "Version", "Lite?", "UUID"))
			for i, m := range members {
				l := fmt.Sprintf("%d\t%s\t%s\t%s\t%t\t%s", i, m.Addr, m.Address, m.Version, m.LiteMember, m.UUID)
				cmd.Println(l)
			}
			return nil
		},
	}
	return &cmd
}
