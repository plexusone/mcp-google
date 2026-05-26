// mcp-google is an MCP server for reading Google Slides and Docs.
// It can also be used as a CLI tool for testing and scripting.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/mcp-google/internal/auth"
	"github.com/plexusone/mcp-google/skills/docs"
	"github.com/plexusone/mcp-google/skills/slides"
	runtime "github.com/plexusone/omniskill/mcp/server"
	"github.com/spf13/cobra"

	// Register desktop vault providers (1Password, etc.)
	_ "github.com/plexusone/omnivault-desktop"
)

const (
	serverName    = "mcp-google"
	serverVersion = "v0.5.0"
)

var (
	// Credential flags (persistent across all commands)
	credentialsFile       string
	goauthCredentialsFile string
	goauthCredentialsKey  string
	vaultURI              string
	credentialsName       string

	// Output format flag
	outputFormat string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mcp-google",
	Short: "MCP server and CLI for Google Slides and Docs",
	Long: `An MCP (Model Context Protocol) server for reading Google Slides presentations
and Google Docs documents. Can also be used as a CLI tool for testing and scripting.

Running without a subcommand starts the MCP server (default behavior).

Credentials can be provided via:
  - Google service account JSON file
  - goauth CredentialsSet file
  - Vault-backed credentials (1Password, HashiCorp Vault, etc.)`,
	Example: `  # Start MCP server (default)
  mcp-google --credentials /path/to/service-account.json

  # Explicitly start MCP server
  mcp-google serve --credentials /path/to/service-account.json

  # CLI: Get presentation metadata
  mcp-google get-presentation <id> --credentials /path/to/service-account.json

  # CLI: Get document metadata
  mcp-google get-document-metadata <id> --credentials /path/to/service-account.json

  # CLI: Get document text
  mcp-google get-document-text <id> --credentials /path/to/service-account.json`,
	SilenceUsage: true,
	RunE:         runServer, // Default: run MCP server
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long:  "Start the MCP server using stdio transport for communication with MCP clients.",
	RunE:  runServer,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s\n", serverName, serverVersion)
	},
}

// Slides CLI commands
var getPresentationCmd = &cobra.Command{
	Use:   "get-presentation <presentation-id>",
	Short: "Get presentation metadata",
	Long:  "Get metadata about a Google Slides presentation including title, slide count, locale, and revision ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSlidesTool("get_presentation", map[string]any{
			"presentation_id": args[0],
		})
	},
}

var listSlidesCmd = &cobra.Command{
	Use:   "list-slides <presentation-id>",
	Short: "List all slides in a presentation",
	Long:  "List all slides in a Google Slides presentation with their titles and element counts.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSlidesTool("list_slides", map[string]any{
			"presentation_id": args[0],
		})
	},
}

var getSlideCmd = &cobra.Command{
	Use:   "get-slide <presentation-id>",
	Short: "Get slide content",
	Long:  "Get the content and elements of a specific slide by index or object ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := map[string]any{
			"presentation_id": args[0],
		}
		if slideIndex >= 0 {
			params["slide_index"] = slideIndex
		}
		if slideObjectID != "" {
			params["slide_object_id"] = slideObjectID
		}
		return runSlidesTool("get_slide", params)
	},
}

var getSlideNotesCmd = &cobra.Command{
	Use:   "get-slide-notes <presentation-id>",
	Short: "Get speaker notes for a slide",
	Long:  "Get the speaker notes for a specific slide by index or object ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := map[string]any{
			"presentation_id": args[0],
		}
		if slideIndex >= 0 {
			params["slide_index"] = slideIndex
		}
		if slideObjectID != "" {
			params["slide_object_id"] = slideObjectID
		}
		return runSlidesTool("get_slide_notes", params)
	},
}

var getPresentationContentCmd = &cobra.Command{
	Use:   "get-presentation-content <presentation-id>",
	Short: "Get all slide content",
	Long:  "Get all slide content (text and images) in a single call, ideal for AI analysis.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSlidesTool("get_presentation_content", map[string]any{
			"presentation_id": args[0],
			"include_notes":   includeNotes,
		})
	},
}

// Docs CLI commands
var getDocumentMetadataCmd = &cobra.Command{
	Use:   "get-document-metadata <document-id>",
	Short: "Get document metadata",
	Long:  "Get metadata about a Google Doc including title, word count, and element counts.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsTool("get_document_metadata", map[string]any{
			"document_id": args[0],
		})
	},
}

