# main.go changes

In `main()`, after the PR-URL detection block and before `resolveTarget`, add the
multi-repo branch. Everything else in `main()` stays as is for single-repo runs.

```go
	// px0 dirA dirB ... opens a multi-repository workspace (#162).
	if !isPR && flag.NArg() > 1 {
		var roots []string
		for _, a := range flag.Args() {
			r, f, _, err := resolveTarget(a)
			if err != nil {
				fatal(err)
			}
			if f != "" {
				fatal(fmt.Errorf("%s: pass directories when opening several repositories", a))
			}
			roots = append(roots, r)
		}
		bp := "/"
		if *basePathFlag != "" {
			bp = cleanBasePath(*basePathFlag)
		}
		hub, err := newWorkspaceHub(roots, bp, !*noLSP, *agentCmd, *noAgent)
		if err != nil {
			fatal(err)
		}
		ln, addr, err := listen(*host, *port)
		if err != nil {
			fatal(err)
		}
		url := viewerURL(addr, "", 0, bp)
		uiHeading("px0 "+version, nil, os.Stdout)
		for _, r := range hub.repos {
			uiKV("workspace", r.Root, 11, os.Stdout)
		}
		uiKV("url", uiAccent(url, os.Stdout), 11, os.Stdout)
		uiHint("ctrl-c to stop", os.Stdout)
		if !*noOpen {
			go openBrowser(url)
		}
		go hub.Build()
		srv := &http.Server{Handler: hub}
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-stop
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
		}()
		err = srv.Serve(ln)
		hub.Close()
		if err != nil && err != http.ErrServerClosed {
			fatal(err)
		}
		return
	}
```

Also update the usage string:

```go
fmt.Fprintf(os.Stderr, "px0 %s - a code navigator\n\nusage:\n  px0 [flags] [file or directory]\n  px0 [flags] <dir> <dir> ...\n  px0 [flags] <pr-url>\n\nflags:\n", version)
```

## web/src/main.js

Import and call the switcher during `boot()`:

```js
import { initWorkspace } from './workspace.js';
// inside boot(), after meta is loaded:
initWorkspace();
```

Then rebuild the bundle: `node ./scripts/build-web.js`.
