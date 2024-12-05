package huggingface_test

import (
	"bytes"
	"encoding/json"
	code "fmt"
	"os"

	"github.com/couchbaselabs/cbmigrate/cmd/common"
	"github.com/couchbaselabs/cbmigrate/cmd/huggingface"
	"github.com/couchbaselabs/cbmigrate/internal/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type OutputCapturer struct {
	originalStdout *os.File
	originalStderr *os.File
	//stdoutReader   *os.File
	stdoutWriter *os.File
	stdoutBuf    struct {
		buffer bytes.Buffer
		err    error
	}
	stderrBuf struct {
		buffer bytes.Buffer
		err    error
	}
	//stderrReader *os.File
	stderrWriter *os.File
}

func NewOutputCapturer() *OutputCapturer {
	return &OutputCapturer{}
}

func (c *OutputCapturer) Start() error {
	var err error
	var stdoutReader, stderrReader *os.File
	// Save original outputs
	c.originalStdout = os.Stdout
	c.originalStderr = os.Stderr

	// Create pipes for capturing output
	stdoutReader, c.stdoutWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	stderrReader, c.stderrWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	// Redirect outputs
	os.Stdout = c.stdoutWriter
	os.Stderr = c.stderrWriter

	// Read captured output
	go func() {
		_, c.stdoutBuf.err = c.stdoutBuf.buffer.ReadFrom(stdoutReader)
	}()
	go func() {
		_, c.stderrBuf.err = c.stderrBuf.buffer.ReadFrom(stderrReader)
	}()

	return nil
}

func (c *OutputCapturer) Stop() (string, string, error) {
	zap.L().Sync()
	// Restore original outputs
	os.Stdout = c.originalStdout
	os.Stderr = c.originalStderr

	// Close writers to flush data
	c.stdoutWriter.Close()
	c.stderrWriter.Close()

	if c.stdoutBuf.err != nil || c.stderrBuf.err != nil {
		return "", "", fmt.Errorf("error reading from stdout or stderr: %v, %v", c.stdoutBuf.err, c.stderrBuf.err)
	}
	return c.stdoutBuf.buffer.String(), c.stderrBuf.buffer.String(), nil
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
			_, err = common.ExecuteCommand(cmd, "list-configs", "--path", "path", "--json-output")
			Expect(err).NotTo(HaveOccurred())
			output, errorOutput, err := oc.Stop()
			Expect(err).NotTo(HaveOccurred())
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
				"json_output":          true,
				"debug":                false,
			}), "'options' should be a list")
		})
	})
})
