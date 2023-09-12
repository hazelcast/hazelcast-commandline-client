package project

import "fmt"

func longHelp() string {
	help := fmt.Sprintf(`Create project from the given template.
	
This command is in BETA, it may change in future versions.
	
Templates are located in %s.
You can override it by using CLC_EXPERIMENTAL_TEMPLATE_SOURCE environment variable.

Rules while creating your own templates:

	* Templates are in Go template format.
	  See: https://pkg.go.dev/text/template
	* You can create a "defaults.yaml" file for default values in template's root directory.
	* Template files must have the ".template" extension.
	* Files with "." and "_" prefixes are ignored unless they have the ".keep" extension.
	* All files with ".keep" extension are copied by stripping the ".keep" extension.
	* Other files are copied verbatim.

Properties are read from the following resources in order:

	1. defaults.yaml (keys should be in lowercase letters, digits or underscore)
	2. config.yaml
	3. User passed key-values in the "KEY=VALUE" format. The keys can only contain lowercase letters, digits or underscore.

You can use the placeholders in "defaults.yaml" and the following configuration item placeholders:

	* cluster_name
	* cluster_address
	* cluster_user
	* cluster_password
	* cluster_discovery_token
	* cluster_api_base
	* cluster_viridian_id
	* ssl_enabled
	* ssl_server
	* ssl_skip_verify
	* ssl_ca_path
	* ssl_key_path
	* ssl_key_password
	* log_path
	* log_level

Example (Linux and macOS):

$ clc project create \
	simple-streaming-pipeline\
	--output-dir my-project\
	my_key1=my_value1 my_key2=my_value2

Example (Windows):

> clc project create^
	simple-streaming-pipeline^
	--output-dir my-project^
	my_key1=my_value1 my_key2=my_value2
`, hzTemplatesOrganization)
	return help
}
