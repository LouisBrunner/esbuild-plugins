# esbuild-plugins

A collection of Go plugins for esbuild.

## Plugins

### `postcss`

This plugin can be used to automatically process included CSS files using `postcss`. It supports a few options:

- `Command`: defaults to `npx postcss`, the command to run postcss. This must be a command name with all required arguments. Quotes are supported (pseudo-shell).
- `Filter`: defaults to `\.(s?css|sass)$`, a regular expression to match files that should be processed by postcss.
- `Loader`: defaults to `api.LoaderCSS`, you can pass a function which will be given the path of the file being processed so you can pick another loader instead, e.g. `api.LoaderLocalCSS` if dealing with CSS modules.

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

- By default, this plugin assumes you have `npx` and `postcss` installed and in your PATH.

### `cssmodules`

This plugin can be used to automatically generate JS mappings for your CSS modules. For example if doing `import styles from "./styles.module.css"`, it will generate a mapping of the CSS classes in that file to their generated names, and export it as the default export of the module. It assumes that your file will be loaded with `api.LoaderLocalCSS` and the mappings are built that way.

Options include:

- `Filter`: defaults to `\.module\.(s?css|sass)$`, a regular expression to match files that should be processed as CSS modules.

```go
package main

import (
  "os"

  "github.com/evanw/esbuild/pkg/api"
  "github.com/LouisBrunner/esbuild-plugins/pkg/cssmodules"
)

func main() {
  result := api.Build(api.BuildOptions{
    EntryPoints: []string{"app.js"},
    Bundle:      true,
    Outfile:     "out.js",
    Loaders: map[string]api.Loader{
      ".css": api.LoaderLocalCSS,
    },
    Plugins: []api.Plugin{cssmodules.NewPlugin(cssmodules.Options{
      Filter: `\.module\.css$`,
    })},
    Write: true,
  })

  if len(result.Errors) > 0 {
    os.Exit(1)
  }
}
```
