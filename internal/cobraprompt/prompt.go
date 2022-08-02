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
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/google/shlex"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	goprompt "github.com/hazelcast/hazelcast-commandline-client/internal/go-prompt"
	"github.com/hazelcast/hazelcast-commandline-client/rootcmd"
)

// DynamicSuggestionsAnnotation for dynamic suggestions.
const DynamicSuggestionsAnnotation = "cobra-prompt-dynamic-suggestions"

// CobraPrompt given a Cobra command it will make every flag and sub commands available as suggestions.
// Command.Short will be used as description for the suggestion.
type CobraPrompt struct {
	// GoPromptOptions is for customize go-goprompt
	// see https://github.com/c-bata/go-prompt/blob/master/option.go
	GoPromptOptions []goprompt.Option
	// DynamicSuggestionsFunc will be executed if a command has CallbackAnnotation as an annotation. If it's included
	// the value will be provided to the DynamicSuggestionsFunc function.
	DynamicSuggestionsFunc func(annotationValue string, document *goprompt.Document) []goprompt.Suggest
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
	// AddDefaultExitCommand adds a command for exiting goprompt loop
	AddDefaultExitCommand bool
	// OnErrorFunc handle error for command.Execute, if not set print error and exit
	OnErrorFunc func(err error)
	// Persister is used to interact with name persistence mechanism.
	Persister map[string]string
	// NoColor is used to disable colors
	NoColor bool
	// DisableSuggestions disables the suggestion prompt
	DisableSuggestions bool
}

var ErrExit = errors.New("exit prompt")

// Terminal breaks on os.Exit for go-goprompt https://github.com/c-bata/go-prompt/issues/59#issuecomment-376002177
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

var SuggestionColorOptions = []goprompt.Option{
	goprompt.OptionSelectedSuggestionTextColor(goprompt.White), goprompt.OptionSuggestionTextColor(goprompt.White),
	goprompt.OptionSelectedDescriptionTextColor(goprompt.LightGray), goprompt.OptionDescriptionTextColor(goprompt.LightGray),
	goprompt.OptionSelectedSuggestionBGColor(goprompt.Blue), goprompt.OptionSuggestionBGColor(goprompt.DarkGray),
	goprompt.OptionSelectedDescriptionBGColor(goprompt.Blue), goprompt.OptionDescriptionBGColor(goprompt.DarkGray),
}

var NoColorOptions = []goprompt.Option{
	goprompt.OptionSelectedSuggestionTextColor(goprompt.NoColor), goprompt.OptionSuggestionTextColor(goprompt.NoColor),
	goprompt.OptionSelectedDescriptionTextColor(goprompt.NoColor), goprompt.OptionDescriptionTextColor(goprompt.NoColor),
	goprompt.OptionSelectedSuggestionBGColor(goprompt.NoColor), goprompt.OptionSuggestionBGColor(goprompt.NoColor),
	goprompt.OptionPrefixTextColor(goprompt.NoColor), goprompt.OptionPrefixBackgroundColor(goprompt.NoColor),
	goprompt.OptionInputTextColor(goprompt.NoColor), goprompt.OptionPreviewSuggestionTextColor(goprompt.NoColor),
	goprompt.OptionSelectedDescriptionBGColor(goprompt.NoColor), goprompt.OptionDescriptionBGColor(goprompt.NoColor),
}

// Run will automatically generate suggestions for all cobra commands and flags defined by RootCmd and execute the selected commands.
// Run will also reset all given flags by default, see PersistFlagValues
func (co CobraPrompt) Run(ctx context.Context, root *cobra.Command, cnfg *hazelcast.Config, cmdHistoryPath string) {
	defer handleExit()
	// let ctrl+c exit goprompt
	co.GoPromptOptions = append(co.GoPromptOptions, goprompt.OptionAddKeyBind(goprompt.KeyBind{
		Key: goprompt.ControlC,
		Fn: func(_ *goprompt.Buffer) {
			exitPromptSafely()
		},
	}), goprompt.OptionAddKeyBind(goprompt.KeyBind{
		Key: goprompt.ControlLeft,
		Fn: func(b *goprompt.Buffer) {
			to := b.Document().FindPreviousWordStart()
			b.CursorLeft(to)
		},
	}), goprompt.OptionAddKeyBind(goprompt.KeyBind{
		Key: goprompt.ControlRight,
		Fn: func(b *goprompt.Buffer) {
			to := b.Document().FindEndOfCurrentWordWithSpace()
			b.CursorRight(to)
		},
	}))
	if co.NoColor {
		co.GoPromptOptions = append(co.GoPromptOptions, NoColorOptions...)
		//co.GoPromptOptions = append(co.GoPromptOptions, SuggestionColorOptions...)
	} else {
		co.GoPromptOptions = append(co.GoPromptOptions, SuggestionColorOptions...)
	}
	history := goprompt.NewHistory()
	f, err := os.OpenFile(cmdHistoryPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	defer func() {
		f.Close()
	}()
	if err != nil {
		// todo log this once we have a logging solution
	} else {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			history.Add(scanner.Text())
		}
		if scanner.Err() != nil {
			root.Printf("Cannot load command history, will proceed without one\nhistory file on %s:%s...\n", cmdHistoryPath, err)
			history.Clear()
		}
	}
	ctx = internal.ContextWithPersistedNames(ctx, co.Persister)
	var p *goprompt.Prompt
	p = goprompt.New(
		func(in string) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, os.Kill)
			go func() {
				select {
				case <-c:
					cancel()
				case <-ctx.Done():
				}
			}()
			// do not execute root command if no input given
			if in == "" {
				return
			}
			promptArgs, err := shlex.Split(in)
			if err != nil {
				fmt.Println("unable to parse commands")
				return
			}
			// re-init command chain every iteration
			// ignore global flags, they are already parsed
			root, _ = rootcmd.New(cnfg)
			prepareRootCmdForPrompt(co, root)
			root.SetArgs(promptArgs)
			root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
				return hzcerrors.FlagError(err)
			})
			os.Args = promptArgs
			err = root.ExecuteContext(ctx)
			if _, writeErr := f.WriteString(fmt.Sprintln(in)); writeErr != nil {
				// todo log this once we have a logging solution
			}
			if err != nil {
				if errors.Is(err, ErrExit) {
					exitPromptSafely()
					return
				}
				// todo find a better approach than string comparison, this is fragile
				if err.Error() == `required flag(s) "name" not set` {
					// todo make this applicable for all data types
					err = fmt.Errorf(`%w. Add it or consider "map use <name>"`, err)
				}
				if co.OnErrorFunc != nil {
					co.OnErrorFunc(err)
				} else {
					root.PrintErrln(err)
					exitPromptSafely()
				}
			}
			// clear screen only after sql browser command executed successfully
			if strings.Trim(in, " ") == "sql" {
				// lets us invoke ctrl+L shortcut which clears the screen
				p.Feed([]byte{0xc})
			}
		},
		func(d goprompt.Document) []goprompt.Suggest {
			if co.DisableSuggestions {
				return nil
			}
			// no suggestion on new line
			if d.Text == "" {
				return nil
			}
			return findSuggestions(&co, root, &d)
		},
		co.GoPromptOptions...,
	)
	p.History = history
	p.Run()
}

