package actions

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type completionFunc func(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective)

func completePayloadName(acts *Actions) completionFunc {
	return func(
		cmd *cobra.Command,
		_ []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		payloads, err := (*acts).PayloadsList(cmd.Context(), PayloadsListParams{
			Name: toComplete,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		names := make([]string, len(payloads))

		for i, p := range payloads {
			names[i] = p.Name
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeOne(list []string) completionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
		_ string,
	) ([]string, cobra.ShellCompDirective) {
		return list, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeMany(completions []string) completionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		// aaa,bbb,c -> [aaa, bbb, c]
		parts := strings.Split(toComplete, ",")

		// [aaa, bb, c] -> prefix = "aaa,bbb,"
		prefix := strings.Join(parts[:len(parts)-1], ",")
		if prefix != "" {
			prefix += ","
		}

		// lastPart = "c"
		lastPart := parts[len(parts)-1]

		// Filter completions based on the current input
		var result []string
		for _, comp := range completions {
			if slices.Contains(parts, comp) {
				continue
			}
			if strings.HasPrefix(comp, lastPart) {
				result = append(result, prefix+comp)
			}
		}

		return result, cobra.ShellCompDirectiveNoSpace
	}
}
