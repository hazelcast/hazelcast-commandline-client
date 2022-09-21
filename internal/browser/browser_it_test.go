package browser_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/require"

	"github.com/hazelcast/hazelcast-commandline-client/internal/browser"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestBrowser(t *testing.T) {
	it.SQLTester(t, func(t *testing.T, client *hazelcast.Client, config *hazelcast.Config, m *hazelcast.Map, mapName string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// assert that sql query on non-existing mapping returns an error
		_, err := client.SQL().Execute(ctx, fmt.Sprintf(`select * from "%s"`, mapName))
		require.Error(t, err)
		// create a mapping via SQLBrowser
		var out bytes.Buffer
		reader, writer := io.Pipe()
		p := browser.InitSQLBrowser(client, reader, &out)
		done := make(chan error, 1)
		go func() {
			done <- p.Start()
		}()
		// send the windowsSizeMsg manually, since we do not connect a tty.
		p.Send(tea.WindowSizeMsg{
			Width:  190,
			Height: 70,
		})
		// type the query to the browser
		_, err = writer.Write([]byte(fmt.Sprintf("CREATE MAPPING \"%s\"\nTYPE IMap\nOPTIONS (\n    'keyFormat'='int',\n    'valueFormat'='varchar'\n);", mapName)))
		require.NoError(t, err)
		// send ctrl+E, execute shortcut
		_, err = writer.Write([]byte{0x5})
		require.NoError(t, err)
		// wait for browser to process the query
		time.Sleep(3 * time.Second)
		// redo the sql query but this time expect no error
		_, err = client.SQL().Execute(ctx, fmt.Sprintf(`select * from "%s"`, mapName))
		require.NoError(t, err)
		// exit from the browser, ctrl+Q
		_, err = writer.Write([]byte{0x11})
		require.NoError(t, <-done)
	})
}
