package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/k-kanke/code-stash-cli/internal/api"
	"github.com/k-kanke/code-stash-cli/internal/auth"
	"github.com/k-kanke/code-stash-cli/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with CodeStash via device authorization",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg.APIBaseURL, cfg.ClientID, cfg.ClientSecret)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		deviceResp, err := client.StartDeviceCode(ctx)
		if err != nil {
			return fmt.Errorf("failed to request device code: %w", err)
		}

		printInstructions(cmd, deviceResp)

		interval := time.Duration(deviceResp.Interval) * time.Second
		if interval <= 0 {
			interval = 5 * time.Second
		}
		expiration := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

		for {
			if time.Now().After(expiration) {
				return errors.New("device code expired, please retry login")
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}

			tokenResp, apiErr, err := client.ExchangeDeviceCode(ctx, deviceResp.DeviceCode)
			if err != nil {
				return fmt.Errorf("token exchange failed: %w", err)
			}
			if apiErr != nil {
				switch apiErr.Code {
				case "authorization_pending":
					cmd.Print(".")
					continue
				case "slow_down":
					interval += time.Second
					cmd.Printf("\nServer asked to slow down, next attempt in %s\n", interval)
					continue
				case "expired_token":
					return errors.New("device code expired, run login again")
				case "access_denied":
					return errors.New("authorization denied in the browser")
				default:
					return fmt.Errorf("token exchange error: %s", apiErr.Code)
				}
			}

			scope := strings.Fields(tokenResp.Scope)
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			var refresh string
			if tokenResp.RefreshToken != nil {
				refresh = *tokenResp.RefreshToken
			}

			if err := auth.SaveToken(cfg.TokenPath, auth.Token{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: refresh,
				Scope:        scope,
				ExpiresAt:    expiresAt,
			}); err != nil {
				return fmt.Errorf("save token: %w", err)
			}

			cmd.Printf("\nLogin successful! Token saved to %s\n", cfg.TokenPath)
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func printInstructions(cmd *cobra.Command, resp *api.DeviceCodeResponse) {
	cmd.Println("To authorize this CLI:")
	target := strings.TrimSpace(resp.VerificationURI)
	if target == "" {
		target = strings.TrimSpace(resp.VerificationURIComplete)
	}
	cmd.Printf("1. Open: %s\n", target)
	cmd.Printf("2. Enter code: %s\n", formatUserCode(resp.UserCode))
	cmd.Println("Waiting for authorization...")
}

func formatUserCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return code
	}
	return code[:3] + "-" + code[3:]
}
