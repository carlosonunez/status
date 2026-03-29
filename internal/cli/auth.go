package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	internalconfig "github.com/carlosonunez/status/internal/config"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/carlosonunez/status/internal/tokenstore"
	"github.com/spf13/cobra"
)

func newAuthCommand(gr *getter.Registry, sr *setter.Registry) *cobra.Command {
	return newAuthCommandWithStore(gr, sr, nil)
}

func newAuthCommandWithStore(gr *getter.Registry, sr *setter.Registry, store pluginspec.TokenStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication for integrations",
	}
	cmd.AddCommand(newAuthLoginCommand(gr, sr, store))
	return cmd
}

func newAuthLoginCommand(gr *getter.Registry, sr *setter.Registry, store pluginspec.TokenStore) *cobra.Command {
	var all bool
	var integrations string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate into one or more integrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := store
			if s == nil {
				path := cfgPath
				if path == "" {
					path = internalconfig.DefaultPath()
				}
				cfg, err := internalconfig.Load(path)
				if err != nil {
					if !errors.Is(err, internalconfig.ErrNotFound) {
						return fmt.Errorf("load config: %w", err)
					}
					// No config file yet; fall back to the default file-backed store.
					ts, tsErr := tokenstore.FromConfig(internalconfig.TokenStoreConfig{})
					if tsErr != nil {
						return fmt.Errorf("token store: %w", tsErr)
					}
					s = ts
				} else {
					ts, tsErr := tokenstore.FromConfig(cfg.TokenStore)
					if tsErr != nil {
						return fmt.Errorf("token store: %w", tsErr)
					}
					s = ts
				}
			}
			// Default to authenticating all integrations when no selector is given.
			if !all && integrations == "" {
				all = true
			}
			return runAuthLogin(cmd, gr, sr, s, all, integrations)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Authenticate into all integrations")
	cmd.Flags().StringVar(&integrations, "integrations", "", "Comma-separated integration names to authenticate")
	return cmd
}

func runAuthLogin(
	cmd *cobra.Command,
	gr *getter.Registry,
	sr *setter.Registry,
	store pluginspec.TokenStore,
	all bool,
	integrations string,
) error {
	ctx := context.Background()

	authenticators, err := collectAuthenticators(gr, sr, all, integrations)
	if err != nil {
		return err
	}

	if len(authenticators) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No integrations require authentication.")
		return nil
	}

	for name, auth := range authenticators {
		fmt.Fprintf(cmd.OutOrStdout(), "➡️  Authenticating %q...\n", name)
		if err := auth.Authenticate(ctx, store); err != nil {
			return fmt.Errorf("auth %q: %w", name, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✅ Authenticated %q.\n", name)
	}
	return nil
}

// collectAuthenticators gathers the integrations to authenticate.
// If all is true, returns every getter and setter that implements Authenticator.
// If integrations is non-empty, looks up each name and errors if not found or
// if the named integration does not implement Authenticator.
func collectAuthenticators(
	gr *getter.Registry,
	sr *setter.Registry,
	all bool,
	integrations string,
) (map[string]pluginspec.Authenticator, error) {
	result := make(map[string]pluginspec.Authenticator)

	if all {
		for _, g := range gr.All() {
			if a, ok := g.(pluginspec.Authenticator); ok {
				result[g.Name()] = a
			}
		}
		for _, s := range sr.All() {
			if a, ok := s.(pluginspec.Authenticator); ok {
				result[s.Name()] = a
			}
		}
		return result, nil
	}

	for _, name := range strings.Split(integrations, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if g, err := gr.Get(name); err == nil {
			if a, ok := g.(pluginspec.Authenticator); ok {
				result[name] = a
				continue
			}
		}
		if s, err := sr.Get(name); err == nil {
			if a, ok := s.(pluginspec.Authenticator); ok {
				result[name] = a
				continue
			}
		}
		return nil, fmt.Errorf("integration %q not found or does not support authentication", name)
	}
	return result, nil
}
