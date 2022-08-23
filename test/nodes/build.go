//go:build linux

package nodes

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/onsi/gomega/gexec"
	"go.uber.org/zap"
	"riasc.eu/wice/test"
)

var (
	testBinaryPath string
)

func BuildTestBinary(name string) (string, []any, error) {
	var err error
	var runArgs = []any{}

	profileFlags := getProfileFlags()

	// Set some agent specific paths for the profile
	if len(profileFlags) > 0 {
		if _, ok := profileFlags["outputdir"]; !ok {
			profileFlags["outputdir"] = "."
		}

		for _, prof := range []string{"blockprofile", "coverprofile", "cpuprofile", "memprofile", "mutexprofile", "trace"} {
			if path, ok := profileFlags[prof]; ok {
				fn := filepath.Base(path)

				profileFlags[prof] = fmt.Sprintf("%s.%s", name, fn)
			}
		}

		for k, v := range profileFlags {
			runArg := fmt.Sprintf("-test.%s=%v", k, v)

			runArgs = append(runArgs, runArg)
		}
	}

	if testBinaryPath == "" {
		buildArgs := []string{}

		// Pass-through -race option from Ginkgo to wice binary
		if test.IsRace {
			buildArgs = append(buildArgs, "-race")
		}

		// Build a test binary if profiling is requested
		if len(profileFlags) > 0 {
			buildArgs = append(buildArgs, "-tags", "test")

			for k, v := range profileFlags {
				buildArg := fmt.Sprintf("-%s=%v", k, v)
				buildArgs = append(buildArgs, buildArg)
			}

			zap.L().Info("building test binary")

			// We compile a dummy go test binary here which just
			// invokes main(), but is instrumented for profiling.
			testBinaryPath, err = gexec.CompileTest("../cmd/wice", buildArgs...)
		} else {
			testBinaryPath, err = gexec.Build("../cmd/wice", buildArgs...)
		}
	}

	return testBinaryPath, runArgs, err
}

func getProfileFlags() map[string]string {
	flags := map[string]string{}

	for _, fn := range []string{"benchmem", "blockprofile", "blockprofilerate", "coverprofile", "cpuprofile", "memprofile", "memprofilerate", "mutexprofile", "mutexprofilefraction", "outputdir", "trace"} {
		if f := flag.Lookup("test." + fn); f != nil && f.Value.String() != f.DefValue {
			flags[fn] = fmt.Sprintf("%v", f.Value)
		}
	}

	return flags
}