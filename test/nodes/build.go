//go:build linux

package nodes

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/onsi/gomega/gexec"
	"riasc.eu/wice/test"
)

var (
	testBinaryPath string
)

func BuildTestBinary(name string) (string, []any, error) {
	var err error

	flags, profile := getProfileFlags()

	// Set some agent specific paths for the profile
	if _, ok := flags["outputdir"]; !ok {
		flags["outputdir"] = "."
	}

	for _, prof := range []string{"blockprofile", "coverprofile", "cpuprofile", "memprofile", "mutexprofile", "trace"} {
		if path, ok := flags[prof]; ok {
			fn := filepath.Base(path)

			flags[prof] = fmt.Sprintf("%s.%s", name, fn)
		}
	}

	runArgs := []any{}
	for k, v := range flags {
		runArg := fmt.Sprintf("-test.%s=%v", k, v)

		runArgs = append(runArgs, runArg)
	}

	if testBinaryPath == "" {
		buildArgs := []string{}

		// Pass-through -race option from Ginkgo to wice binary
		if test.IsRace {
			buildArgs = append(buildArgs, "-race")
		}

		// Pass-through profiling flags to wice binary
		if profile {
			buildArgs = append(buildArgs, "-tags", "test")

			for k, v := range flags {
				buildArg := fmt.Sprintf("-%s=%v", k, v)
				buildArgs = append(buildArgs, buildArg)
			}

			// We compile a dummy go test binary here which just
			// invokes main(), but is instrumented for profiling.
			testBinaryPath, err = gexec.CompileTest("..", buildArgs...)
		} else {
			testBinaryPath, err = gexec.Build("..", buildArgs...)
		}
	}

	return testBinaryPath, runArgs, err
}

func getProfileFlags() (map[string]string, bool) {
	flags := map[string]string{}

	for _, fn := range []string{"benchmem", "blockprofile", "blockprofilerate", "coverprofile", "cpuprofile", "memprofile", "memprofilerate", "mutexprofile", "mutexprofilefraction", "outputdir", "trace"} {
		if f := flag.Lookup("test." + fn); f != nil && f.Value.String() != f.DefValue {
			flags[fn] = fmt.Sprintf("%v", f.Value)
		}
	}

	return flags, len(flags) > 0
}
