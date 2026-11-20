package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

const countdownPath = "/api/servers/alpha/power/countdown"

type fakeCountdownTimer struct {
	at       time.Time
	callback func()
	stopped  bool
}

func (t *fakeCountdownTimer) Stop() bool { wasActive := !t.stopped; t.stopped = true; return wasActive }

type fakeCountdownClock struct {
	now    time.Time
	timers []*fakeCountdownTimer
}

func (c *fakeCountdownClock) after(d time.Duration, callback func()) countdownTimer {
	t := &fakeCountdownTimer{at: c.now.Add(d), callback: callback}
	c.timers = append(c.timers, t)
	return t
}
func (c *fakeCountdownClock) advance(d time.Duration) {
	end := c.now.Add(d)
	for {
		var next *fakeCountdownTimer
		for _, timer := range c.timers {
			if !timer.stopped && !timer.at.After(end) && (next == nil || timer.at.Before(next.at)) {
				next = timer
			}
		}
		if next == nil {
			break
		}
		c.now = next.at
		next.stopped = true
		next.callback()
	}
	c.now = end
}
func (c *fakeCountdownClock) active() int {
	n := 0
	for _, timer := range c.timers {
		if !timer.stopped {
			n++
		}
	}
	return n
}

type fakeCountdownManager struct {
	online               bool
	commands             []string
	actions              []string
	commandErr, powerErr error
	onPower              func()
}

func (m *fakeCountdownManager) GetPowerRun(string) (runner.PowerRun, bool) {
	return runner.PowerRun{}, m.online
}
func (m *fakeCountdownManager) SendCommand(_ string, command string) error {
	if m.commandErr != nil {
		return m.commandErr
	}
	m.commands = append(m.commands, command)
	return nil
}
func (m *fakeCountdownManager) CountdownPower(srv *database.Server, action string, _ runner.PowerRun) error {
	if m.onPower != nil {
		m.onPower()
	}
	m.actions = append(m.actions, srv.ID+":"+action)
	return m.powerErr
}

func installCountdownClock(t *testing.T, f *reliabilityFixture) (*powerCountdownCoordinator, *fakeCountdownClock, *fakeCountdownManager) {
	t.Helper()
	c := f.router.servers.powerCountdowns()
	clock := &fakeCountdownClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	manager := &fakeCountdownManager{online: true}
	c.now, c.afterFunc, c.manager = func() time.Time { return clock.now }, clock.after, manager
	t.Cleanup(func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		for id := range c.pending {
			c.cancelLocked(id)
		}
		if clock.active() != 0 {
			t.Error("countdown timer leaked")
		}
	})
	return c, clock, manager
}

func countdownGrants(t *testing.T, f *reliabilityFixture, grants ...string) {
	t.Helper()
	subs, err := f.db.ListServerSubusers("alpha")
	if err != nil || len(subs) != 1 {
		t.Fatal(subs, err)
	}
	if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, grants); err != nil {
		t.Fatal(err)
	}
}

