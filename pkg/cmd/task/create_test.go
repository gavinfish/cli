package task

import (
	"io"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/tektoncd/cli/pkg/test"
	cb "github.com/tektoncd/cli/pkg/test/builder"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinetest "github.com/tektoncd/pipeline/test"
	tb "github.com/tektoncd/pipeline/test/builder"
)

func TestTaskCreate(t *testing.T) {
	clock := clockwork.NewFakeClock()

	seeds := make([]pipelinetest.Clients, 0)
	for i := 0; i < 3; i++ {
		cs, _ := test.SeedTestData(t, pipelinetest.Data{})
		seeds = append(seeds, cs)
	}

	tasks := []*v1alpha1.Task{
		tb.Task("build-docker-image-from-git-source", "ns", cb.TaskCreationTime(clock.Now().Add(-1*time.Minute))),
	}
	cs, _ := test.SeedTestData(t, pipelinetest.Data{Tasks: tasks})
	seeds = append(seeds, cs)

	testParams := []struct {
		name        string
		command     []string
		input       pipelinetest.Clients
		inputStream io.Reader
		wantError   bool
		want        string
	}{
		{
			name:        "Create successfully",
			command:     []string{"create", "--from", "../../../test/resources/task.yaml", "-n", "ns"},
			input:       seeds[0],
			inputStream: nil,
			wantError:   false,
			want:        "Task created: task.yaml\n",
		},
		{
			name:        "Create successfully",
			command:     []string{"create", "-f", "../../../test/resources/task.yaml", "-n", "ns"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   false,
			want:        "Task created: task.yaml\n",
		},
		{
			name:        "Filename with wildcard",
			command:     []string{"create", "-f", "../../../test/resources/*.yaml", "-n", "ns"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   true,
			want:        "open ../../../test/resources/*.yaml: The filename, directory name, or volume label syntax is incorrect.",
		},
		{
			name:        "Filename does not exist",
			command:     []string{"create", "-f", "test/resources/task.yaml", "-n", "ns"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   true,
			want:        "open test/resources/task.yaml: The system cannot find the path specified.",
		},
		{
			name:        "Unsupported file type",
			command:     []string{"create", "-f", "../../../test/resources/task.txt", "-n", "ns"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   true,
			want:        "does not support such extension for ../../../test/resources/task.txt",
		},
		{
			name:        "Missing from",
			command:     []string{"create", "-n", "ns", "-f"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   true,
			want:        "flag needs an argument: 'f' in -f",
		},
		{
			name:        "Mismatched resource file",
			command:     []string{"create", "-f", "../../../test/resources/taskrun.yaml", "-n", "ns"},
			input:       seeds[1],
			inputStream: nil,
			wantError:   true,
			want:        "provided TaskRun instead of Task kind",
		},
		{
			name:        "",
			command:     []string{"create", "-f", "https://gist.githubusercontent.com/gavinfish/bab9ee1b0d068f8c0f33c92417dcf178/raw/5a4da2ecc1090f4877af2ab02da74424b5166eb2/task.yaml", "-n", "ns"},
			input:       seeds[2],
			inputStream: nil,
			wantError:   false,
			want:        "Task created: task.yaml\n",
		},
		{
			name:        "Existing task",
			command:     []string{"create", "-f", "../../../test/resources/task.yaml", "-n", "ns"},
			input:       seeds[3],
			inputStream: nil,
			wantError:   true,
			want:        "task \"build-docker-image-from-git-source\" has already exists in namespace ns",
		},
	}

	for _, tp := range testParams {
		t.Run(tp.name, func(t *testing.T) {
			p := &test.Params{Tekton: tp.input.Pipeline}
			task := Command(p)

			if tp.inputStream != nil {
				task.SetIn(tp.inputStream)
			}

			out, err := test.ExecuteCommand(task, tp.command...)
			if tp.wantError {
				if err == nil {
					t.Errorf("Error expected here")
				}
				test.AssertOutput(t, tp.want, err.Error())
			} else {
				if err != nil {
					t.Errorf("Unexpected Error")
				}
				test.AssertOutput(t, tp.want, out)
			}
		})
	}
}
