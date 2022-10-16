package moreoutputtypes

import (
	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	plug.Registry.RegisterPrinter("html", &base.TablePrinter{Mode: output.TableOutputModeHTML})
	plug.Registry.RegisterPrinter("markdown", &base.TablePrinter{Mode: output.TableOutputModeMarkDown})
}
