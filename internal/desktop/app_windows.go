//go:build windows && !node

package desktop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/daemon"
	"github.com/Pekly/NeoPanel/web"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	winopts "github.com/wailsapp/wails/v2/pkg/options/windows"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// Run creates a Windows window, not a browser tab or a second daemon process.
func Run(dataDir string) error {
	release, acquired, err := AcquireInstance(InstanceID(dataDir))
	if err != nil || !acquired {
		return err
	}
	defer release()
	cfg, err := LoadConfig(dataDir)
	if err != nil {
		return err
	}
	logPath := filepath.Join(dataDir, "desktop.log")
	if stat, err := os.Stat(logPath); err == nil && stat.Size() > 4<<20 {
		_ = os.Remove(logPath + ".1")
		_ = os.Rename(logPath, logPath+".1")
	}
	output, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer output.Close()
	previous := log.Writer()
	log.SetOutput(output)
	defer log.SetOutput(previous)
	log.Printf("Starting %s %s; data directory: %s", Title, Version, dataDir)
	service, err := startBackend(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := shutdownBackend(service); err != nil {
			log.Printf("Backend shutdown: %v", err)
		}
	}()
	app := &desktopApp{service: service, dataDir: dataDir, state: LoadWindowState(dataDir)}
	state := options.Normal
	if app.state.Maximized {
		state = options.Maximised
	}
	return wails.Run(&options.App{
		Title: Title, Width: app.state.Width, Height: app.state.Height,
		MinWidth: 900, MinHeight: 620, WindowStartState: state,
		BackgroundColour: options.NewRGB(17, 21, 28),
		AssetServer:      &assetserver.Options{Handler: http.HandlerFunc(loadingPage)},
		Menu:             app.menu(), OnStartup: app.startup, OnDomReady: app.domReady,
		OnBeforeClose: app.beforeClose,
		// No Bind and no extra allowed origins: the panel has no privileged Go bridge.
		Windows: &winopts.Options{
			Theme: winopts.Dark, WindowClassName: InstanceID(dataDir),
			WebviewUserDataPath:  filepath.Join(dataDir, "webview"),
			IsZoomControlEnabled: true,
			DLLSearchPaths:       winopts.DLLSearchSystem32,
		},
	})
}

// Bound HTTP draining, not server cleanup. Service.Shutdown still joins managed
// process cleanup after the deadline closes stalled network connections.
func shutdownBackend(service interface{ Shutdown(context.Context) error }) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return service.Shutdown(ctx)
}

func startBackend(cfg *config.Config) (*daemon.Service, error) {
	service, err := daemon.Start(cfg, web.GetFS())
	if err != nil && cfg.Server.Port != 0 && errors.Is(err, windows.WSAEADDRINUSE) {
		log.Printf("Remembered desktop port is occupied; choosing a free loopback port")
		cfg.Server.Port = 0
		service, err = daemon.Start(cfg, web.GetFS())
	}
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(service.URL())
	if err != nil {
		_ = shutdownBackend(service)
		return nil, err
	}
	_, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		_ = shutdownBackend(service)
		return nil, err
	}
	cfg.Server.Port, _ = strconv.Atoi(port)
	// A stable origin preserves WebView localStorage and sign-in across normal restarts.
	if err := config.SaveConfig(cfg.FilePath, cfg); err != nil {
		_ = shutdownBackend(service)
		return nil, fmt.Errorf("save desktop listener: %w", err)
	}
	log.Printf("Desktop backend ready at %s (loopback only)", service.URL())
	return service, nil
}

type desktopApp struct {
	service    *daemon.Service
	dataDir    string
	state      WindowState
	mu         sync.Mutex
	ctx        context.Context
	closing    bool
	closed     bool
	prompt     bool
	navigation sync.Once
}

func (a *desktopApp) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()
	go func() {
		<-a.service.Done()
		if err := a.service.Err(); err != nil {
			log.Printf("Desktop backend stopped unexpectedly: %v", err)
			_, _ = wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{Type: wruntime.ErrorDialog, Title: Title, Message: "The local backend stopped unexpectedly. See desktop.log in the app data folder. NeoPanel will close safely."})
			wruntime.Quit(ctx)
		}
	}()
}

func (a *desktopApp) domReady(ctx context.Context) {
	// Navigate the native WebView to the real loopback origin. This retains ordinary
	// WebSocket, streaming upload/download, authentication and browser security semantics.
	a.navigation.Do(func() { a.home(ctx) })
}

func (a *desktopApp) home(ctx context.Context) {
	target, _ := json.Marshal(a.service.URL())
	wruntime.WindowExecJS(ctx, "window.location.replace("+string(target)+");")
}