var getDocumentContentCmd = &cobra.Command{
	Use:   "get-document-content <document-id>",
	Short: "Get document structured content",
	Long:  "Get the full structured content of a document including headings, paragraphs, images, and tables.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsTool("get_document_content", map[string]any{
			"document_id":     args[0],
			"include_images":  includeImages,
			"include_tables":  includeTables,
			"include_headers": includeHeaders,
			"include_footers": includeFooters,
		})
	},
}

var getDocumentTextCmd = &cobra.Command{
	Use:   "get-document-text <document-id>",
	Short: "Get document as plain text",
	Long:  "Get all text from a document as a single plain text string.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsTool("get_document_text", map[string]any{
			"document_id": args[0],
		})
	},
}

var getDocumentParagraphsCmd = &cobra.Command{
	Use:   "get-document-paragraphs <document-id>",
	Short: "Get document paragraphs",
	Long:  "Get text organized by paragraphs.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsTool("get_document_paragraphs", map[string]any{
			"document_id": args[0],
		})
	},
}

// Slide command flags
var (
	slideIndex    int
	slideObjectID string
	includeNotes  bool
)

// Document command flags
var (
	includeImages  bool
	includeTables  bool
	includeHeaders bool
	includeFooters bool
)

func init() {
	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&credentialsFile, "credentials", "",
		"path to Google service account credentials JSON file (env: GOOGLE_CREDENTIALS_FILE)")
	rootCmd.PersistentFlags().StringVar(&goauthCredentialsFile, "goauth-credentials-file", "",
		"path to goauth CredentialsSet JSON file (env: GOAUTH_CREDENTIALS_FILE)")
	rootCmd.PersistentFlags().StringVar(&goauthCredentialsKey, "goauth-credentials-account", "",
		"account key within goauth CredentialsSet file (env: GOAUTH_CREDENTIALS_ACCOUNT)")
	rootCmd.PersistentFlags().StringVar(&vaultURI, "vault", "",
		"vault URI for credentials (env: OMNITOKEN_VAULT_URI)")
	rootCmd.PersistentFlags().StringVar(&credentialsName, "credentials-name", "",
		"name of credentials in vault (env: OMNITOKEN_CREDENTIALS_NAME)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json",
		"output format: json, pretty (default: json)")

	// Slide command flags
	getSlideCmd.Flags().IntVar(&slideIndex, "index", -1, "zero-based slide index")
	getSlideCmd.Flags().StringVar(&slideObjectID, "object-id", "", "slide object ID")
	getSlideNotesCmd.Flags().IntVar(&slideIndex, "index", -1, "zero-based slide index")
	getSlideNotesCmd.Flags().StringVar(&slideObjectID, "object-id", "", "slide object ID")
	getPresentationContentCmd.Flags().BoolVar(&includeNotes, "include-notes", false, "include speaker notes")

	// Document command flags
	getDocumentContentCmd.Flags().BoolVar(&includeImages, "include-images", false, "include image information")
	getDocumentContentCmd.Flags().BoolVar(&includeTables, "include-tables", false, "include table content")
	getDocumentContentCmd.Flags().BoolVar(&includeHeaders, "include-headers", false, "include document headers")
	getDocumentContentCmd.Flags().BoolVar(&includeFooters, "include-footers", false, "include document footers")

	// Add commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)

	// Slides commands
	rootCmd.AddCommand(getPresentationCmd)
	rootCmd.AddCommand(listSlidesCmd)
	rootCmd.AddCommand(getSlideCmd)
	rootCmd.AddCommand(getSlideNotesCmd)
	rootCmd.AddCommand(getPresentationContentCmd)

	// Docs commands
	rootCmd.AddCommand(getDocumentMetadataCmd)
	rootCmd.AddCommand(getDocumentContentCmd)
	rootCmd.AddCommand(getDocumentTextCmd)
	rootCmd.AddCommand(getDocumentParagraphsCmd)
}

// applyEnvDefaults applies environment variable defaults to flags
func applyEnvDefaults() {
	if credentialsFile == "" {
		credentialsFile = os.Getenv("GOOGLE_CREDENTIALS_FILE")
	}
	if goauthCredentialsFile == "" {
		goauthCredentialsFile = os.Getenv("GOAUTH_CREDENTIALS_FILE")
	}
	if goauthCredentialsKey == "" {
		goauthCredentialsKey = os.Getenv("GOAUTH_CREDENTIALS_ACCOUNT")
	}
	if vaultURI == "" {
		vaultURI = os.Getenv("OMNITOKEN_VAULT_URI")
	}
	if credentialsName == "" {
		credentialsName = os.Getenv("OMNITOKEN_CREDENTIALS_NAME")
	}
	if credentialsName == "" {
		credentialsName = "google"
	}
}

