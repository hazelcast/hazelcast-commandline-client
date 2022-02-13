/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cobraprompt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/hazelcast/hazelcast-commandline-client/commands/common"
)

// DynamicSuggestionsAnnotation for dynamic suggestions.
const DynamicSuggestionsAnnotation = "cobra-prompt-dynamic-suggestions"

// CobraPrompt given a Cobra command it will make every flag and sub commands available as suggestions.
// Command.Short will be used as description for the suggestion.
type CobraPrompt struct {
	// RootCmd is the start point, all its sub commands and flags will be available as suggestions
	RootCmd *cobra.Command
	// GoPromptOptions is for customize go-prompt
	// see https://github.com/c-bata/go-prompt/blob/master/option.go
	GoPromptOptions []prompt.Option
	// DynamicSuggestionsFunc will be executed if a command has CallbackAnnotation as an annotation. If it's included
	// the value will be provided to the DynamicSuggestionsFunc function.
	DynamicSuggestionsFunc func(annotationValue string, document *prompt.Document) []prompt.Suggest
	// PersistFlagValues will persist flags. For example have verbose turned on every command.
	PersistFlagValues bool
	// FlagsToExclude is a list of flag names to specify flags to exclude from suggestions
	FlagsToExclude []string
	// Suggests flags without user type "-"
	SuggestFlagsWithoutDash bool
	// ShowHelpCommandAndFlags will make help command and flag for every command available.
	ShowHelpCommandAndFlags bool
	// DisableCompletionCommand will disable the default completion command for cobra
	DisableCompletionCommand bool
	// ShowHiddenCommands makes hidden commands available
	ShowHiddenCommands bool
	// ShowHiddenFlags makes hidden flags available
	ShowHiddenFlags bool
	// AddDefaultExitCommand adds a command for exiting prompt loop
	AddDefaultExitCommand bool
	// OnErrorFunc handle error for command.Execute, if not set print error and exit
	OnErrorFunc func(err error)
	Persister   common.NamePersister
}

var ErrExit = errors.New("exit prompt")

// Terminal breaks on os.Exit for go-prompt https://github.com/c-bata/go-prompt/issues/59#issuecomment-376002177
func exitPromptSafely() {
	panic(ErrExit)
}

func handleExit() {
	switch v := recover().(type) {
	case nil:
		return
	case error:
		if errors.Is(v, ErrExit) {
			return
		}
		fmt.Println(v)
	default:
		fmt.Println(v)
		fmt.Println(string(debug.Stack()))
	}
}

// Run will automatically generate suggestions for all cobra commands and flags defined by RootCmd and execute the selected commands.
// Run will also reset all given flags by default, see PersistFlagValues
func (co CobraPrompt) Run(ctx context.Context) {
	defer handleExit()
	// let ctrl+c exit prompt
	co.GoPromptOptions = append(co.GoPromptOptions, prompt.OptionAddKeyBind(prompt.KeyBind{
		Key: prompt.ControlC,
		Fn: func(_ *prompt.Buffer) {
			exitPromptSafely()
		},
	}), prompt.OptionAddKeyBind(prompt.KeyBind{
		Key: prompt.Key(86),
		Fn: func(b *prompt.Buffer) {
			to := b.Document().FindEndOfCurrentWordWithSpace()
			b.CursorRight(to)
		},
	}), prompt.OptionAddKeyBind(prompt.KeyBind{
		Key: prompt.ControlRight,
		Fn: func(b *prompt.Buffer) {
			to := b.Document().FindEndOfCurrentWordWithSpace()
			b.CursorRight(to)
		},
	}))
	if co.RootCmd == nil {
		panic("RootCmd is not set. Please set RootCmd")
	}
	ctx = common.SetContext(ctx, co.Persister)
	co.prepare()
	p := prompt.New(
		func(in string) {
			// do not execute root command if no input given
			if in == "" {
				return
			}
			promptArgs, err := shlex.Split(in)
			if err != nil {
				co.RootCmd.Println("unable to parse commands")
				return
			}
			os.Args = append([]string{os.Args[0]}, promptArgs...)
			if err := co.RootCmd.ExecuteContext(ctx); err != nil {
				if errors.Is(err, ErrExit) {
					exitPromptSafely()
					return
				}
				if co.OnErrorFunc != nil {
					co.OnErrorFunc(err)
				} else {
					co.RootCmd.PrintErrln(err)
					exitPromptSafely()
				}
			}
			// clear all flag values after each command in interactive mode
			clearAllFlags(co.RootCmd, co.FlagsToExclude)
		},
		func(d prompt.Document) []prompt.Suggest {
			// no suggestion on new line
			if d.Text == "" {
				return nil
			}
			return findSuggestions(&co, &d)
		},
		co.GoPromptOptions...,
	)
	p.Run()
}

func clearAllFlags(p *cobra.Command, exclude []string) {
	p.Flags().Visit(func(flag *pflag.Flag) {
		if stringInSlice(exclude, flag.Name) {
			return
		}
		// ignore err since we are setting default value
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	for _, cmd := range p.Commands() {
		clearAllFlags(cmd, exclude)
	}
}

func (co CobraPrompt) prepare() {
	if co.ShowHelpCommandAndFlags {
		co.RootCmd.InitDefaultHelpCmd()
	}
	if co.DisableCompletionCommand {
		co.RootCmd.CompletionOptions.DisableDefaultCmd = true
	}
	if co.AddDefaultExitCommand {
		co.RootCmd.AddCommand(&cobra.Command{
			Use:           "exit",
			Short:         "Exit prompt",
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return ErrExit
			},
		})
	}
}

func findSuggestions(co *CobraPrompt, d *prompt.Document) []prompt.Suggest {
	command := co.RootCmd
	args := strings.Fields(d.CurrentLine())
	if found, _, err := command.Find(args); err == nil {
		command = found
	}
	var suggestions []prompt.Suggest
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden && !co.ShowHiddenFlags {
			return
		}
		if stringInSlice(co.FlagsToExclude, flag.Name) {
			return
		}
		flagUsage := "--" + flag.Name
		if strings.HasPrefix(d.GetWordBeforeCursor(), "--") {
			suggestions = append(suggestions, prompt.Suggest{Text: flagUsage, Description: flag.Usage})
		} else if (co.SuggestFlagsWithoutDash && d.GetWordBeforeCursor() == "") || strings.HasPrefix(d.GetWordBeforeCursor(), "-") {
			if flag.Shorthand != "" {
				suggestions = append(suggestions, prompt.Suggest{Text: fmt.Sprintf("-%s", flag.Shorthand), Description: fmt.Sprintf("or %s %s", flagUsage, flag.Usage)})
				return
			}
			suggestions = append(suggestions, prompt.Suggest{Text: flagUsage, Description: flag.Usage})
		}
	}
	command.LocalFlags().VisitAll(addFlags)
	command.InheritedFlags().VisitAll(addFlags)
	if command.HasAvailableSubCommands() {
		for _, c := range command.Commands() {
			if !c.Hidden && !co.ShowHiddenCommands {
				suggestions = append(suggestions, prompt.Suggest{Text: c.Name(), Description: c.Short})
			}
			if co.ShowHelpCommandAndFlags {
				c.InitDefaultHelpFlag()
			}
		}
	}
	annotation := command.Annotations[DynamicSuggestionsAnnotation]
	if co.DynamicSuggestionsFunc != nil && annotation != "" {
		suggestions = append(suggestions, co.DynamicSuggestionsFunc(annotation, d)...)
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func stringInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}