func (a *desktopApp) beforeClose(ctx context.Context) bool {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return false
	}
	if a.closing || a.prompt {
		a.mu.Unlock()
		return true
	}
	a.prompt = true
	a.mu.Unlock()
	if count := a.service.ActiveServerCount(); count > 0 {
		choice, err := wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{
			Type: wruntime.QuestionDialog, Title: "Stop local servers and quit?", DefaultButton: "No",
			Message: fmt.Sprintf("%d local server(s) are active.\n\nYes: stop and save local servers, then quit NeoPanel.\nNo: minimize the window and keep them running.\n\nRemote worker nodes are not stopped.", count),
		})
		if err != nil || choice != "Yes" {
			a.mu.Lock()
			a.prompt = false
			a.mu.Unlock()
			if err == nil {
				wruntime.WindowMinimise(ctx)
			}
			return true
		}
	}
	a.mu.Lock()
	a.prompt = false
	a.closing = true
	a.mu.Unlock()
	if !wruntime.WindowIsMinimised(ctx) {
		a.state.Maximized = wruntime.WindowIsMaximised(ctx)
		if !a.state.Maximized {
			a.state.Width, a.state.Height = wruntime.WindowGetSize(ctx)
		}
		if err := SaveWindowState(a.dataDir, a.state); err != nil {
			log.Printf("Window preferences: %v", err)
		}
	}
	wruntime.WindowSetTitle(ctx, Title+" — Saving servers and shutting down…")
	// Keep the native message loop responsive while the existing runner performs safe stop.
	go func() {
		err := shutdownBackend(a.service)
		if err != nil {
			log.Printf("Desktop shutdown: %v", err)
		}
		a.mu.Lock()
		a.closed = true
		a.mu.Unlock()
		wruntime.Quit(ctx)
	}()
	return true
}

func (a *desktopApp) action(fn func(context.Context)) func(*menu.CallbackData) {
	return func(_ *menu.CallbackData) {
		a.mu.Lock()
		ctx, closing := a.ctx, a.closing
		a.mu.Unlock()
		if ctx != nil && !closing {
			fn(ctx)
		}
	}
}

func (a *desktopApp) menu() *menu.Menu {
	root := menu.NewMenu()
	file := root.AddSubmenu("NeoPanel")
	file.AddText("Dashboard", nil, a.action(a.home))
	file.AddText("Open app data folder", nil, a.action(func(ctx context.Context) {
		if err := openLocalPath(a.dataDir); err != nil {
			log.Printf("Open data folder: %v", err)
		}
	}))
	file.AddSeparator()
	file.AddText("Minimize — keep servers running", nil, a.action(wruntime.WindowMinimise))
	file.AddText("Quit safely", nil, a.action(wruntime.Quit))
	view := root.AddSubmenu("View")
	view.AddText("Reload interface", nil, a.action(func(ctx context.Context) { wruntime.WindowExecJS(ctx, "window.location.reload();") }))
	view.AddText("Reset window size", nil, a.action(func(ctx context.Context) {
		wruntime.WindowUnmaximise(ctx)
		wruntime.WindowSetSize(ctx, 1280, 820)
		wruntime.WindowCenter(ctx)
	}))
	help := root.AddSubmenu("Help")
	help.AddText("About NeoPanel Desktop", nil, a.action(func(ctx context.Context) {
		_, _ = wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{Type: wruntime.InfoDialog, Title: Title, Message: "NeoPanel Desktop " + Version + "\n\nWindows head panel · Local hosting · Web-less workers\n\nMinimize to keep local servers running. Quit safely to save and stop them. Windows local workloads are trusted-only, without filesystem or network isolation.\n\nData: " + a.dataDir})
	}))
	return root
}

func loadingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, `<!doctype html><html><head><meta charset="utf-8"><title>NeoPanel Desktop</title><style>body{margin:0;min-height:100vh;display:grid;place-content:center;background:#11151c;color:#eef2f7;font:15px Segoe UI,sans-serif;text-align:center}h1{font-size:30px}p{color:#98a6bb}.bar{height:4px;background:#25324a;overflow:hidden;border-radius:4px}.bar:after{content:"";display:block;width:40%;height:100%;background:#387aff;animation:load 1.5s infinite}@keyframes load{from{transform:translateX(-100%)}to{transform:translateX(350%)}}@media(prefers-reduced-motion:reduce){.bar:after{animation:none}}</style></head><body><main><h1>NeoPanel Desktop</h1><p>Opening your local panel…</p><div class="bar" role="progressbar" aria-label="Opening NeoPanel"></div></main></body></html>`)
}
