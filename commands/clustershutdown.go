package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var clustershutdownCmd = &cobra.Command{
	Use:   "clustershutdown",
	Short: "shutdown your local cluster",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := retrieveFlagValues(cmd)
		if err != nil {
			log.Fatal(err)
		}
		params := fmt.Sprintf("%s&%s", config.Cluster.Name, config.Cluster.Security.Credentials.Password)
		pr := strings.NewReader(params)
		url := fmt.Sprintf("%s?%s", "http://127.0.0.1:5701/hazelcast/rest/management/cluster/clusterShutdown", params)
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

func init() {
	clusterCmd.AddCommand(clustershutdownCmd)
}
