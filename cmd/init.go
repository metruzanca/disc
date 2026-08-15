package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/metruzanca/disc/internal/config"
	"github.com/metruzanca/disc/internal/discord"
	"github.com/metruzanca/disc/internal/util"
	"github.com/spf13/cobra"
)

var (
	initFlags = struct {
		token string
	}{}
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up disc with your bot token",
	Long: `Initialize disc by providing your Discord bot token.

The token will be saved to the local .env file for future use.
Pass --token to provide it non-interactively (required with --agent).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readToken()
		if err != nil {
			return err
		}
		if strings.TrimSpace(token) == "" {
			return fmt.Errorf("no token provided")
		}
		// Validate the token by connecting briefly.
		client, err := discord.New(strings.TrimSpace(token))
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
		defer client.Close()

		if err := client.WaitReady(); err != nil {
			return err
		}

		if err := config.SetEnvVar(config.TokenEnv, strings.TrimSpace(token)); err != nil {
			return err
		}

		util.Green.Printf("Token saved to %s.\n", config.EnvFileName)
		fmt.Println()
		fmt.Println("Invite your bot to a server using this link:")
		fmt.Println()
		fmt.Println(discord.InviteLink(client.User().ID))
		return nil
	},
}

func readToken() (string, error) {
	if initFlags.token != "" {
		return strings.TrimSpace(initFlags.token), nil
	}
	if agentMode {
		return "", fmt.Errorf("no token provided; pass --token <token> (the interactive prompt is disabled by --agent)")
	}
	fmt.Println("Enter your Discord bot token:")
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func init() {
	initCmd.Flags().StringVar(&initFlags.token, "token", "", "Bot token (skips the interactive prompt; required with --agent)")
}