func TestPowerCountdownScheduleAndExecute(t *testing.T) {
	for _, action := range []string{"restart", "stop"} {
		t.Run(action, func(t *testing.T) {
			f := newReliabilityFixture(t)
			c, clock, manager := installCountdownClock(t, f)
			w := f.request("POST", countdownPath, fmt.Sprintf(`{"action":%q,"seconds":90,"message":"Maintenance"}`, action), "owner")
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			var body struct {
				Countdown *PowerCountdownState `json:"countdown"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Countdown == nil {
				t.Fatal(body, err)
			}
			if body.Countdown.Action != action || body.Countdown.Message != "Maintenance" || !body.Countdown.ExecuteAt.Equal(clock.now.Add(90*time.Second)) {
				t.Fatal(body.Countdown)
			}
			if w := f.request("GET", countdownPath, "", "subuser"); w.Code != 200 || !strings.Contains(w.Body.String(), "Maintenance") {
				t.Fatal(w.Code, w.Body)
			}
			manager.onPower = func() {
				if c.mu.TryLock() {
					c.mu.Unlock()
					t.Error("final power execution released cancellation mutex")
				}
			}
			clock.advance(89 * time.Second)
			if len(manager.actions) != 0 {
				t.Fatal("executed early")
			}
			want := []string{}
			for _, seconds := range []int{90, 60, 30, 10} {
				want = append(want, fmt.Sprintf("say Server will %s in %d seconds. Maintenance", action, seconds))
			}
			if !reflect.DeepEqual(manager.commands, want) {
				t.Fatal(manager.commands)
			}
			clock.advance(time.Second)
			if !reflect.DeepEqual(manager.actions, []string{"alpha:" + action}) || clock.active() != 0 {
				t.Fatal(manager.actions, clock.active())
			}
			if w := f.request("GET", countdownPath, "", "owner"); w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"countdown":null}` {
				t.Fatal(w.Code, w.Body)
			}
			clock.advance(time.Hour)
			if len(manager.actions) != 1 {
				t.Fatal("execution repeated")
			}
		})
	}
}

func TestPowerCountdownCancelDuplicatesAndDirectPower(t *testing.T) {
	f := newReliabilityFixture(t)
	c, clock, manager := installCountdownClock(t, f)
	body := `{"action":"restart","seconds":10}`
	var wg sync.WaitGroup
	codes := make(chan int, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- f.request("POST", countdownPath, body, "owner").Code }()
	}
	wg.Wait()
	close(codes)
	accepted := 0
	for code := range codes {
		if code == 200 {
			accepted++
		} else if code != 409 {
			t.Fatal(code)
		}
	}
	if accepted != 1 || len(manager.commands) != 1 || clock.active() != 1 {
		t.Fatal(accepted, manager.commands, clock.active())
	}
	stale := c.pending["alpha"].timer.(*fakeCountdownTimer)
	for _, action := range []string{"start", "restart", "stop", "kill"} {
		if w := f.request("POST", "/api/servers/alpha/power", fmt.Sprintf(`{"action":%q}`, action), "owner"); w.Code != 409 {
			t.Fatal(action, w.Code, w.Body)
		}
	}
	for i := 0; i < 2; i++ {
		if w := f.request("DELETE", countdownPath, "", "owner"); w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"countdown":null}` {
			t.Fatal(w.Code, w.Body)
		}
	}
	clock.advance(20 * time.Second)
	stale.callback() // simulate a callback queued before Timer.Stop
	if len(manager.actions) != 0 || clock.active() != 0 {
		t.Fatal(manager.actions, clock.active())
	}
	if w := f.request("POST", countdownPath, body, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	stale.callback()
	if c.pending["alpha"] == nil {
		t.Fatal("stale callback cancelled replacement")
	}
	clock.advance(10 * time.Second)
	if len(manager.actions) != 1 {
		t.Fatal(manager.actions)
	}
}

func TestPowerCountdownPermissionsAndValidation(t *testing.T) {
	f := newReliabilityFixture(t)
	_, clock, manager := installCountdownClock(t, f)
	body := `{"action":"stop","seconds":10}`
	for _, user := range []string{"", "stranger"} {
		want := 403
		if user == "" {
			want = 401
		}
		for _, method := range []string{"GET", "POST", "DELETE"} {
			if w := f.request(method, countdownPath, body, user); w.Code != want {
				t.Fatal(user, method, w.Code, w.Body)
			}
		}
	}
	for _, grants := range [][]string{{"console.view"}, {"server.power"}, {"console.command"}, {"files.view"}, {"server.power", "console.command"}} {
		countdownGrants(t, f, grants...)
		getWant, postWant, deleteWant := 403, 403, 403
		for _, grant := range grants {
			if grant == "console.view" || grant == "server.power" {
				getWant = 200
			}
			if grant == "server.power" {
				deleteWant = 200
			}
		}
		if len(grants) == 2 {
			postWant = 200
		}
		for _, check := range []struct {
			method string
			want   int
		}{{"GET", getWant}, {"POST", postWant}, {"DELETE", deleteWant}} {
			if w := f.request(check.method, countdownPath, body, "subuser"); w.Code != check.want {
				t.Fatal(grants, check, w.Code, w.Body)
			}
		}
	}
	invalid := []string{`{}`, `null`, `{"action":"start","seconds":10}`, `{"action":"stop","seconds":9}`, `{"action":"stop","seconds":3601}`, `{"action":"stop","seconds":10.5}`, `{"action":"stop","seconds":"10"}`, `{"action":"stop","seconds":null}`, `{"action":"stop","seconds":10,"unknown":true}`, body + ` {}`, body + ` junk`}
	invalid = append(invalid, `{"action":"stop","seconds":10,"message":null}`, `{"action":"stop","seconds":10,"message":123}`)
	for _, message := range []string{"a\nb", "a\rb", "a\tb", "a\x00b", "a\x7fb", "a\u0085b", strings.Repeat("界", 201)} {
		b, _ := json.Marshal(map[string]interface{}{"action": "stop", "seconds": 10, "message": message})
		invalid = append(invalid, string(b))
	}
	for _, input := range invalid {
		if w := f.request("POST", countdownPath, input, "owner"); w.Code != 400 {
			t.Fatal(input, w.Code, w.Body)
		}
	}
	if clock.active() != 0 {
		t.Fatal("invalid request scheduled timer")
	}
	for _, seconds := range []int{10, 3600} {
		b, _ := json.Marshal(map[string]interface{}{"action": "stop", "seconds": seconds, "message": strings.Repeat("界", 200)})
		if w := f.request("POST", countdownPath, string(b), "admin"); w.Code != 200 {
			t.Fatal(w.Code, w.Body)
		}
		f.request("DELETE", countdownPath, "", "admin")
	}
	manager.online = false
	if w := f.request("POST", countdownPath, body, "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body)
	}
	manager.online = true
	manager.commandErr = errors.New("stdin closed")
	if w := f.request("POST", countdownPath, body, "owner"); w.Code != 409 || clock.active() != 0 {
		t.Fatal(w.Code, w.Body)
	}
	if w := f.request("GET", "/api/servers/missing/power/countdown", "", "owner"); w.Code != 404 {
		t.Fatal(w.Code, w.Body)
	}
}

func TestPowerCountdownExpires(t *testing.T) {
	for _, reason := range []string{"stopped", "deleted", "permission", "user-deleted", "admin-demoted", "promoted-without-grants", "warning-failed", "power-failed", "http-deleted"} {
		t.Run(reason, func(t *testing.T) {
			f := newReliabilityFixture(t)
			c, clock, manager := installCountdownClock(t, f)
			countdownGrants(t, f, "server.power", "console.command")
			user := "subuser"
			if reason == "admin-demoted" {
				user = "admin"
			}
			if w := f.request("POST", countdownPath, `{"action":"restart","seconds":90}`, user); w.Code != 200 {
				t.Fatal(w.Code, w.Body)
			}
			var err error
			switch reason {
			case "stopped":
				manager.online = false
			case "deleted":
				err = f.db.DeleteServer("alpha")
			case "permission":
				countdownGrants(t, f, "server.power")
			case "user-deleted":
				err = f.db.DeleteUser(f.users[user].ID)
			case "admin-demoted":
				err = f.db.UpdateUserRole(f.users[user].ID, "user")
			case "promoted-without-grants":
				countdownGrants(t, f, "files.view")
				err = f.db.UpdateUserRole(f.users[user].ID, "admin")
			case "warning-failed":
				manager.commandErr = errors.New("stdin closed")
			case "power-failed":
				manager.powerErr = errors.New("restore gate busy")
			case "http-deleted":
				if w := f.request("DELETE", "/api/servers/alpha", "", "owner"); w.Code != 200 {
					t.Fatal(w.Code, w.Body)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			clock.advance(100 * time.Second)
			want := 0
			if reason == "power-failed" {
				want = 1
			}
			if len(manager.actions) != want || clock.active() != 0 || len(c.pending) != 0 {
				t.Fatal(manager.actions, clock.active(), c.pending)
			}
		})
	}
}

func TestPowerCountdownWorkerProxy(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	c, clock, manager := installCountdownClock(t, worker)
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	path := "/api/nodes/remote/proxy" + countdownPath
	body := `{"action":"restart","seconds":10}`
	if w := master.request("POST", path, body, "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body)
	}
	countdownGrants(t, worker, "server.power", "console.command")
	for _, user := range []string{"subuser", "admin"} {
		if w := master.request("POST", path, body, user); w.Code != 200 {
			t.Fatal(w.Code, w.Body)
		}
		if c.pending["alpha"].origin.UserID != master.users[user].ID {
			t.Fatal("proxy lost origin identity")
		}
		if w := master.request("GET", path, "", user); w.Code != 200 || !strings.Contains(w.Body.String(), "execute_at") {
			t.Fatal(w.Code, w.Body)
		}
		if w := master.request("POST", "/api/nodes/remote/proxy/api/servers/alpha/power", `{"action":"stop"}`, user); w.Code != 409 {
			t.Fatal(w.Code, w.Body)
		}
		if w := master.request("DELETE", path, "", user); w.Code != 200 {
			t.Fatal(w.Code, w.Body)
		}
	}
	if w := master.request("POST", path, body, "subuser"); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	clock.advance(10 * time.Second) // worker executes independently of HTTP/browser
	if !reflect.DeepEqual(manager.actions, []string{"alpha:restart"}) {
		t.Fatal(manager.actions)
	}
}
