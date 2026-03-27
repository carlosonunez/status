package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// integrationEntry is the wire type used for JSON/YAML serialisation.
type integrationEntry struct {
	Type       string   `json:"type" yaml:"type"`
	Name       string   `json:"name" yaml:"name"`
	Source     string   `json:"source" yaml:"source"`
	Parameters []string `json:"parameters" yaml:"parameters"`
}

// integrationDetail is the wire type used for "integration show" JSON/YAML.
type integrationDetail struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string            `json:"type" yaml:"type"`
	Source      string            `json:"source" yaml:"source"`
	Parameters  []paramDetail     `json:"parameters" yaml:"parameters"`
	AuthParams  []authParamDetail `json:"auth_parameters,omitempty" yaml:"auth_parameters,omitempty"`
	AuthNotes   []string          `json:"auth_notes,omitempty" yaml:"auth_notes,omitempty"`
}

type paramDetail struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Required    bool   `json:"required" yaml:"required"`
}

type authParamDetail struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool   `json:"required" yaml:"required"`
	EnvVar      string `json:"env_var,omitempty" yaml:"env_var,omitempty"`
}

func newIntegrationCommand(gr *getter.Registry, sr *setter.Registry) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "Browse available integrations",
	}
	cmd.AddCommand(newIntegrationListCommand(gr, sr))
	cmd.AddCommand(newIntegrationShowCommand(gr, sr))
	return cmd
}

// ── integration list ──────────────────────────────────────────────────────────

func newIntegrationListCommand(gr *getter.Registry, sr *setter.Registry) *cobra.Command {
	var typeFilter, format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available integrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries := buildListEntries(gr, sr, typeFilter)
			return writeListOutput(cmd.OutOrStdout(), entries, format)
		},
	}
	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by type: event-getter or status-setter")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, or yaml")
	return cmd
}

func buildListEntries(gr *getter.Registry, sr *setter.Registry, typeFilter string) []integrationEntry {
	var entries []integrationEntry
	if typeFilter == "" || typeFilter == "event-getter" {
		for _, g := range gr.All() {
			entries = append(entries, integrationEntry{
				Type:       "event-getter",
				Name:       g.Name(),
				Source:     sourceOf(g),
				Parameters: paramNames(g.ParamSpecs()),
			})
		}
	}
	if typeFilter == "" || typeFilter == "status-setter" {
		for _, s := range sr.All() {
			entries = append(entries, integrationEntry{
				Type:       "status-setter",
				Name:       s.Name(),
				Source:     sourceOf(s),
				Parameters: paramNames(s.ParamSpecs()),
			})
		}
	}
	return entries
}

func writeListOutput(w io.Writer, entries []integrationEntry, format string) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	case "yaml":
		return yaml.NewEncoder(w).Encode(entries)
	default:
		return writeListTable(w, entries)
	}
}

func writeListTable(w io.Writer, entries []integrationEntry) error {
	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tNAME\tSOURCE\tPARAMETERS")
	for _, e := range entries {
		if len(e.Parameters) == 0 {
			fmt.Fprintf(tw, "%s\t%s\t%s\t(none)\n", e.Type, e.Name, e.Source)
			continue
		}
		for i, p := range e.Parameters {
			if i == 0 {
				fmt.Fprintf(tw, "%s\t%s\t%s\t- %s\n", e.Type, e.Name, e.Source, p)
			} else {
				fmt.Fprintf(tw, "\t\t\t- %s\n", p)
			}
		}
	}
	return tw.Flush()
}

// ── integration show ──────────────────────────────────────────────────────────

func newIntegrationShowCommand(gr *getter.Registry, sr *setter.Registry) *cobra.Command {
	var names, format string
	var all, paginate bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show detailed information about one or more integrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			details, err := buildShowDetails(gr, sr, names, all)
			if err != nil {
				return err
			}
			var sb strings.Builder
			if err := writeShowOutput(&sb, details, format); err != nil {
				return err
			}
			return pageOrPrint(cmd.OutOrStdout(), sb.String(), paginate)
		},
	}
	cmd.Flags().StringVar(&names, "name", "", "Comma-separated integration names to show")
	cmd.Flags().BoolVar(&all, "all", false, "Show all integrations")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, or yaml")
	cmd.Flags().BoolVar(&paginate, "paginate", false, "Pipe output through a pager (less)")
	return cmd
}

func buildShowDetails(gr *getter.Registry, sr *setter.Registry, names string, all bool) ([]integrationDetail, error) {
	if all {
		var details []integrationDetail
		for _, g := range gr.All() {
			details = append(details, detailFromGetter(g))
		}
		for _, s := range sr.All() {
			details = append(details, detailFromSetter(s))
		}
		return details, nil
	}
	if names == "" {
		return nil, fmt.Errorf("provide --name or --all")
	}
	var details []integrationDetail
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if g, err := gr.Get(name); err == nil {
			details = append(details, detailFromGetter(g))
			continue
		}
		if s, err := sr.Get(name); err == nil {
			details = append(details, detailFromSetter(s))
			continue
		}
		return nil, fmt.Errorf("integration %q not found", name)
	}
	return details, nil
}

