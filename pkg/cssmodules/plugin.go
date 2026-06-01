package cssmodules

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type Options struct {
	Filter string
	Logger *log.Logger
}

func Must(plugin *api.Plugin, err error) api.Plugin {
	if err != nil {
		panic(err)
	}
	return *plugin
}

func NewPlugin(opts Options) api.Plugin {
	if opts.Filter == "" {
		opts.Filter = `\.module\.(s?css|sass)$`
	}
	if opts.Logger == nil {
		opts.Logger = log.New(io.Discard, "esbuild-cssmodules", log.LstdFlags)
	}

	return api.Plugin{
		Name: "css-modules",
		Setup: func(pb api.PluginBuild) {
			pb.OnResolve(api.OnResolveOptions{Filter: opts.Filter}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				abs := filepath.Join(args.ResolveDir, args.Path)
				return api.OnResolveResult{Path: abs, Namespace: "css-module"}, nil
			})

			pb.OnLoad(api.OnLoadOptions{Filter: opts.Filter, Namespace: "css-module"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				b, err := os.ReadFile(args.Path)
				if err != nil {
					return api.OnLoadResult{Errors: []api.Message{{Text: fmt.Sprintf("failed to read CSS module %s: %v", args.Path, err)}}}, nil
				}

				filename := filepath.Base(args.Path)
				for {
					ext := filepath.Ext(filename)
					if ext == "" {
						break
					}
					filename = strings.TrimSuffix(filename, ext)
				}
				localPrefix := strings.ReplaceAll(filename, "-", "_") + "_"

				re := regexp.MustCompile(`\.([A-Za-z0-9_-]+)`)
				matches := re.FindAllStringSubmatch(string(b), -1)
				mappings := make(map[string]string, len(matches))
				for _, m := range matches {
					if len(m) < 2 {
						continue
					}
					cssClass := m[1]
					mappings[cssClass] = localPrefix + cssClass
				}
				mappingsJSON, err := json.Marshal(mappings)
				if err != nil {
					return api.OnLoadResult{Errors: []api.Message{{Text: fmt.Sprintf("failed to marshal CSS module mappings for %s: %v", args.Path, err)}}}, nil
				}

				opts.Logger.Printf("Loaded CSS module %s with prefix %q and classes: %v\n", args.Path, localPrefix, slices.Collect(maps.Keys(mappings)))
				wrapper := fmt.Sprintf("import %q;\nexport default %s;", args.Path+"?raw", mappingsJSON)
				return api.OnLoadResult{Contents: &wrapper, Loader: api.LoaderJS}, nil
			})

			pb.OnResolve(api.OnResolveOptions{Filter: `.*\.module\.css\?raw$`}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				orig := strings.TrimSuffix(args.Path, "?raw")
				return api.OnResolveResult{Path: orig, Namespace: "file"}, nil
			})
		},
	}
}
