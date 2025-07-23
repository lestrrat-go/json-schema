package main

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	schema "github.com/lestrrat-go/json-schema"
	"github.com/lestrrat-go/json-schema/validator"
)

func main() {
	app := &cli.Command{
		Name:  "json-schema",
		Usage: "JSON Schema validation and code generation tool",
		Commands: []*cli.Command{
			{
				Name:      "lint",
				Usage:     "report formatting errors found in schema file",
				ArgsUsage: "[filename]",
				Action:    lintCommand,
			},
			{
				Name:      "gen-validator",
				Usage:     "create a pre-compiled validator code from schema file",
				ArgsUsage: "[filename]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Value: "val",
						Usage: "assign the resulting validator to this variable name",
					},
				},
				Action: genValidatorCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func lintCommand(ctx context.Context, c *cli.Command) error {
	filename := c.Args().First()
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	// Read the schema file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse the JSON schema
	var s schema.Schema
	if err := s.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	// Try to compile the validator to check for semantic errors
	_, err = validator.Compile(context.Background(), &s)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	fmt.Printf("Schema %s is valid\n", filename)
	return nil
}

func genValidatorCommand(ctx context.Context, c *cli.Command) error {
	filename := c.Args().First()
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	validatorName := c.String("name")
	if validatorName == "" {
		validatorName = "val"
	}

	// Read the schema file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse the JSON schema
	var s schema.Schema
	if err := s.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("failed to parse JSON schema: %w", err)
	}

	// Compile the validator
	v, err := validator.Compile(context.Background(), &s)
	if err != nil {
		return fmt.Errorf("failed to compile validator: %w", err)
	}

	// Generate the validator code
	generator := validator.NewCodeGenerator()
	
	var buf bytes.Buffer
	if err := generator.Generate(&buf, v); err != nil {
		return fmt.Errorf("failed to generate validator code: %w", err)
	}

	// Output the formatted generated code as a variable assignment
	generatedCode := buf.String()
	// Remove any leading/trailing whitespace and put it on same line as :=
	generatedCode = strings.TrimSpace(generatedCode)
	code := fmt.Sprintf("%s := %s", validatorName, generatedCode)
	
	// Format the code properly using go/format
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// If formatting fails, output the unformatted code
		fmt.Print(code)
	} else {
		fmt.Print(string(formatted))
	}
	return nil
}