package runner

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestJavaRequiredCatalog(t *testing.T) {
	want := []string{"25", "21", "17", "16", "11", "8"}
	if !reflect.DeepEqual(SupportedJavaMajors(), want) {
		t.Fatal(SupportedJavaMajors())
	}
	copy := SupportedJavaMajors()
	copy[0] = "7"
	for _, ver := range want {
		if !javaSupported[ver] || parseJavaTag(ver) != ver {
			t.Fatalf("unsupported required family %s", ver)
		}
	}
	for _, ver := range []string{"7", "19", "99", "auto"} {
		if javaSupported[ver] {
			t.Fatalf("unexpected managed family %s", ver)
		}
	}
	for _, alias := range []string{"", "java", "default"} {
		if parseJavaTag(alias) != "default" {
			t.Fatalf("changed system semantics for %q", alias)
		}
	}
}

func TestJavaRequiredProvisionMissingAndIdempotence(t *testing.T) {
	installed := map[string]string{"25": "25", "17": "17", "16": "11"} // wrong-major directory is NOT installed
	var calls []string
	resolve := func(ver string) (string, string) { return ver, "" }
	detect := func(bin string) string { return installed[bin] }
	install := func(ver string) (JavaRuntimeInfo, error) {
		calls = append(calls, ver)
		installed[ver] = ver
		return JavaRuntimeInfo{Version: ver, Path: ver, Installed: true}, nil
	}
	for i := 0; i < 2; i++ {
		if err := installRequiredJavaRuntimes(resolve, detect, install); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(calls, []string{"21", "16", "11", "8"}) {
		t.Fatalf("install order/reuse: %v", calls)
	}
}

func TestJavaRequiredPartialFailureAndWrongMajor(t *testing.T) {
	for _, scenario := range []string{"unavailable", "wrong-major", "unparseable"} {
		t.Run(scenario, func(t *testing.T) {
			var calls []string
			detect := func(bin string) string {
				if bin == "16" && scenario == "wrong-major" {
					return "17"
				}
				if bin == "16" && scenario == "unparseable" {
					return ""
				}
				return bin
			}
			install := func(ver string) (JavaRuntimeInfo, error) {
				calls = append(calls, ver)
				if ver == "16" && scenario == "unavailable" {
					return JavaRuntimeInfo{}, fmt.Errorf("no build for architecture")
				}
				return JavaRuntimeInfo{Version: ver, Path: ver, Installed: true}, nil
			}
			err := installRequiredJavaRuntimes(func(string) (string, string) { return "", "" }, detect, install)
			if err == nil || !strings.Contains(err.Error(), "Java 16") {
				t.Fatalf("lost failed family: %v", err)
			}
			if !reflect.DeepEqual(calls, []string{"25", "21", "17", "16"}) {
				t.Fatalf("did not stop at failure: %v", calls)
			}
		})
	}
}
