module github.com/hazelcast/hazelcast-commandline-client

go 1.15

require (
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/chroma v0.9.2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hazelcast/hazelcast-go-client v1.2.0
	github.com/mattn/go-colorable v0.1.7
	github.com/mattn/go-runewidth v0.0.13
	github.com/mattn/go-tty v0.0.3
	github.com/nathan-fiscaletti/consolesize-go v0.0.0-20210105204122-a87d9f614b9d
	github.com/pkg/term v1.1.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac
	gopkg.in/yaml.v2 v2.4.0
)

// sql browser
require (
	github.com/atotto/clipboard v0.1.4
	github.com/charmbracelet/bubbles v0.10.3
	github.com/charmbracelet/bubbletea v0.20.0
	github.com/charmbracelet/lipgloss v0.5.0
	github.com/muesli/termenv v0.11.1-0.20220212125758-44cd13922739
)

// termdbms
require (
	github.com/muesli/reflow v0.3.0
	github.com/sahilm/fuzzy v0.1.0
)

replace github.com/hazelcast/hazelcast-go-client v1.2.0 => github.com/hazelcast/hazelcast-go-client v1.1.2-0.20220124142245-1906eb58ac78
