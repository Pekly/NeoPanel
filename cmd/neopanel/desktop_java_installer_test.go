package main

import (
	"os"
	"strings"
	"testing"
)

func TestDesktopInstallerProvisionsBeforeOptionalLaunch(t *testing.T) {
	raw, err := os.ReadFile("../../build/windows-desktop/installer.iss")
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, required := range []string{
		`Source: "{#SourceDir}\NeoPanel-Runtime.exe"; DestDir: "{app}"`,
		`Check: JavaProvisioned`,
		`Result := RuntimesReady;`,
		`if CurStep = ssPostInstall then`,
		`RuntimesReady := False;`,
		`Exec(ExpandConstant('{app}\NeoPanel-Runtime.exe'), 'java install-all'`,
		`SW_HIDE, ewWaitUntilTerminated, ResultCode`,
		`if ResultCode <> 0 then`,
		`RaiseException(`,
		`RuntimesReady := True;`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("missing installer guard: %s", required)
		}
	}
	if strings.Index(text, "RuntimesReady := True;") < strings.Index(text, "if ResultCode <> 0 then") {
		t.Fatal("launch enabled before runtime verification")
	}
	raw, err = os.ReadFile("../../build-windows-desktop.ps1")
	if err != nil {
		t.Fatal(err)
	}
	text = string(raw)
	for _, required := range []string{`'NeoPanel-Runtime.exe'`, `-tags node '-ldflags=-s -w' -o $runtimeHelper ./cmd/neopanel`, `if ($LASTEXITCODE -ne 0) { throw 'Java provisioning helper build failed.' }`, `$files = @($exe, $runtimeHelper, $readme)`} {
		if !strings.Contains(text, required) {
			t.Fatalf("helper absent from build/package/checksum inputs: %s", required)
		}
	}
}
