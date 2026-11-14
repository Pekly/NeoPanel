package runner

import (
	"fmt"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

// javaVMArgs accepts self-contained VM options, not Java launcher arguments.
// Argument/VM-option files could be tenant-writable, and launcher operands can
// consume the managed heap flags or stop option parsing before they are applied.
// Keep explicit heap sizes last so legacy -Xmx/-XX:MaxHeapSize cannot override
// the panel's configured budget. This is not a sandbox: cgroups bound native
// allocations and child processes as well as the Java heap.
func javaVMArgs(srv database.Server) ([]string, error) {
	args := strings.Fields(srv.JavaArgs)
	for _, arg := range args {
		allowed := strings.HasPrefix(arg, "-D") || strings.HasPrefix(arg, "-X")
		switch arg {
		case "-server", "-client", "-ea", "-enableassertions", "-da", "-disableassertions", "-esa", "-enablesystemassertions", "-dsa", "-disablesystemassertions", "-verbose", "-noverify", "-verify", "--enable-preview":
			allowed = true
		}
		for _, prefix := range []string{"-ea:", "-enableassertions:", "-da:", "-disableassertions:", "-verbose:", "-javaagent:", "-agentlib:", "-agentpath:", "--add-opens=", "--add-exports=", "--add-reads=", "--add-modules=", "--limit-modules=", "--illegal-access=", "--enable-native-access="} {
			allowed = allowed || strings.HasPrefix(arg, prefix)
		}
		if strings.HasPrefix(arg, "-XX:Flags=") || strings.HasPrefix(arg, "-XX:VMOptionsFile=") {
			allowed = false
		}
		if !allowed {
			return nil, fmt.Errorf("unsupported custom JVM argument %q: use self-contained VM flags, not launcher arguments or option files", arg)
		}
	}
	if srv.MemoryMinMB > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", srv.MemoryMinMB))
	}
	if srv.MemoryMaxMB > 0 {
		args = append(args, fmt.Sprintf("-Xmx%dM", srv.MemoryMaxMB))
	}
	return args, nil
}
