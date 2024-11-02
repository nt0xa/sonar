package actions

import (
	"fmt"
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

func completeDNSRecord(acts *Actions) completionFunc {
	return func(
		cmd *cobra.Command,
		args []string,
		_ string,
	) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveError
		}

		payload, err := cmd.Flags().GetString("payload")
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		records, err := (*acts).DNSRecordsList(cmd.Context(), DNSRecordsListParams{
			PayloadName: payload,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, len(records))

		for i, r := range records {
			completions[i] = fmt.Sprintf(
				"%d\t%s.%s %d IN %s %s",
				r.Index,
				r.Name,
				r.PayloadSubdomain,
				r.TTL,
				r.Type,
				strings.Join(r.Values, " "),
			)
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeHTTPRoute(acts *Actions) completionFunc {
	return func(
		cmd *cobra.Command,
		args []string,
		_ string,
	) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveError
		}

		payload, err := cmd.Flags().GetString("payload")
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		routes, err := (*acts).HTTPRoutesList(cmd.Context(), HTTPRoutesListParams{
			PayloadName: payload,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, len(routes))

		for i, r := range routes {
			completions[i] = fmt.Sprintf("%d\t%s %s -> %d", r.Index, r.Method, r.Path, r.Code)
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
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
