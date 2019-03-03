// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/railgun/railgun/util"
	"github.com/spatialcurrent/viper"
	"gopkg.in/yaml.v2"
)

var GO_DFL_DEFAULT_QUOTES = []string{"'", "\"", "`"}

func evalFunction(cmd *cobra.Command, args []string) {

	v := viper.New()
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	start := time.Now()

	verbose := v.GetBool("verbose")
	inputExpression := v.GetString("expression")
	inputFile := v.GetString("file")
	loadContextFromEnvironment := v.GetBool("env")
	dryRun := v.GetBool("dry-run")
	pretty := v.GetBool("pretty")
	sql := v.GetBool("sql")

	if len(inputExpression) == 0 && len(inputFile) == 0 {
		err := cmd.Usage()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	var root dfl.Node
	if len(inputFile) > 0 {
		f, _, err := grw.ReadFromResource(inputFile, "none", 4096, false, nil)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error opening dfl file"))
			os.Exit(1)
		}
		content, err := f.ReadAll()
		if err != nil {
			fmt.Println(errors.Wrap(err, "error reading all from dfl file"))
			os.Exit(1)
		}
		inputExpression = strings.TrimSpace(dfl.RemoveComments(string(content)))
	}

	if len(inputExpression) > 0 {
		n, _, err := dfl.Parse(inputExpression)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error parsing dfl expression"))
			os.Exit(1)
		}
		root = n
	}

	ctx := map[string]interface{}{}

	if loadContextFromEnvironment {
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			ctx[strings.TrimSpace(pair[0])] = dfl.TryConvertString(strings.TrimSpace(pair[1]))
		}
	}

	funcs := dfl.NewFuntionMapWithDefaults()

	for _, a := range args {
		if !strings.Contains(a, "=") {
			fmt.Println(errors.New("Context attribute \"" + a + "\" does not contain \"=\"."))
			os.Exit(1)
		}
		pair := strings.SplitN(a, "=", 2)
		value, err := dfl.ParseCompile(strings.TrimSpace(pair[1]))
		if err != nil {
			fmt.Println(errors.Wrap(err, "Could not parse context variable"))
			os.Exit(1)
		}
		switch value.(type) {
		case dfl.Array:
			_, arr, err := value.(dfl.Array).Evaluate(map[string]interface{}{}, map[string]interface{}{}, funcs, GO_DFL_DEFAULT_QUOTES[1:])
			if err != nil {
				fmt.Println(errors.Wrap(err, "error evaluating context expression for "+strings.TrimSpace(pair[0])))
				os.Exit(1)
			}
			ctx[strings.TrimSpace(pair[0])] = arr
		case dfl.Set:
			_, arr, err := value.(dfl.Set).Evaluate(map[string]interface{}{}, map[string]interface{}{}, funcs, GO_DFL_DEFAULT_QUOTES[1:])
			if err != nil {
				fmt.Println(errors.Wrap(err, "error evaluating context expression for "+strings.TrimSpace(pair[0])))
				os.Exit(1)
			}
			ctx[strings.TrimSpace(pair[0])] = arr
		case dfl.Literal:
			ctx[strings.TrimSpace(pair[0])] = value.(dfl.Literal).Value
		case *dfl.Literal:
			ctx[strings.TrimSpace(pair[0])] = value.(*dfl.Literal).Value
		default:
			ctx[strings.TrimSpace(pair[0])] = dfl.TryConvertString(pair[1])
		}
	}

	if pretty {
		fmt.Println("# Pretty Version \n" + root.Dfl(dfl.DefaultQuotes[1:], true, 0) + "\n")
	}

	if sql {
		fmt.Println("# SQL Version \n" + root.Sql(pretty, 0) + "\n")
	}

	if verbose {

		fmt.Println("******************* Context *******************")
		out, err := yaml.Marshal(ctx)
		if err != nil {
			fmt.Println("Error marshaling context to yaml.")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))

		fmt.Println("******************* Parsed *******************")
		out, err = yaml.Marshal(root.Map())
		if err != nil {
			fmt.Println("Error marshaling expression to yaml.")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))

	}

	root = root.Compile()

	if verbose {
		fmt.Println("******************* Compiled *******************")
		out, err := yaml.Marshal(root.Map())
		if err != nil {
			fmt.Println("Error marshaling expression to yaml.")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("# YAML Version\n" + string(out))
		fmt.Println("# DFL Version\n" + GO_DFL_DEFAULT_QUOTES[0] + root.Dfl(GO_DFL_DEFAULT_QUOTES[1:], false, 0) + GO_DFL_DEFAULT_QUOTES[0])
		if sql {
			fmt.Println("# SQL Version\n" + root.Sql(pretty, 0) + "\n")
		}
	}

	if dryRun {
		os.Exit(0)
	}

	_, result, err := root.Evaluate(map[string]interface{}{}, ctx, funcs, GO_DFL_DEFAULT_QUOTES[1:])
	if err != nil {
		fmt.Println(errors.Wrap(err, "error evaluating expression"))
		os.Exit(1)
	}

	switch result.(type) {
	case bool:
		result_bool := result.(bool)
		if verbose {
			fmt.Println("******************* Result *******************")
			fmt.Println(dfl.TryFormatLiteral(result, GO_DFL_DEFAULT_QUOTES[1:], false, 0))
			elapsed := time.Since(start)
			fmt.Println("Done in " + elapsed.String())
		}
		if result_bool {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	default:
		if verbose {
			fmt.Println("******************* Result *******************")
			fmt.Println("Type:", reflect.TypeOf(result))
			fmt.Println("Value:", dfl.TryFormatLiteral(result, GO_DFL_DEFAULT_QUOTES[1:], false, 0))
			elapsed := time.Since(start)
			fmt.Println("Done in " + elapsed.String())
		}
	}

}

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "evaluate a DFL expression using go-dfl",
	Long:  "Evaluate a DFL expression using go-dfl.",
	Run:   evalFunction,
}

func init() {
	dflCmd.AddCommand(evalCmd)

	evalCmd.Flags().StringP("expression", "e", "", "evaluate dfl expression")
	evalCmd.Flags().StringP("file", "f", "", "evaluate dfl file")
	evalCmd.Flags().BoolP("env", "", false, "load environment variables into context")
	evalCmd.Flags().BoolP("dry-run", "d", false, "parse and compile expression, but do not evaluate against context")
	evalCmd.Flags().BoolP("pretty", "p", false, "print pretty output")
	evalCmd.Flags().BoolP("sql", "s", false, "print SQL versio of expression to output")

}
