// Copyright 2020 CUE Authors
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

package exec

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"cuelang.org/go/cue"
	"cuelang.org/go/internal/task"
	"cuelang.org/go/pkg/internal"
)

func TestEnv(t *testing.T) {
	testCases := []struct {
		desc string
		val  string
		env  []string
	}{
		{
			desc: "mapped",
			val: `
		cmd: "echo"
		env: {
			WHO:  "World"
			WHAT: "Hello"
			WHEN: "Now!"
		}
		`,
			env: []string{"WHO=World", "WHAT=Hello", "WHEN=Now!"},
		},
		{
			desc: "list",
			val: `
		cmd: "echo"
		env: ["WHO=World", "WHAT=Hello", "WHEN=Now!"]
		`,
			env: []string{"WHO=World", "WHAT=Hello", "WHEN=Now!"},
		},
		{
			desc: "struct handles default values",
			val: `
		cmd: "echo"
		env: {
			WHEN: *"Now!" | string
			HOW: *WHEN | string
		}
		`,
			env: []string{"WHEN=Now!", "HOW=Now!"},
		},
		{
			desc: "list handles default values",
			val: `
		cmd: "echo"
		env: ["WHO=World", "WHAT=Hello", *"COMMAND=\(cmd)" | string]
		`,
			env: []string{"WHO=World", "WHAT=Hello", "COMMAND=echo"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := internal.NewContext()
			v := ctx.CompileString(tc.val, cue.Filename(tc.desc))
			if err := v.Err(); err != nil {
				t.Fatal(err)
			}

			cmd, _, err := mkCommand(&task.Context{
				Context: context.Background(),
				Obj:     v,
			})
			if err != nil {
				t.Fatalf("mkCommand error = %v", err)
			}

			if diff := cmp.Diff(cmd.Env, tc.env); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestArgs(t *testing.T) {
	testCases := []struct {
		desc string
		val  string
		args []string
		doc  string
		err  string
	}{
		{
			desc: "string cmd parses args correctly",
			val: `
			cmd: "bash -c \"echo hello world\""
			`,
			doc:  `bash -c 'echo hello world'`,
			args: []string{"bash", "-c", "echo hello world"},
		},
		{
			desc: "list cmd doc is joined correctly",
			val: `
			cmd: ["bash", "-c", "echo hello world"]
			`,
			doc:  `bash -c 'echo hello world'`,
			args: []string{"bash", "-c", "echo hello world"},
		},
		{
			desc: "empty cmd string errors correctly",
			val: `
			cmd: ""
			`,
			err: "empty command",
		},
		{
			desc: "empty cmd list errors correctly",
			val: `
			cmd: []
			`,
			err: "empty command",
		},
		{
			desc: "cmd string with starting space",
			val: `
			cmd: " abc"
			`,
			doc:  "abc",
			args: []string{"abc"},
		},
		{
			desc: "cmd list with empty bin",
			val: `
			cmd: [""]
			`,
			err: "empty command",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := internal.NewContext()
			v := ctx.CompileString(tc.val, cue.Filename(tc.desc))
			if err := v.Err(); err != nil {
				t.Fatal(err)
			}

			cmd, doc, err := mkCommand(&task.Context{
				Context: context.Background(),
				Obj:     v,
			})
			if err != nil {
				if diff := cmp.Diff(err.Error(), tc.err); diff != "" {
					t.Error(diff)
				}
				return
			}

			if diff := cmp.Diff(doc, tc.doc); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(cmd.Args, tc.args); diff != "" {
				t.Error(diff)
			}
		})
	}
}
