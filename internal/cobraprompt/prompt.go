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
)

// DynamicSuggestionsAnnotation for dynamic suggestions.
const DynamicSuggestionsAnnotation = "cobra-prompt-dynamic-suggestions"

// PersistFlagValuesFlag the flag that will be available when PersistFlagValues is true
const PersistFlagValuesFlag = "persist-flag-values"

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
}

var ExitError = errors.New("exit prompt")

// Terminal breaks on os.Exit for go-prompt https://github.com/c-bata/go-prompt/issues/59#issuecomment-376002177
func exitPromptSafely() {
	panic(ExitError)
}

func handleExit() {
	switch v := recover().(type) {
	case nil:
		return
	case error:
		if v == ExitError {
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
	}))
	if co.RootCmd == nil {
		panic("RootCmd is not set. Please set RootCmd")
	}
	co.prepare()
	p := prompt.New(
		func(in string) {
			// do not execute root command if no input given
			if in == "" {
				return
			}
			promptArgs, err := shlex.Split(in)
			if err != nil {
				fmt.Println("unable to parse commands")
				return
			}
			os.Args = append([]string{os.Args[0]}, promptArgs...)
			if err := co.RootCmd.ExecuteContext(ctx); err != nil {
				if err == ExitError {
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
				return ExitError
			},
		})
	}
}

func RegisterPersistFlag(co *cobra.Command) {
	co.PersistentFlags().BoolP(PersistFlagValuesFlag, "",
		false, "Persist last given value for flags")
}

func findSuggestions(co *CobraPrompt, d *prompt.Document) []prompt.Suggest {
	command := co.RootCmd
	args := strings.Fields(d.CurrentLine())
	if found, _, err := command.Find(args); err == nil {
		command = found
	}
	var suggestions []prompt.Suggest
	persistFlagValues, err := command.Flags().GetBool(PersistFlagValuesFlag)
	if err != nil {
		fmt.Println("cannot parse persist flag err: ", err)
	}
	addFlags := func(flag *pflag.Flag) {
		if flag.Changed && !persistFlagValues {
			flag.Value.Set(flag.DefValue)
		}
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
				suggestions = append(suggestions, prompt.Suggest{Text: "-" + flag.Shorthand, Description: "or " + flagUsage + " " + flag.Usage})
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
