// Package discord wraps the discordgo session used by disc.
package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Permissions requested when generating invite links.
const invitePermissions int64 = int64(discordgo.PermissionManageChannels) |
	int64(discordgo.PermissionManageRoles) |
	int64(discordgo.PermissionViewChannel) |
	int64(discordgo.PermissionSendMessages) |
	int64(discordgo.PermissionReadMessageHistory) |
	int64(discordgo.PermissionCreateEvents)

// Client wraps a discordgo session.
type Client struct {
	s *discordgo.Session
}

// New creates a client from a bot token and opens a websocket connection.
func New(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("no bot token configured")
	}

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.Open(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	return &Client{s: s}, nil
}

// Close closes the websocket connection.
func (c *Client) Close() {
	_ = c.s.Close()
}

// Session exposes the underlying discordgo session.
func (c *Client) Session() *discordgo.Session {
	return c.s
}

// User returns the connected bot user.
func (c *Client) User() *discordgo.User {
	return c.s.State.User
}

// Guilds returns the guilds the bot is a member of.
func (c *Client) Guilds() []*discordgo.Guild {
	return c.s.State.Guilds
}

// InviteLink returns an OAuth2 invite URL with the requested permissions.
func InviteLink(clientID string) string {
	return fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&scope=bot&permissions=%d",
		clientID, invitePermissions,
	)
}

// WaitReady blocks until the websocket is ready or a timeout occurs.
func (c *Client) WaitReady() error {
	return c.waitReady(15 * time.Second)
}

func (c *Client) waitReady(timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for Discord connection")
		case <-ticker.C:
			if c.s.State.User != nil && c.s.State.User.ID != "" {
				return nil
			}
		}
	}
}