// validateCredentials checks that exactly one credential source is specified
func validateCredentials() error {
	hasGoogleCreds := credentialsFile != ""
	hasGoauthCreds := goauthCredentialsFile != ""
	hasVaultCreds := vaultURI != ""

	credSources := 0
	if hasGoogleCreds {
		credSources++
	}
	if hasGoauthCreds {
		credSources++
	}
	if hasVaultCreds {
		credSources++
	}

	if credSources == 0 {
		return fmt.Errorf("credentials required: use --credentials, --goauth-credentials-file, or --vault")
	}
	if credSources > 1 {
		return fmt.Errorf("only one credential source can be specified")
	}
	return nil
}

// getHTTPClient creates an authenticated HTTP client
func getHTTPClient(ctx context.Context) (*http.Client, func(), error) {
	applyEnvDefaults()
	if err := validateCredentials(); err != nil {
		return nil, nil, err
	}

	opts := auth.TokenManagerOptions{
		CredentialsName: credentialsName,
	}

	// Determine which credentials name to use for GetClient
	clientCredentialsName := credentialsName

	if credentialsFile != "" {
		opts.ServiceAccountFile = credentialsFile
	} else if goauthCredentialsFile != "" {
		opts.GoauthFile = goauthCredentialsFile
		opts.GoauthKey = goauthCredentialsKey
		// Use goauth account key for GetClient
		clientCredentialsName = goauthCredentialsKey
	} else if vaultURI != "" {
		opts.VaultURI = vaultURI
	}

	tokenMgr, err := auth.NewTokenManager(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token manager: %w", err)
	}

	cleanup := func() {
		if err := tokenMgr.Close(); err != nil {
			log.Printf("Warning: failed to close token manager: %v", err)
		}
	}

	httpClient, err := auth.GetClient(ctx, tokenMgr, clientCredentialsName)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to create authenticated client: %w", err)
	}

	return httpClient, cleanup, nil
}

// outputResult outputs the result in the specified format
func outputResult(result any) error {
	var data []byte
	var err error

	switch outputFormat {
	case "pretty":
		data, err = json.MarshalIndent(result, "", "  ")
	default:
		data, err = json.Marshal(result)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// runSlidesTool runs a slides tool by name with the given params
func runSlidesTool(toolName string, params map[string]any) error {
	ctx := context.Background()

	httpClient, cleanup, err := getHTTPClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	skill := slides.New(httpClient)
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize slides skill: %w", err)
	}
	defer skill.Close()

	// Find and call the tool
	for _, tool := range skill.Tools() {
		if tool.Name() == toolName {
			result, err := tool.Call(ctx, params)
			if err != nil {
				return err
			}
			return outputResult(result)
		}
	}
	return fmt.Errorf("tool not found: %s", toolName)
}

// runDocsTool runs a docs tool by name with the given params
func runDocsTool(toolName string, params map[string]any) error {
	ctx := context.Background()

	httpClient, cleanup, err := getHTTPClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	skill := docs.New(httpClient)
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize docs skill: %w", err)
	}
	defer skill.Close()

	// Find and call the tool
	for _, tool := range skill.Tools() {
		if tool.Name() == toolName {
			result, err := tool.Call(ctx, params)
			if err != nil {
				return err
			}
			return outputResult(result)
		}
	}
	return fmt.Errorf("tool not found: %s", toolName)
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	httpClient, cleanup, err := getHTTPClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	// Create omniskill Runtime
	rt := runtime.New(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	// Create and initialize Slides skill
	slidesSkill := slides.New(httpClient)
	if err := slidesSkill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize Slides skill: %w", err)
	}
	defer func() {
		if err := slidesSkill.Close(); err != nil {
			log.Printf("Warning: failed to close Slides skill: %v", err)
		}
	}()

	// Create and initialize Docs skill
	docsSkill := docs.New(httpClient)
	if err := docsSkill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize Docs skill: %w", err)
	}
	defer func() {
		if err := docsSkill.Close(); err != nil {
			log.Printf("Warning: failed to close Docs skill: %v", err)
		}
	}()

	// Register skills with the runtime
	rt.RegisterSkill(slidesSkill)
	rt.RegisterSkill(docsSkill)

	// Run server with stdio transport
	if err := rt.ServeStdio(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
