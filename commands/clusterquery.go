package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var clusterqueryCmd = &cobra.Command{
	Use:   "clusterquery",
	Short: "retrieve information from your local cluster",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get("http://127.0.0.1:5701/hazelcast/rest/management/cluster/version")
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		str := string(body)
		fmt.Println(str)
	},
}

func init() {
	clusterCmd.AddCommand(clusterqueryCmd)
}
