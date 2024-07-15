package appcli

import (
	"fmt"
	"log"
	"time"

	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
	"github.com/urfave/cli/v2"
)

type appImpl struct {
	// The name of the program. Defaults to path.Base(os.Args[0])
	Name string
	// Full name of command for help, defaults to Name
	HelpName string
	// Description of the program.
	Usage string
	// Text to override the USAGE section of help
	UsageText string
	// Description of the program argument format.
	ArgsUsage string
	// Version of the program
	Version string
	// Description of the program
	Description string
	// DefaultCommand is the (optional) name of a command
	// to run if no command names are passed as CLI arguments.
	DefaultCommand string
	// List of commands to execute
	Commands []*Command
	// List of flags to parse
	Flags []*Flag

	Injects []*Inject
	// Boolean to enable bash completion commands
	EnableBashCompletion bool
	// Boolean to hide built-in help command and help flag
	HideHelp bool
	// Boolean to hide built-in help command but keep help flag.
	// Ignored if HideHelp is true.
	HideHelpCommand bool
	// Boolean to hide built-in version flag and the VERSION section of help
	HideVersion bool
	// categories contains the categorized commands and is populated on app startup
	Action Action

	Before       Action
	GlobalBefore Action
	After        Action
	GlobalAfter  Action

	Compiled time.Time
	// List of all authors who contributed
	Authors []*cli.Author
	// Copyright of the binary if any
	Copyright string

	CustomAppHelpTemplate string
	// Boolean to enable short-option handling so user can combine several
	// single-character bool arguments into one
	// i.e. foobar -o -v -> foobar -ov
	UseShortOptionHandling bool
	// Enable suggestions for commands and flags
	Suggest bool

	Command *cli.Command

	Output Output `default:"json"`

	Geninject string `default:"appinject.go"`

	globalFlags map[node][]*Flag

	nodeTree map[node]node
}

type node interface {
	// String() string
}

type Command struct {
	// The name of the command
	Name string
	// A list of aliases for the command
	Aliases []string
	// A short description of the usage of this command
	Usage string
	// Custom text to show on USAGE section of help
	UsageText string
	// A longer explanation of how the command works
	Description string
	// A short description of the arguments of this command
	ArgsUsage string
	// The category the command is part of
	Category string
	// The function to call when checking for bash command completions
	// List of child commands
	Subcommands []*Command
	// List of flags to parse
	Flags []*Flag
	// Treat all flags as normal arguments if true
	SkipFlagParsing bool
	// Boolean to hide built-in help command and help flag
	HideHelp bool
	// Boolean to hide built-in help command but keep help flag
	// Ignored if HideHelp is true.
	HideHelpCommand bool
	// Boolean to hide this command from help or completion
	Hidden bool
	// Boolean to enable short-option handling so user can combine several
	// single-character bool arguments into one
	// i.e. foobar -o -v -> foobar -ov
	UseShortOptionHandling bool

	// Full name of command for help, defaults to full command name, including parent commands.
	HelpName string

	Action Action

	Before Action

	After Action
	// CustomHelpTemplate the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	CustomHelpTemplate string
}

func (cmd *Command) String() string {
	return fmt.Sprintf("Command %s", cmd.Name)
}

type Map = map[string]interface{}

func (app *appImpl) Run(args []string) error {
	var cliApp = cli.App{}
	app.globalFlags = make(map[node][]*Flag)
	app.nodeTree = make(map[node]node)

	if err := copier.Copy(&cliApp, app); err != nil {
		return err
	}

	if err := app.traverseFlagsCopy(&cliApp, app); err != nil {
		return err
	}

	if err := app.traverseBuildActions(&cliApp); err != nil {
		return err
	}

	return cliApp.Run(args)
}

func (app *appImpl) traverseBuildActions(cliapp *cli.App) error {
	var (
		buildAction   = app.makeSetupAction()
		compileAction = app.makeCompileAction()
	)

	cliapp.Action = buildAction(app.Action, app.buildFlagEnv(app.Flags, app), true)
	cliapp.Before = compileAction(app.Before, app.buildFlagEnv(app.Flags, app), false)
	cliapp.After = compileAction(app.After, app.buildFlagEnv(app.Flags, app), false)

	for i, cmd := range app.Commands {
		if err := app.traverseBuildCmdActions(cliapp.Commands[i], cmd); err != nil {
			return err
		}
	}

	return nil
}

