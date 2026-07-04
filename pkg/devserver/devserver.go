package devserver

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/evanw/esbuild/pkg/api"
)

type Logger interface {
	Printf(format string, v ...any)
	Fatal(v ...any)
}

const defaultPort = 8000

// Options configures a build or dev server run.
type Options struct {
	// Logger is used to log build/serve output.
	Logger Logger
	// Build holds the project-specific configuration, some settings are overridden.
	Build api.BuildOptions
	// Output is the build output directory.
	Output string
	// PublicDir, if set, is copied into Output before building, it's also served by the dev server.
	PublicDir string
	// Port the dev server binds to (127.0.0.1), defaults to 8000.
	Port int
	// OpenBrowser opens Output's URL once the dev server is listening.
	OpenBrowser bool
}

func (me *Options) getLogger() Logger {
	if me.Logger == nil {
		me.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return me.Logger
}

// BuildOptions returns a copy of build tuned for either dev or prod.
func BuildOptions(build api.BuildOptions, isDev bool) api.BuildOptions {
	define := map[string]string{}
	maps.Copy(define, build.Define)
	define["process.env.WATCH_MODE"] = fmt.Sprintf("%t", isDev)

	if isDev {
		define["process.env.ENV"] = `"development"`
		build.Sourcemap = api.SourceMapInline
		build.Write = false
	} else {
		define["process.env.ENV"] = `"production"`
		build.MinifyWhitespace = true
		build.MinifyIdentifiers = true
		build.MinifySyntax = true
		build.Write = true
	}
	build.Define = define

	return build
}

// Build runs a one-off production build.
func Build(opts Options) error {
	opts.getLogger().Printf("Building")

	if opts.Output != "" {
		opts.getLogger().Printf("> Removing %s", opts.Output)
		err := os.RemoveAll(opts.Output)
		if err != nil {
			return err
		}

		if opts.PublicDir != "" {
			opts.getLogger().Printf("> Copying %s to %s", opts.PublicDir, opts.Output)
			err = copyDir(opts.PublicDir, opts.Output)
			if err != nil {
				return err
			}
		}
	} else {
		opts.Output = opts.PublicDir
	}

	build := BuildOptions(opts.Build, false)
	build.Outdir = opts.Output

	opts.getLogger().Printf("> Building...")
	result := api.Build(build)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			opts.getLogger().Printf("%v: %s", e.Location, e.Text)
		}
		return fmt.Errorf("build failed")
	}

	opts.getLogger().Printf("> Done")
	return nil
}

// Start builds and watches, serving on 127.0.0.1 and blocking until ctx is cancelled.
func Start(ctx context.Context, opts Options) error {
	build := BuildOptions(opts.Build, true)
	build.Outdir = opts.PublicDir

	buildCtx, ctxErr := api.Context(build)
	if ctxErr != nil {
		return ctxErr
	}
	defer buildCtx.Dispose()

	opts.getLogger().Printf("Watching")
	watchErr := buildCtx.Watch(api.WatchOptions{})
	if watchErr != nil {
		return watchErr
	}

	port := opts.Port
	if port == 0 {
		port = defaultPort
	}

	serveResult, serveErr := buildCtx.Serve(api.ServeOptions{
		Host:     "127.0.0.1",
		Port:     port,
		Servedir: opts.PublicDir,
	})
	if serveErr != nil {
		return serveErr
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", serveResult.Port)
	opts.getLogger().Printf("Listening on %s", addr)

	if opts.OpenBrowser {
		openErr := exec.Command("open", addr).Start()
		if openErr != nil {
			opts.getLogger().Printf("could not open browser: %v", openErr)
		}
	}

	<-ctx.Done()
	return nil
}

// Run parses flags for the CLI and runs either a one-off build or a dev server.
func Run(defaults Options) {
	dev := false
	port := defaults.Port
	openBrowser := defaults.OpenBrowser

	flag.BoolVar(&dev, "dev", dev, "run the dev server instead of a one-off build")
	flag.IntVar(&port, "port", port, "dev server port")
	flag.BoolVar(&openBrowser, "open", openBrowser, "open the browser once the dev server is listening")
	flag.Parse()

	opts := defaults
	opts.Port = port
	opts.OpenBrowser = openBrowser

	var err error
	if dev {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		err = Start(ctx, opts)
	} else {
		err = Build(opts)
	}
	if err != nil {
		opts.getLogger().Fatal(err)
	}
}

func copyDir(src, dst string) error {
	return fs.WalkDir(os.DirFS(src), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		srcPath := filepath.Join(src, path)
		dstPath := filepath.Join(dst, path)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
