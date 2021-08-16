package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/commands/internal"
	"github.com/spf13/cobra"
)

var (
	newState              string
	clusterchangestateCmd = &cobra.Command{
		Use:   "clusterchangestate",
		Short: "change state of your local cluster",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := retrieveFlagValues(cmd)
			if err != nil {
				log.Fatal(err)
			}
			var params string
			switch newState {
			case internal.Active, internal.NoMigration, internal.Frozen, internal.Passive:
				params = fmt.Sprintf("%s&%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password, newState)
			default:
				log.Fatal("Invalid new state.")
			}
			pr := strings.NewReader(params)
			url := fmt.Sprintf("%s?%s", "http://127.0.0.1:5701/hazelcast/rest/management/cluster/changeState", params)
			resp, err := http.Post(url, "application/x-www-form-urlencoded", pr)
			if err != nil {
				log.Fatal(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			sb := string(body)
			fmt.Println(sb)
		},
	}
)

func init() {
	clusterCmd.AddCommand(clusterchangestateCmd)
	clusterCmd.PersistentFlags().StringVarP(&newState, "state", "s", "", "new state of the cluster")
	clusterCmd.MarkFlagRequired("state")
}
