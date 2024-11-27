# esbuild-plugins

A collection of Go plugins for esbuild.

## Plugins

### `postcss`

This plugin can be used to automatically process included CSS files using `postcss`.

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
    Plugins:     []api.Plugin{postcss.Plugin},
    Write:       true,
  })

  if len(result.Errors) > 0 {
    os.Exit(1)
  }
}
```

Notes:

* This plugin assumes that you have `npx` installed (which is available through modern `npm` distribution).
* This plugin assumes that you have `postcss` installed in your project.
