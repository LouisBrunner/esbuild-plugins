package postcss

import (
	"os/exec"

	"github.com/evanw/esbuild/pkg/api"
)

var Plugin = api.Plugin{
	Name: "postcss",
	Setup: func(build api.PluginBuild) {
		build.OnLoad(api.OnLoadOptions{Filter: `\.css$`, Namespace: "file"},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				cmd := exec.Command(
					"npx",
					"postcss",
					args.Path,
				)
				res, err := cmd.CombinedOutput()
				if err != nil {
					return api.OnLoadResult{
						Errors: []api.Message{
							{Text: string(res)},
						},
					}, err
				}
				contents := string(res)
				return api.OnLoadResult{
					Contents: &contents,
					Loader:   api.LoaderCSS,
					WatchFiles: []string{
						args.Path,
					},
				}, nil
			})
	},
}
