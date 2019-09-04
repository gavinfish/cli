// Copyright © 2019 The Tekton Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliopts "k8s.io/cli-runtime/pkg/genericclioptions"
)

type createOptions struct {
	from string
	edit bool
}

func createCommand(p cli.Params) *cobra.Command {
	f := cliopts.NewPrintFlags("create")
	opts := &createOptions{from: ""}
	eg := `
# Create a Task defined by foo.yaml in namespace 'bar'
tkn task create -f foo.yaml -n bar
`

	c := &cobra.Command{
		Use:          "create",
		Short:        "Create a task resource in a namespace",
		Example:      eg,
		SilenceUsage: true,
		Annotations: map[string]string{
			"commandType": "main",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &cli.Stream{
				In:  cmd.InOrStdin(),
				Out: cmd.OutOrStdout(),
				Err: cmd.OutOrStderr(),
			}

			return createTask(s, p, opts.from)
		},
	}
	f.AddFlags(c)
	c.Flags().StringVarP(&opts.from, "from", "f", "", "Filename to use to create the resource")
	c.Flags().BoolVar(&opts.edit, "edit", false, "Edit the task resource before creating")
	return c
}

func createTask(s *cli.Stream, p cli.Params, path string) error {
	cs, err := p.Clients()
	if err != nil {
		return fmt.Errorf("failed to create tekton client")
	}

	task, err := loadTask(path)
	if err != nil {
		return err
	}

	existTask, err := cs.Tekton.TektonV1alpha1().Tasks(p.Namespace()).Get(task.Name, v1.GetOptions{})
	if existTask != nil {
		return fmt.Errorf("task %q has already exists in namespace %s", task.Name, p.Namespace())
	}

	if _, err = cs.Tekton.TektonV1alpha1().Tasks(p.Namespace()).Create(task); err != nil {
		return fmt.Errorf("failed to creat task %q: %s", task.Name, err)
	}

	filename := filepath.Base(path)
	fmt.Fprintf(s.Out, "Task created: %s\n", filename)
	return nil
}

func loadFileContent(target string) ([]byte, error) {
	var content []byte
	var err error
	if strings.HasSuffix(target, ".yaml") || strings.HasSuffix(target, ".yml") {
		if strings.HasPrefix(target, "http") {
			resp, err := http.Get(target)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(resp.Body)
			content = buf.Bytes()
		} else {
			content, err = ioutil.ReadFile(target)
			if err != nil {
				return nil, err
			}
		}
		content, err = yaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("does not support such extension for %s", target)
	}
	return content, nil
}

func loadTask(target string) (*v1alpha1.Task, error) {
	content, err := loadFileContent(target)
	if err != nil {
		return nil, err
	}
	var task v1alpha1.Task
	err = json.Unmarshal(content, &task)
	if err != nil {
		return nil, err
	}

	if task.Kind != "Task" {
		return nil, fmt.Errorf("provided %s instead of Task kind", task.Kind)
	}
	return &task, nil
}

//func RunEditOnCreate(printFlags *genericclioptions.PrintFlags, recordFlags *genericclioptions.RecordFlags, ioStreams genericclioptions.IOStreams, cmd *cobra.Command) {
//
//	editOptions := editor.NewEditOptions(editor.EditBeforeCreateMode, ioStreams)
//	//editOptions.FilenameOptions = *options
//	//editOptions.ValidateOptions = cmdutil.ValidateOptions{
//	//	EnableValidation: cmdutil.GetFlagBool(cmd, "validate"),
//	//}
//	editOptions.PrintFlags = printFlags
//	editOptions.ApplyAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)
//	editOptions.RecordFlags = recordFlags
//
//	err := editOptions.Complete(f, []string{}, cmd)
//	if err != nil {
//		return err
//	}
//	return editOptions.Run()
//}
