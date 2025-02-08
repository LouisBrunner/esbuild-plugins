# esbuild-plugins

A collection of Go plugins for esbuild.

## Plugins

### `postcss`

This plugin can be used to automatically process included CSS files using `postcss`. It supports a few options:

* `Command`: defaults to `npx postcss`, the command to run postcss. This must be a command name with all required arguments. Quotes are supported (pseudo-shell).
* `Filter`: defaults to `\.(s?css|sass)$`, a regular expression to match files that should be processed by postcss.

```go
package main

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/LouisBrunner/esbuild-plugins/pkg/postcss"
)

func main() {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"app.js"},
		Bundle:      true,
		Outfile:     "out.js",
		Plugins: []api.Plugin{postcss.Must(postcss.NewPlugin(postcss.Options{
			Command: "npx postcss",
			Filter:  `\.(s?css|sass)$`,
		}))},
		Write: true,
	})

	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}
```

Notes:

* By default, this plugin assumes you have `npx` and `postcss` installed and in your PATH.
