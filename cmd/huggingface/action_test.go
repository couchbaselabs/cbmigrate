package huggingface_test

import (
	"bytes"
	"encoding/json"
	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/huggingface"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
	"github.com/spf13/cobra"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type OutputCapturer struct {
	originalStdout *os.File
	originalStderr *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	stderrReader   *os.File
	stderrWriter   *os.File
}

func NewOutputCapturer() *OutputCapturer {
	return &OutputCapturer{}
}

func (c *OutputCapturer) Start() error {
	var err error

	// Save original outputs
	c.originalStdout = os.Stdout
	c.originalStderr = os.Stderr

	// Create pipes for capturing output
	c.stdoutReader, c.stdoutWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	c.stderrReader, c.stderrWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	// Redirect outputs
	os.Stdout = c.stdoutWriter
	os.Stderr = c.stderrWriter

	return nil
}

func (c *OutputCapturer) Stop() (string, string, error) {
	// Restore original outputs
	os.Stdout = c.originalStdout
	os.Stderr = c.originalStderr

	// Close writers to flush data
	c.stdoutWriter.Close()
	c.stderrWriter.Close()

	// Read captured output
	var stdoutBuf, stderrBuf bytes.Buffer
	_, err := stdoutBuf.ReadFrom(c.stdoutReader)
	if err != nil {
		return "", "", err
	}

	_, err = stderrBuf.ReadFrom(c.stderrReader)
	if err != nil {
		return "", "", err
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

var _ = Describe("Huggingface", func() {
	Context("when MOCK_CLI_FOR_CBMIGRATE=true is set", func() {
		var (
			cmd *cobra.Command
		)
		BeforeEach(func() {
			// Set the environment variable
			err := os.Setenv("MOCK_CLI_FOR_CBMIGRATE", "true")
			Expect(err).NotTo(HaveOccurred())
			cmd = huggingface.GetHuggingFaceMigrateCommand()
		})
		AfterEach(func() {
			os.Unsetenv("MOCK_CLI_FOR_CBMIGRATE")
		})
		It("should present options in JSON format", func() {

			oc := NewOutputCapturer()
			err := oc.Start()
			logger.Init()
			Expect(err).NotTo(HaveOccurred())
			_, err = common.ExecuteCommand(cmd, "list-configs", "--path", "path")
			Expect(err).NotTo(HaveOccurred())
			output, errorOutput, err := oc.Stop()
			Expect(errorOutput).To(BeEmpty())
			var jsonOutput = map[string]interface{}{}
			err = json.Unmarshal([]byte(output), &jsonOutput)
			Expect(err).NotTo(HaveOccurred(), "Output should be valid JSON")

			Expect(jsonOutput).To(Equal(map[string]interface{}{
				"path":                 "path",
				"revision":             nil,
				"download_config":      nil,
				"download_mode":        nil,
				"dynamic_modules_path": nil,
				"data_files":           []interface{}{},
				"token":                nil,
				"json_output":          false,
				"debug":                false,
			}), "'options' should be a list")
		})
	})
})
