// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/test"
)

// We implement our own functions for building the binary-under-test
// as Gomega's gexec.Build() and gexec.CompileTest() always use a temporary
// directory which invalidates Go's build cache hence the binary will be relinked
// which consume some time we can save during the tests.

//nolint:gochecknoglobals, errname
var (
	binaryError   error
	binaryPath    string
	binaryRunArgs []string
	binaryOnce    sync.Once
)

func BuildBinary(name string) (string, []any, error) {
	binaryOnce.Do(func() {
		binaryPath, binaryRunArgs, binaryError = buildBinary("../../cmd/cunicu")
	})
	if binaryError != nil {
		return "", nil, binaryError
	}

	runArgs := []any{}
	for _, runArg := range binaryRunArgs {
		// We build the binary once but customize the profile output
		// paths per based on the node name.
		runArgs = append(runArgs, strings.ReplaceAll(runArg, "{name}", name))
	}

	return binaryPath, runArgs, nil
}

// You should call CleanupBinary before your test ends to clean up any temporary artifacts.
func CleanupBinary() {
	cacheDir := os.Getenv("GINKGO_CACHE_DIR")
	if dir := filepath.Dir(binaryPath); dir != "" && dir != cacheDir {
		os.RemoveAll(dir)
	}
}

func buildBinary(packagePath string) (string, []string, error) {
	path, err := newExecutablePath(packagePath)
	if err != nil {
		return "", nil, err
	}

	runArgs := []string{}
	profileArgs := profileArgs()
	buildArgs := []string{
		"-buildvcs=false", // avoid build cache invalidation
	}

	// Pass-through -race option from Ginkgo to cunīcu binary
	if test.IsRace {
		buildArgs = append(buildArgs, "-race")
	}

	// Set some agent specific paths for the profile
	if len(profileArgs) > 0 {
		if _, ok := profileArgs["outputdir"]; !ok {
			profileArgs["outputdir"] = "."
		}

		for _, prof := range []string{"blockprofile", "coverprofile", "cpuprofile", "memprofile", "mutexprofile", "trace"} {
			if _, ok := profileArgs[prof]; ok {
				profileArgs[prof] = fmt.Sprintf("{name}.%s.out", prof)
			}
		}

		for k, v := range profileArgs {
			runArg := fmt.Sprintf("-test.%s=%v", k, v)

			runArgs = append(runArgs, runArg)
		}
	}

	logger := log.Global.Named("builder")

	var cmdArgs []string
	var test bool

	// Build a test binary if profiling is requested
	if test = len(profileArgs) > 0; test {
		// Generate coverage for all cunīcu packages
		profileArgs["coverpkg"] = "../../..."

		buildArgs = append(buildArgs, "-tags", "test")

		for k, v := range profileArgs {
			buildArg := fmt.Sprintf("-%s=%v", k, v)
			buildArgs = append(buildArgs, buildArg)
		}

		// We compile a dummy go test binary here which just
		// invokes main(), but is instrumented for profiling.

		cmdArgs = append([]string{"test", "-c"}, buildArgs...)
		cmdArgs = append(cmdArgs, "-o", path, packagePath)
	} else {
		cmdArgs = append([]string{"build"}, buildArgs...)
		cmdArgs = append(cmdArgs, "-o", path, packagePath)
	}

	start := time.Now()

	logger.Debug("Start building test",
		zap.Strings("go_args", cmdArgs),
		zap.Strings("run_args", runArgs),
		zap.Bool("test", test))

	if output, err := exec.Command("go", cmdArgs...).CombinedOutput(); err != nil {
		return "", nil, fmt.Errorf("failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, path, string(output))
	}

	logger.Debug("Finished building",
		zap.Error(binaryError), zap.Duration("time", time.Since(start)))

	return path, runArgs, nil
}

func profileArgs() map[string]string {
	flags := map[string]string{}

	for _, fn := range []string{"benchmem", "blockprofile", "blockprofilerate", "coverprofile", "cpuprofile", "memprofile", "memprofilerate", "mutexprofile", "mutexprofilefraction", "outputdir", "trace", "coverpkg"} {
		if f := flag.Lookup("test." + fn); f != nil && f.Value.String() != f.DefValue {
			flags[fn] = fmt.Sprintf("%v", f.Value)
		}
	}

	return flags
}

func newExecutablePath(packagePath string) (string, error) {
	tmpDir, err := temporaryDirectory()
	if err != nil {
		return "", err
	}

	executable := filepath.Join(tmpDir, filepath.Base(packagePath))

	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	return executable, nil
}

func temporaryDirectory() (string, error) {
	if tmpDir := os.Getenv("GINKGO_CACHE_DIR"); tmpDir != "" {
		return tmpDir, nil
	}

	tmpDir, err := os.MkdirTemp("", "cunicu_test_artifacts")
	if err != nil {
		return "", err
	}

	return tmpDir, nil
}
