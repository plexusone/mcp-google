// Package docs provides an omniskill Skill for reading Google Docs documents.
//
// This package can be used standalone with mcp-google or composed
// with other skills in a multi-service MCP server.
package docs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	docsutil "github.com/grokify/gogoogle/docsutil/v1"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// Skill provides Google Docs reading tools.
type Skill struct {
	httpClient *http.Client
	documents  *docs.DocumentsService
}

// New creates a new Docs skill with the given authenticated HTTP client.
func New(httpClient *http.Client) *Skill {
	return &Skill{httpClient: httpClient}
}

// Name returns the skill identifier.
func (s *Skill) Name() string {
	return "docs"
}

// Description returns what this skill does.
func (s *Skill) Description() string {
	return "Read Google Docs documents including content, paragraphs, and metadata"
}

// Init initializes the Google Docs API client.
func (s *Skill) Init(ctx context.Context) error {
	svc, err := docs.NewService(ctx, option.WithHTTPClient(s.httpClient))
	if err != nil {
		return fmt.Errorf("failed to create Docs service: %w", err)
	}
	s.documents = docs.NewDocumentsService(svc)
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error {
	return nil
}

// Tools returns all tools provided by this skill.
func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.getDocumentMetadataTool(),
		s.getDocumentContentTool(),
		s.getDocumentTextTool(),
		s.getDocumentParagraphsTool(),
	}
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

func (s *Skill) getDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	return s.documents.Get(documentID).Context(ctx).Do()
}

func documentMetadata(doc *docs.Document) map[string]any {
	var wordCount, charCount, imageCount, tableCount int
	if doc.Body != nil {
		for _, elem := range doc.Body.Content {
			if elem.Paragraph != nil {
				for _, pe := range elem.Paragraph.Elements {
					if pe.TextRun != nil {
						text := pe.TextRun.Content
						charCount += len(text)
						wordCount += len(strings.Fields(text))
					}
					if pe.InlineObjectElement != nil {
						imageCount++
					}
				}
			}
			if elem.Table != nil {
				tableCount++
			}
		}
	}

	return map[string]any{
		"title":        doc.Title,
		"document_id":  doc.DocumentId,
		"revision_id":  doc.RevisionId,
		"word_count":   wordCount,
		"char_count":   charCount,
		"image_count":  imageCount,
		"table_count":  tableCount,
		"header_count": len(doc.Headers),
		"footer_count": len(doc.Footers),
	}
}

func (s *Skill) getDocumentMetadataTool() skill.Tool {
	return skill.NewTool(
		"get_document_metadata",
		"Get metadata about a Google Doc document including title, word count, and element counts. Accepts document ID or full URL.",
		map[string]skill.Parameter{
			"document_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Doc document",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			documentIDOrURL, _ := params["document_id"].(string)
			if documentIDOrURL == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			docID, err := docsutil.ParseDocumentURL(documentIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid document ID or URL: %w", err)
			}

			doc, err := s.getDocument(ctx, docID)
			if err != nil {
				return nil, fmt.Errorf("failed to get document: %w", err)
			}

			return documentMetadata(doc), nil
		},
	)
}

