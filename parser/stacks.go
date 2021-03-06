package parser

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/cnupp/cli/cmd"
	cli "gopkg.in/urfave/cli.v2"
)

func StacksCommand() *cli.Command {
	return &cli.Command{
		Name:  "stacks",
		Usage: "Stacks Commands",
		Subcommands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "List all Stacks",
				ArgsUsage: " ",
				Action: func(c *cli.Context) error {
					err := cmd.StacksList()
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "info",
				Usage:     "Get info of a Stack",
				ArgsUsage: "<stack-name>",
				Action: func(c *cli.Context) error {
					err := cmd.GetStack(c.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "create",
				Usage:     "Create a new Stack",
				ArgsUsage: "<stack-file>",
				Action: func(c *cli.Context) error {
					err := cmd.StackCreate(c.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "update",
				Usage:     "Update an existing Stack",
				ArgsUsage: "<stack-id> <stack-file>",
				Action: func(c *cli.Context) error {
					err := cmd.StackUpdate(c.Args().First(), c.Args().Get(1))
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "remove",
				Usage:     "Delete a Stack",
				ArgsUsage: "<stack-name>",
				Action: func(c *cli.Context) error {
					err := cmd.StackRemove(c.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "publish",
				Usage:     "Publish a Stack",
				ArgsUsage: "<stack-id>",
				Action: func(c *cli.Context) error {
					err := cmd.StackPublish(c.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil
				},
			},
			{
				Name:      "unpublish",
				Usage:     "Unpublish a Stack",
				ArgsUsage: "<stack-id>",
				Action: func(c *cli.Context) error {
					err := cmd.StackUnPublish(c.Args().First())
					if err != nil {
						return cli.Exit(fmt.Sprintf("%v", err), 1)
					}
					return nil

				},
			},
		},
	}
}

func Stacks(argv []string) error {
	usage := `
Valid commands for apps:

stacks:create        create a new stack
stacks:list          list accessible stacks
stacks:info          view info about a stack
stacks:remove        remove an existing stack
stacks:update        update stack
stacks:publish       publish stack
stacks:unpublish     unpublish stack

Use 'cde help [command]' to learn more.
`
	switch argv[0] {
	case "stacks:create":
		return stackCreate(argv)
	case "stacks:list":
		return stackList()
	case "stacks:info":
		return stackInfo(argv)
	case "stacks:remove":
		return stackRemove(argv)
	case "stacks:update":
		return stackUpdate(argv)
	case "stacks:publish":
		return stackPublish(argv)
	case "stacks:unpublish":
		return stackUnPublish(argv)
	default:
		if printHelp(argv, usage) {
			return nil
		}

		if argv[0] == "stacks" {
			argv[0] = "stacks:list"
			return stackList()
		}

		PrintUsage()
		return nil
	}

}

func stackCreate(argv []string) error {
	usage := `
Create a stack.

Usage: cde stacks:create <stackfile>

Arguments:
  <stackfile>
    the stack file.
`

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.StackCreate(safeGetValue(args, "<stackfile>"))
}

func stackList() error {
	return cmd.StacksList()
}

func stackInfo(argv []string) error {
	usage := `
View info about a stack

Usage: cde stacks:info <stack-name>

Arguments:
  <stack-name>
    a stack name.
`
	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	stackName := safeGetValue(args, "<stack-name>")

	return cmd.GetStack(stackName)
}

func stackRemove(argv []string) error {
	usage := `
Remove an existing stack.

Usage: cde stacks:remove <stack>

Arguments:
  <stack>
    the stack name.
`

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.StackRemove(safeGetValue(args, "<stack>"))
}

func stackUpdate(argv []string) error {
	usage := `
Update a stack

Usage: cde stacks:update <stack-id> <stackfile>

Arguments:
  <stack-id>
    the stack id.
  <stackfile>
    this stackfile.
`

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.StackUpdate(safeGetValue(args, "<stack-id>"), safeGetValue(args, "<stackfile>"))
}

func stackPublish(argv []string) error {
	usage := `
Update a stack

Usage: cde stacks:publish <stack-id>

Arguments:
  <stack-id>
    the stack id.
`

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.StackPublish(safeGetValue(args, "<stack-id>"))
}

func stackUnPublish(argv []string) error {
	usage := `
Update a stack

Usage: cde stacks:unpublish <stack-id>

Arguments:
  <stack-id>
    the stack id.
`

	args, err := docopt.Parse(usage, argv, true, "", false, true)

	if err != nil {
		return err
	}

	return cmd.StackUnPublish(safeGetValue(args, "<stack-id>"))
}