func (app *appImpl) makeSetupAction() func(action Action, env Map, out bool) ActionFunc {
	return func(action Action, env Map, out bool) ActionFunc {
		if action == "" {
			return nil
		}

		return func(ctx *cli.Context) error {

			var result any = nil
			if app.GlobalBefore != "" {
				beforeProg, err := app.GlobalBefore.Compile(env)
				if err != nil {
					return err
				}

				result, err = beforeProg.Run(buildCtxEnv(ctx))
				if err != nil {
					return err
				}
			}
			_ = result
			prog, err := action.Compile(env)
			if err != nil {
				return err
			}

			output, err := prog.Run(buildCtxEnv(ctx))
			if err != nil {
				log.Printf("run action error %s", err)
				return err
			}
			if out {
				return app.writeOutput(output)
			}
			return nil
		}
	}
}

func resultCtx(result interface{}) Map {
	return Map{"$_": result}
}

func (app *appImpl) makeCompileAction() func(action Action, env Map, out bool) ActionFunc {
	return func(action Action, env Map, out bool) ActionFunc {
		if action == "" {
			return nil
		}

		prog, err := action.Compile(env)
		if err != nil {
			log.Fatalf("compile program error %s", err)
		}

		return func(ctx *cli.Context) error {
			output, err := prog.Run(buildCtxEnv(ctx))
			if err != nil {
				return err
			}
			if out {
				return app.writeOutput(output)
			}
			return nil
		}
	}
}

func (app *appImpl) traverseBuildCmdActions(dst *cli.Command, src *Command) error {
	for i, subcmd := range src.Subcommands {
		if err := app.traverseBuildCmdActions(dst.Subcommands[i], subcmd); err != nil {
			return err
		}
	}

	var (
		buildAction   = app.makeSetupAction()
		compileAction = app.makeCompileAction()
	)

	dst.Before = compileAction(src.Before, app.buildFlagEnv(src.Flags, src), false)
	dst.Action = buildAction(src.Action, app.buildFlagEnv(src.Flags, src), true)
	dst.After = compileAction(src.After, app.buildFlagEnv(src.Flags, src), false)
	return nil
}

func (app *appImpl) traverseFlagsCopy(dst *cli.App, src *appImpl) error {
	for i, flag := range src.Flags {
		dst.Flags[i] = flag.Flag
	}

	app.globalFlags[app] = src.Flags
	for i, cmd := range src.Commands {
		_ = app.traverseCmd(dst.Commands[i], cmd)
		app.nodeTree[cmd] = app
	}

	return nil
}

func (app *appImpl) traverseCmd(dst *cli.Command, src *Command) error {
	_ = app.traverseFlags(dst, src)

	for i, subcmd := range src.Subcommands {
		_ = app.traverseCmd(dst.Subcommands[i], subcmd)
		app.nodeTree[subcmd] = src

	}
	return nil
}

func (app *appImpl) traverseFlags(dst *cli.Command, src *Command) error {
	for i, flag := range src.Flags {
		dst.Flags[i] = flag.Flag
	}

	app.globalFlags[src] = src.Flags
	return nil
}

func (app *appImpl) buildFlagEnv(flags []*Flag, curr node) Map {
	var (
		env   = make(Map)
		nodes []node
	)

	//  buildFlagEnv(flags)
	for parent := app.nodeTree[curr]; parent != nil; parent = app.nodeTree[parent] {
		nodes = append([]node{parent}, nodes...)
	}

	for _, node := range nodes {
		_ = mergo.MapWithOverwrite(&env, buildFlagEnv(app.globalFlags[node]))
	}

	_ = mergo.MapWithOverwrite(&env, buildFlagEnv(flags))
	return env
}

func (app *appImpl) String() string {
	return fmt.Sprintf("App %s", app.Name)
}
