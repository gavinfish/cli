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

package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/tektoncd/cli/pkg/cli"
)

type TypeValidator func(target string) bool

func IsYamlFile() TypeValidator {
	return func(target string) bool {
		if strings.HasSuffix(target, ".yaml") || strings.HasSuffix(target, ".yml") {
			return true
		}
		return false
	}
}

func LoadFileContent(p cli.Params, target string, validate TypeValidator) ([]byte, error) {
	if !validate(target) {
		return nil, fmt.Errorf("the path %s does not match target pattern", target)
	}

	var content []byte
	cs, err := p.Clients()
	if err != nil {
		return nil, fmt.Errorf("failed to create tekton client")
	}
	if strings.HasPrefix(target, "http") {
		content, err = getRemoteContent(cs, target)
	} else {
		content, err = ioutil.ReadFile(target)
	}
	if err != nil {
		return nil, err
	}
	content, err = yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func getRemoteContent(cs *cli.Clients, url string) ([]byte, error) {
	resp, err := cs.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	content := buf.Bytes()
	return content, nil
}