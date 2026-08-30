package sandbox_test

import (
	"bytes"
	"strings"
	"testing"

	"athenaeum/internal/sandbox"
	"athenaeum/internal/term"
)

func TestStatusSummary(t *testing.T) {
	st := sandbox.Status{
		Mode: sandbox.ModeTry,
		Landlock: sandbox.Component{
			State:  sandbox.StateApplied,
			Detail: "request=V9 effective=V6 best-effort kernel_abi=6",
		},
		Seccomp: sandbox.Component{
			State:  sandbox.StateApplied,
			Detail: "denylist=18 no_new_privs arch=amd64",
		},
	}
	sum := st.Summary()
	if !strings.Contains(sum, "mode=try") {
		t.Fatalf("summary missing mode: %s", sum)
	}
	if !strings.Contains(sum, "landlock=applied") {
		t.Fatalf("summary missing landlock: %s", sum)
	}
	if !strings.Contains(sum, "seccomp=applied") {
		t.Fatalf("summary missing seccomp: %s", sum)
	}
}

func TestStatusLogArgs(t *testing.T) {
	st := sandbox.Status{
		Mode:     sandbox.ModeOff,
		Landlock: sandbox.Component{State: sandbox.StateOff},
		Seccomp:  sandbox.Component{State: sandbox.StateOff},
	}
	args := st.LogArgs()
	if len(args) != 6 {
		t.Fatalf("want 6 args got %d", len(args))
	}
	if args[0] != "sandbox" || args[1] != "off" {
		t.Fatalf("sandbox args: %#v", args[:2])
	}
	if args[2] != "landlock" || !strings.HasPrefix(args[3].(string), "off") {
		t.Fatalf("landlock args: %#v", args[2:4])
	}
	if args[4] != "seccomp" || !strings.HasPrefix(args[5].(string), "off") {
		t.Fatalf("seccomp args: %#v", args[4:6])
	}
}

func TestStatusPrintColored(t *testing.T) {
	term.Apply(term.ModeAlways)
	defer term.Apply(term.ModeNever)

	st := sandbox.Status{
		Mode: sandbox.ModeStrict,
		Landlock: sandbox.Component{
			State:  sandbox.StateApplied,
			Detail: "request=V9 effective=V9 strict kernel_abi=9",
		},
		Seccomp: sandbox.Component{
			State:  sandbox.StateSkipped,
			Detail: "denylist=18",
			Err:    "permission denied",
		},
	}
	var buf bytes.Buffer
	st.Print(&buf)
	out := buf.String()
	if !strings.Contains(out, "sandbox") {
		t.Fatalf("missing label: %q", out)
	}
	if !strings.Contains(out, "landlock=") {
		t.Fatalf("missing landlock: %q", out)
	}
	if !strings.Contains(out, "seccomp=") {
		t.Fatalf("missing seccomp: %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI color codes: %q", out)
	}
}

func TestComponentReason(t *testing.T) {
	off := sandbox.Component{State: sandbox.StateOff}
	if got := off.Reason(); got != "sandbox mode is off" {
		t.Fatalf("off reason=%q", got)
	}
	skipped := sandbox.Component{
		State:  sandbox.StateSkipped,
		Detail: "denylist=18",
		Err:    "operation not permitted",
	}
	if got := skipped.Reason(); !strings.Contains(got, "denylist") || !strings.Contains(got, "operation not permitted") {
		t.Fatalf("skipped reason=%q", got)
	}
	pub := sandbox.Status{
		Mode:     sandbox.ModeTry,
		Landlock: skipped,
		Seccomp:  sandbox.Component{State: sandbox.StateApplied, Detail: "ok"},
	}.Public()
	ll, _ := pub["landlock"].(map[string]any)
	if ll["reason"] == "" || ll["state"] != "skipped" {
		t.Fatalf("public landlock=%#v", ll)
	}
}

func TestApplyOff(t *testing.T) {
	st, err := sandbox.Apply(sandbox.Config{Mode: sandbox.ModeOff}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode != sandbox.ModeOff {
		t.Fatalf("mode=%v", st.Mode)
	}
	if st.Landlock.State != sandbox.StateOff || st.Seccomp.State != sandbox.StateOff {
		t.Fatalf("want off components: %+v", st)
	}
}

func TestStatusAnnounceAndPrintModes(t *testing.T) {
	term.Apply(term.ModeAlways)
	defer term.Apply(term.ModeNever)

	st := sandbox.Status{
		Mode:     sandbox.ModeTry,
		Landlock: sandbox.Component{State: sandbox.StateDisabled, Detail: "toggle off"},
		Seccomp:  sandbox.Component{State: sandbox.StateUnsupported, Detail: "platform"},
	}
	st.Announce()

	var buf bytes.Buffer
	st.Print(&buf)
	if !strings.Contains(buf.String(), "mode=") {
		t.Fatalf("print: %q", buf.String())
	}

	st.Mode = sandbox.ModeOff
	st.Landlock = sandbox.Component{State: sandbox.StateOff}
	st.Seccomp = sandbox.Component{State: sandbox.StateOff, Err: "n/a"}
	buf.Reset()
	st.Print(&buf)
	if !strings.Contains(buf.String(), "landlock=") {
		t.Fatalf("off print: %q", buf.String())
	}
}
