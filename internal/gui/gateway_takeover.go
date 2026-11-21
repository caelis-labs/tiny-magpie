package gui

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/proc"
)

// gatewayFixJSON reports a restart of this process's own gateway.
type gatewayFixJSON struct {
	OK      bool   `json:"ok"`
	Reason  string `json:"reason,omitempty"`
	PID     int    `json:"pid,omitempty"`
	Path    string `json:"path,omitempty"`
	Port    string `json:"port,omitempty"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"` // the system's own words
	// Name, Tailscale and Alt: for "taken", the program that has the port
	// and whether it is Tailscale's serve or funnel, with another port to
	// give that (gateway.PortHeld); left out when the holder isn't read
	Name      string `json:"name,omitempty"`
	Tailscale bool   `json:"tailscale,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

func gatewayFixRoutes(mux *http.ServeMux) {

	mux.HandleFunc("POST /api/gateway/restart", func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, restartGateway())
	})
}

var startWait = 5 * time.Second

// restartGateway only stops a gateway owned by this process.
func restartGateway() gatewayFixJSON {
	gatewayMu.Lock()
	defer gatewayMu.Unlock()
	if gw := served.Load(); gw != nil {
		if servedRun.stop != nil {
			servedRun.stop()
			<-servedRun.done
		}
		served.CompareAndSwap(gw, nil)
	} else if o := gateway.ServedBy(); o.Running {
		return gatewayFixJSON{Reason: "other", Version: o.Version, Port: gateway.Port()}
	}
	return startHere()
}

// startHere serves the gateway here and waits until it answers, or says
// what has the port.
func startHere() gatewayFixJSON {
	gw := serveGatewayLocked()
	for deadline := time.Now().Add(startWait); gw != nil && served.Load() == gw; time.Sleep(50 * time.Millisecond) {
		if gateway.Running() {
			return gatewayFixJSON{OK: true}
		}
		if time.Now().After(deadline) {
			break
		}
	}
	// it didn't take the port: who has it
	out := gatewayFixJSON{Reason: "taken", Port: gateway.Port()}
	if o := gateway.ServedBy(); o.Running {
		out.Version = o.Version
	}
	port, _ := strconv.Atoi(out.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if pids, err := proc.ListeningOn(ctx, port); err == nil {
		for _, pid := range pids {
			if pid == os.Getpid() {
				continue
			}
			out.PID = pid
			out.Path, _ = proc.Executable(pid)
			if proc.RunsMagpie(pid, out.Path) {
				// a magpie took it back as soon as the other quit
				out.Reason = "respawned"
			} else if proc.IsForwarder(out.Path) {
				out.Reason = "container"
			}
			break
		}
	}
	if out.Reason == "taken" {
		if h := gateway.LastHeld(); h != nil || out.PID == 0 {
			if h == nil {
				if n, ok := gateway.NoteHeld(); ok {
					h = &n
				}
			}
			if h != nil {
				out.Name, out.Tailscale, out.Alt = h.Name, h.Tailscale, h.Alt
				if out.PID == 0 {
					out.PID, out.Path = h.PID, h.Path
				}
			}
		}
	}
	return out
}
