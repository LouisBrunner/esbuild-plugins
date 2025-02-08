package postcss

import (
	"fmt"
	"os/exec"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/google/shlex"
)

type Options struct {
	Filter  string
	Command string
}

func Must(plugin *api.Plugin, err error) api.Plugin {
	if err != nil {
		panic(err)
	}
	return *plugin
}

func NewPlugin(opts Options) (*api.Plugin, error) {
	if opts.Filter == "" {
		opts.Filter = `\.(s?css|sass)$`
	}
	if opts.Command == "" {
		opts.Command = "npx postcss"
	}

	cmdParts, err := shlex.Split(opts.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	return &api.Plugin{
		Name: "postcss",
		Setup: func(build api.PluginBuild) {
			build.OnLoad(api.OnLoadOptions{Filter: opts.Filter, Namespace: "file"},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					cmd := exec.Command(
						cmdParts[0],
						append(cmdParts[1:], args.Path)...,
					)
					res, err := cmd.CombinedOutput()
					if err != nil {
						return api.OnLoadResult{
							Errors: []api.Message{
								{Text: string(res)},
							},
						}, nil
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
	}, nil
}