func (s *Skill) getDocumentContentTool() skill.Tool {
	return skill.NewTool(
		"get_document_content",
		"Get the full structured content of a Google Doc document including headings, paragraphs, and optionally images/tables. Accepts document ID or full URL.",
		map[string]skill.Parameter{
			"document_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Doc document",
				Required:    true,
			},
			"include_images": {
				Type:        "boolean",
				Description: "Include image information in the output",
				Required:    false,
				Default:     false,
			},
			"include_tables": {
				Type:        "boolean",
				Description: "Include table content in the output",
				Required:    false,
				Default:     false,
			},
			"include_headers": {
				Type:        "boolean",
				Description: "Include document headers in the output",
				Required:    false,
				Default:     false,
			},
			"include_footers": {
				Type:        "boolean",
				Description: "Include document footers in the output",
				Required:    false,
				Default:     false,
			},
			"include_metadata": {
				Type:        "boolean",
				Description: "Include document metadata and content counts in the output",
				Required:    false,
				Default:     false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			documentIDOrURL, _ := params["document_id"].(string)
			if documentIDOrURL == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			docID, err := docsutil.ParseDocumentURL(documentIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid document ID or URL: %w", err)
			}

			doc, err := s.getDocument(ctx, docID)
			if err != nil {
				return nil, fmt.Errorf("failed to get document: %w", err)
			}

			includeImages, _ := params["include_images"].(bool)
			includeTables, _ := params["include_tables"].(bool)
			includeHeaders, _ := params["include_headers"].(bool)
			includeFooters, _ := params["include_footers"].(bool)
			includeMetadata, _ := params["include_metadata"].(bool)

			// Extract content using gogoogle helper
			content := docsutil.ExtractDocumentContent(doc)
			docsutil.EnrichImagesWithURIs(content, doc)

			// Convert sections
			sections := make([]map[string]any, len(content.Sections))
			for i, sec := range content.Sections {
				sections[i] = map[string]any{
					"type":     sec.Type,
					"level":    sec.Level,
					"text":     sec.Text,
					"style_id": sec.StyleID,
				}
			}

			result := map[string]any{
				"title":    content.Title,
				"sections": sections,
			}

			if includeMetadata {
				result["metadata"] = documentMetadata(doc)
			}

			if includeImages {
				images := make([]map[string]any, len(content.Images))
				for i, img := range content.Images {
					images[i] = map[string]any{
						"object_id":   img.ObjectID,
						"content_uri": img.ContentURI,
						"source_uri":  img.SourceURI,
						"title":       img.Title,
						"description": img.Description,
					}
				}
				result["images"] = images
			}

			if includeTables {
				tables := make([]map[string]any, len(content.Tables))
				for i, tbl := range content.Tables {
					tables[i] = map[string]any{
						"rows":    tbl.Rows,
						"columns": tbl.Columns,
						"cells":   tbl.Cells,
					}
				}
				result["tables"] = tables
			}

			if includeHeaders {
				var headers []string
				for _, h := range content.Headers {
					if h != "" {
						headers = append(headers, h)
					}
				}
				result["headers"] = headers
			}

			if includeFooters {
				var footers []string
				for _, f := range content.Footers {
					if f != "" {
						footers = append(footers, f)
					}
				}
				result["footers"] = footers
			}

			return result, nil
		},
	)
}

func (s *Skill) getDocumentTextTool() skill.Tool {
	return skill.NewTool(
		"get_document_text",
		"Get all text content from a Google Doc as a single plain text string. Accepts document ID or full URL.",
		map[string]skill.Parameter{
			"document_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Doc document",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			documentIDOrURL, _ := params["document_id"].(string)
			if documentIDOrURL == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			docID, err := docsutil.ParseDocumentURL(documentIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid document ID or URL: %w", err)
			}

			doc, err := s.getDocument(ctx, docID)
			if err != nil {
				return nil, fmt.Errorf("failed to get document: %w", err)
			}

			text := docsutil.ExtractPlainText(doc)

			return map[string]any{
				"title": doc.Title,
				"text":  text,
			}, nil
		},
	)
}

func (s *Skill) getDocumentParagraphsTool() skill.Tool {
	return skill.NewTool(
		"get_document_paragraphs",
		"Get text from a Google Doc organized by paragraphs. Accepts document ID or full URL.",
		map[string]skill.Parameter{
			"document_id": {
				Type:        "string",
				Description: "The ID or URL of the Google Doc document",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			documentIDOrURL, _ := params["document_id"].(string)
			if documentIDOrURL == "" {
				return nil, fmt.Errorf("document_id is required")
			}

			docID, err := docsutil.ParseDocumentURL(documentIDOrURL)
			if err != nil {
				return nil, fmt.Errorf("invalid document ID or URL: %w", err)
			}

			doc, err := s.getDocument(ctx, docID)
			if err != nil {
				return nil, fmt.Errorf("failed to get document: %w", err)
			}

			paragraphs := docsutil.ExtractTextByParagraph(doc)

			return map[string]any{
				"title":      doc.Title,
				"paragraphs": paragraphs,
			}, nil
		},
	)
}
