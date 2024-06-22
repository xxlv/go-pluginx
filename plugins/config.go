package plugins

import "fmt"

type MetadataPliuginsKey string

var MetadataPliuginsKeyFmt MetadataPliuginsKey = "tools_metadata_%s"

var (
	SoFileBuildPath = "data/dynamiclib"
	BuildInfoPrefix = "x-buildId-"
)

func (mpk MetadataPliuginsKey) Key(tenant string) string {
	return fmt.Sprintf(string(mpk), tenant)
}