func detailFromGetter(g getter.EventGetter) integrationDetail {
	d := integrationDetail{
		Name:       g.Name(),
		Type:       "event-getter",
		Source:     sourceOf(g),
		Parameters: toParamDetails(g.ParamSpecs()),
	}
	if desc, ok := g.(pluginspec.Described); ok {
		d.Description = desc.Description()
	}
	if auth, ok := g.(pluginspec.Authenticatable); ok {
		d.AuthParams = toAuthParamDetails(auth.AuthParams())
		d.AuthNotes = auth.AuthNotes()
	}
	return d
}

func detailFromSetter(s setter.StatusSetter) integrationDetail {
	d := integrationDetail{
		Name:       s.Name(),
		Type:       "status-setter",
		Source:     sourceOf(s),
		Parameters: toParamDetails(s.ParamSpecs()),
	}
	if desc, ok := s.(pluginspec.Described); ok {
		d.Description = desc.Description()
	}
	if auth, ok := s.(pluginspec.Authenticatable); ok {
		d.AuthParams = toAuthParamDetails(auth.AuthParams())
		d.AuthNotes = auth.AuthNotes()
	}
	return d
}

func writeShowOutput(w io.Writer, details []integrationDetail, format string) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(details)
	case "yaml":
		return yaml.NewEncoder(w).Encode(details)
	default:
		for i, d := range details {
			if i > 0 {
				fmt.Fprintln(w, "---")
			}
			writeDetailText(w, d)
		}
		return nil
	}
}

func writeDetailText(w io.Writer, d integrationDetail) {
	fmt.Fprintf(w, "NAME: %s\n", d.Name)
	if d.Description != "" {
		fmt.Fprintf(w, "DESCRIPTION: %s\n", d.Description)
	}
	fmt.Fprintf(w, "TYPE: %s\n", d.Type)
	fmt.Fprintf(w, "SOURCE: %s\n", d.Source)
	fmt.Fprintln(w, "PARAMETERS")
	if len(d.Parameters) == 0 {
		fmt.Fprintln(w, "(none)")
	} else {
		for _, p := range d.Parameters {
			fmt.Fprintf(w, "\n- NAME:        %s\n", p.Name)
			if p.Description != "" {
				fmt.Fprintf(w, "  DESCRIPTION: %s\n", p.Description)
			}
			if p.Type != "" {
				fmt.Fprintf(w, "  TYPE:        %s\n", p.Type)
			}
			fmt.Fprintf(w, "  REQUIRED:    %v\n", p.Required)
		}
	}
	if len(d.AuthParams) > 0 {
		fmt.Fprintln(w, "\nAUTHENTICATION PARAMETERS")
		for _, ap := range d.AuthParams {
			fmt.Fprintf(w, "\n- NAME: %s\n", ap.Name)
			if ap.Description != "" {
				fmt.Fprintf(w, "  DESCRIPTION: %s\n", ap.Description)
			}
			fmt.Fprintf(w, "  REQUIRED: %v\n", ap.Required)
			if ap.EnvVar != "" {
				fmt.Fprintf(w, "  ENVIRONMENT VARIABLE: %s\n", ap.EnvVar)
			}
		}
	}
	if len(d.AuthNotes) > 0 {
		fmt.Fprintln(w, "\nAUTHENTICATION NOTES")
		for _, note := range d.AuthNotes {
			fmt.Fprintf(w, "\n- %s\n", note)
		}
	}
}

func pageOrPrint(w io.Writer, content string, paginate bool) error {
	if !paginate {
		fmt.Fprint(w, content)
		return nil
	}
	pager := exec.Command("less", "-R")
	pager.Stdin = strings.NewReader(content)
	pager.Stdout = w
	if err := pager.Run(); err != nil {
		// Fall back to direct output if less is unavailable.
		fmt.Fprint(w, content)
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sourceOf(v any) string {
	if s, ok := v.(pluginspec.Sourced); ok {
		if src := s.Source(); src != "" {
			return src
		}
	}
	return "builtin"
}

func paramNames(specs []pluginspec.ParamSpec) []string {
	names := make([]string, len(specs))
	for i, s := range specs {
		names[i] = s.Name
	}
	return names
}

func toParamDetails(specs []pluginspec.ParamSpec) []paramDetail {
	out := make([]paramDetail, len(specs))
	for i, s := range specs {
		out[i] = paramDetail{
			Name:        s.Name,
			Description: s.Description,
			Type:        s.Type,
			Required:    s.Required,
		}
	}
	return out
}

func toAuthParamDetails(aps []pluginspec.AuthParam) []authParamDetail {
	out := make([]authParamDetail, len(aps))
	for i, ap := range aps {
		out[i] = authParamDetail{
			Name:        ap.Name,
			Description: ap.Description,
			Required:    ap.Required,
			EnvVar:      ap.EnvVar,
		}
	}
	return out
}
