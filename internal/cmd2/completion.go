package cmd2

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func (c *Command) completePayloadName(
	cmd *cobra.Command, _ []string, toComplete string,
) ([]string, cobra.ShellCompDirective) {
	payloads, err := c.svc.PayloadsList(cmd.Context(), service.PayloadsListInput{Name: toComplete})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := make([]string, len(payloads))
	for i, p := range payloads {
		names[i] = p.Name
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Command) completeDNSRecord(
	cmd *cobra.Command, args []string, _ string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	payload, err := cmd.Flags().GetString("payload")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	records, err := c.svc.DNSRecordsList(cmd.Context(), service.DNSRecordsListInput{PayloadName: payload})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, len(records))
	for i, r := range records {
		completions[i] = fmt.Sprintf(
			"%d\t%s.%s %d IN %s %s",
			r.Index, r.Name, r.PayloadSubdomain, r.TTL, r.Type, strings.Join(r.Values, " "),
		)
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func (c *Command) completeHTTPRoute(
	cmd *cobra.Command, args []string, _ string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	payload, err := cmd.Flags().GetString("payload")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	routes, err := c.svc.HTTPRoutesList(cmd.Context(), service.HTTPRoutesListInput{PayloadName: payload})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, len(routes))
	for i, r := range routes {
		completions[i] = fmt.Sprintf("%d\t%s %s -> %d", r.Index, r.Method, r.Path, r.Code)
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeOne suggests values from a fixed list (single-value flags).
func completeOne(list []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return list, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeMany suggests values from a fixed list for comma-separated flags,
// preserving the already-typed prefix and skipping chosen values.
func completeMany(completions []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		parts := strings.Split(toComplete, ",")

		prefix := strings.Join(parts[:len(parts)-1], ",")
		if prefix != "" {
			prefix += ","
		}

		lastPart := parts[len(parts)-1]

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
