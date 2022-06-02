package examples

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	ovirtclient "github.com/ovirt/go-ovirt-client"
	ovirtclientlog "github.com/ovirt/go-ovirt-client-log/v3"
)

func TestExamples(t *testing.T) {
	ovirtURL := os.Getenv("OVIRT_URL")
	ovirtUsername := os.Getenv("OVIRT_USERNAME")
	ovirtPassword := os.Getenv("OVIRT_PASSWORD")

	if ovirtURL == "" || ovirtUsername == "" || ovirtPassword == "" {
		t.Skipf("OVIRT_URL, OVIRT_USERNAME, or OVIRT_PASSWORD no set, skipping test.")
	}

	helper, err := ovirtclient.NewTestHelper(
		ovirtURL,
		ovirtUsername,
		ovirtPassword,
		nil,
		ovirtclient.TLS().Insecure(),
		false,
		ovirtclientlog.NewTestLogger(t),
	)
	if err != nil {
		t.Fatalf("Failed to initialize the oVirt client (%v)", err)
	}

	tfVars := tfvars{
		ovirtUsername,
		ovirtPassword,
		ovirtURL,
		helper.GetStorageDomainID(),
		helper.GetClusterID(),
		helper.GetVNICProfileID(),
		true,
	}
	t.Run(
		"provider", func(t *testing.T) {
			env := startTerraform(t, "provider")
			runTerraform(
				t, "provider", tfVars, env,
			)
		},
	)

	for _, category := range []string{"resources", "data-sources"} {
		t.Run(
			category, func(t *testing.T) {
				entries, err := ioutil.ReadDir(category)
				if err != nil {
					t.Fatalf("failed to read directory %s (%v)", category, err)
				}
				for _, e := range entries {
					if e.IsDir() {
						t.Run(
							e.Name(), func(t *testing.T) {
								dir, err := filepath.Abs(path.Join(category, e.Name()))
								if err != nil {
									t.Fatalf("Failed to find absolute path for directory (%v)", err)
								}
								env := startTerraform(t, dir)
								runTerraform(
									t, dir, tfVars, env,
								)
							},
						)
					}
				}
			},
		)
	}
}

func runTerraformCommand(t *testing.T, dir string, env []string, vars interface{}, args ...string) {
	tmpDir := t.TempDir()
	varsFileName := path.Join(tmpDir, "tfvars.json")

	if vars != nil {
		varsFile, err := os.Create(varsFileName)
		if err != nil {
			t.Fatalf("Failed to create temporary file for Terraform variables. (%v)", err)
		}
		defer func() {
			_ = os.Remove(varsFileName)
		}()
		encoder := json.NewEncoder(varsFile)
		if err := encoder.Encode(vars); err != nil {
			t.Fatalf("Failed to encode tfvars (%v)", err)
		}
		if err := varsFile.Close(); err != nil {
			t.Fatalf("Failed to close tfvars file (%v)", err)
		}
		args = append(args, fmt.Sprintf("-var-file=%s", varsFileName))
	}
	t.Logf("Executing terraform %s ...", strings.Join(args, " "))

	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Env = env
	lock := &sync.Mutex{}
	prefix := "[terraform " + strings.Join(args, " ") + "] "
	cmd.Stdout = &tLogger{t, lock, prefix}
	cmd.Stderr = &tLogger{t, lock, prefix}
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run terraform %s (%v)", strings.Join(args, " "), err)
	}
	t.Logf("Successfully ran terrafom %s.", strings.Join(args, " "))
}

func startTerraform(t *testing.T, dir string) []string {
	t.Logf("Starting Terraform provider...")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to obtain current working directory (%v)", err)
	}
	cmdLine := []string{"go", "run", path.Join(cwd, "..", "main.go"), "-debug"}
	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	logger := &tfReattachCaptureLogger{
		t:      t,
		ready:  make(chan string),
		error:  make(chan error, 1),
		prefix: "[go run main.go -debug] ",
		lock:   &sync.Mutex{},
	}
	cmd.Dir = dir
	cmd.Stderr = logger
	cmd.Stdout = logger
	go func() {
		defer close(logger.error)
		t.Logf("Executing %s ...", cmd.Args)
		err := cmd.Run()
		if err != nil {
			t.Logf("Program terminated with an error: %s (%v)", cmd.Args, err)
			logger.error <- err
		}
		t.Logf("Program terminated normally: %s", cmd.Args)
	}()

	env := []string{}
	select {
	case line, ok := <-logger.ready:
		if !ok {
			t.Fatalf("Ready channel closed without submitting a TF_REATTACH_PROVIDERS line")
		}
		lineParts := strings.SplitN(line, "=", 2)
		// Remove extra quotes from JSON output
		lineParts[1] = strings.Trim(lineParts[1], "'")
		line = strings.Join(lineParts, "=")
		env = append(env, line)
	case err, ok := <-logger.error:
		if !ok {
			t.Fatalf("Terraform provider closed the error channel without any output.")
		}
		t.Fatalf("Terraform provider exited with an error: %v", err)
	case <-time.After(60 * time.Second):
		t.Fatalf("Timeout while waiting for the TF_REATTACH_PROVIDERS line in the output.")
	}
	t.Cleanup(
		func() {
			t.Logf("Terminating %s ...", strings.Join(cmd.Args, " "))
			if err := cmd.Process.Kill(); err != nil {
				t.Fatalf("Failed to stop Terraform provider running in the background. (%v)", err)
			}
		},
	)
	return env
}

func runTerraform(t *testing.T, dir string, vars tfvars, env []string) {
	runTerraformCommand(t, dir, env, nil, "init")
	t.Cleanup(
		func() {
			runTerraformCommand(t, dir, env, vars, "destroy", "-auto-approve")
		},
	)
	runTerraformCommand(t, dir, env, vars, "apply", "-auto-approve")
}

type tfvars struct {
	Username        string                      `json:"username"`
	Password        string                      `json:"password"`
	URL             string                      `json:"url"`
	StorageDomainID ovirtclient.StorageDomainID `json:"storage_domain_id"`
	ClusterID       ovirtclient.ClusterID       `json:"cluster_id"`
	VNICProfileID   ovirtclient.VNICProfileID   `json:"vnic_profile_id"`
	TLSInsecure     bool                        `json:"tls_insecure"`
}

var tfReattachRe = regexp.MustCompile("TF_REATTACH_PROVIDERS=.*")

type tfReattachCaptureLogger struct {
	t      *testing.T
	ready  chan string
	error  chan error
	prefix string
	lock   *sync.Mutex
}

func (t tfReattachCaptureLogger) Write(p []byte) (n int, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.t.Helper()

	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, line := range lines {
		t.t.Logf("%s%s", t.prefix, line)
	}

	matches := tfReattachRe.Find(p)
	if len(matches) != 0 {
		t.t.Logf("Found TF_REATTACH_PROVIDERS line in output.")
		t.ready <- string(matches)
		close(t.ready)
	}
	return len(p), nil
}

type tLogger struct {
	t      *testing.T
	lock   *sync.Mutex
	prefix string
}

func (t tLogger) Write(p []byte) (n int, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.t.Helper()

	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, line := range lines {
		t.t.Logf("%s%s", t.prefix, line)
	}
	return len(p), nil
}