func prepareRootCmdForPrompt(co CobraPrompt, root *cobra.Command) {
	if co.ShowHelpCommandAndFlags {
		root.InitDefaultHelpCmd()
	}
	if co.DisableCompletionCommand {
		root.CompletionOptions.DisableDefaultCmd = true
	}
	if co.AddDefaultExitCommand {
		root.AddCommand(&cobra.Command{
			Use:           "exit",
			Short:         "Exit goprompt",
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return ErrExit
			},
		})
	}
	root.Example = `> map put -k key -n myMap -v someValue
> map get -k key -m myMap
> cluster version
> sql`
	root.Use = ""
}

func findSuggestions(co *CobraPrompt, cmd *cobra.Command, d *goprompt.Document) []goprompt.Suggest {
	upToCursor := d.CurrentLineBeforeCursor()
	// use line before cursor for command suggestion
	bArgs := strings.Fields(upToCursor)
	command, _, err := cmd.Find(bArgs)
	if err != nil && strings.Contains(upToCursor, " ") {
		return nil
	}
	wordBeforeCursor := d.GetWordBeforeCursor()
	// use whole line for flag suggestions
	args := strings.Fields(d.CurrentLine())
	suggestions := traverseForFlagSuggestions(wordBeforeCursor, args, co, command)
	if command.HasAvailableSubCommands() {
		for _, c := range command.Commands() {
			if !c.Hidden && !co.ShowHiddenCommands {
				suggestions = append(suggestions, goprompt.Suggest{Text: c.Name(), Description: c.Short})
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
	return goprompt.FilterHasPrefix(suggestions, wordBeforeCursor, true)
}

func traverseForFlagSuggestions(wordBeforeCursor string, words []string, co *CobraPrompt, command *cobra.Command) []goprompt.Suggest {
	var suggestions []goprompt.Suggest
	noWordTyped := wordBeforeCursor == ""
	dashPrefix := strings.HasPrefix(wordBeforeCursor, "-")
	if !noWordTyped && !dashPrefix {
		// no flag prefix
		return suggestions
	}
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden && !co.ShowHiddenFlags {
			return
		}
		if stringInSlice(co.FlagsToExclude, flag.Name, true) {
			return
		}
		flagUsage := "--" + flag.Name
		// Check if flag is already used in the command
		if (flag.Shorthand != "" && stringInSlice(words, "-"+flag.Shorthand, false)) ||
			stringInSlice(words, flagUsage, false) {
			return
		}
		if strings.HasPrefix(wordBeforeCursor, "--") {
			suggestions = append(suggestions, goprompt.Suggest{Text: flagUsage, Description: flag.Usage})
		} else if (noWordTyped || dashPrefix) && flag.Shorthand != "" {
			flagShort := fmt.Sprintf("-%s", flag.Shorthand)
			suggestions = append(suggestions, goprompt.Suggest{Text: flagShort, Description: fmt.Sprintf("or %s %s", flagUsage, flag.Usage)})
		} else {
			suggestions = append(suggestions, goprompt.Suggest{Text: flagUsage, Description: flag.Usage})
		}
	}
	command.LocalFlags().VisitAll(addFlags)
	command.InheritedFlags().VisitAll(addFlags)
	return suggestions
}

func stringInSlice(slice []string, str string, caseSensitive bool) bool {
	if !caseSensitive {
		str = strings.ToLower(str)
	}
	for _, s := range slice {
		if !caseSensitive {
			s = strings.ToLower(s)
		}
		if str == s {
			return true
		}
	}
	return false
}
