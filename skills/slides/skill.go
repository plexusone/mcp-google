// Package slides provides an omniskill Skill for reading Google Slides presentations.
//
// This package can be used standalone with mcp-google or composed
// with other skills in a multi-service MCP server.
package slides

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grokify/gogoogle/slidesutil/v1"
	"github.com/plexusone/omniskill/skill"
	"google.golang.org/api/option"
	gslides "google.golang.org/api/slides/v1"
)

// Skill provides Google Slides reading tools.
type Skill struct {
	httpClient    *http.Client
	presentations *gslides.PresentationsService
}

// New creates a new Slides skill with the given authenticated HTTP client.
func New(httpClient *http.Client) *Skill {
	return &Skill{httpClient: httpClient}
}

// Name returns the skill identifier.
func (s *Skill) Name() string {
	return "slides"
}

// Description returns what this skill does.
func (s *Skill) Description() string {
	return "Read Google Slides presentations including content, notes, and images"
}

// Init initializes the Google Slides API client.
func (s *Skill) Init(ctx context.Context) error {
	svc, err := gslides.NewService(ctx, option.WithHTTPClient(s.httpClient))
	if err != nil {
		return fmt.Errorf("failed to create Slides service: %w", err)
	}
	s.presentations = gslides.NewPresentationsService(svc)
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error {
	return nil
}

// Tools returns all tools provided by this skill.
func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.getPresentationTool(),
		s.listSlidesTool(),
		s.getSlideTool(),
		s.getSlideNotesTool(),
		s.getPresentationContentTool(),
	}
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

func (s *Skill) getPresentation(ctx context.Context, presentationID string) (*gslides.Presentation, error) {
	return s.presentations.Get(presentationID).Context(ctx).Do()
}

func (s *Skill) getPresentationTool() skill.Tool {
	return skill.NewTool(
		"get_presentation",
		"Get metadata about a Google Slides presentation including title, slide count, locale, and revision ID",
		map[string]skill.Parameter{
			"presentation_id": {
				Type:        "string",
				Description: "The ID of the Google Slides presentation",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			presentationID, _ := params["presentation_id"].(string)
			if presentationID == "" {
				return nil, fmt.Errorf("presentation_id is required")
			}

			pres, err := s.getPresentation(ctx, presentationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get presentation: %w", err)
			}

			return map[string]any{
				"title":       pres.Title,
				"slide_count": len(pres.Slides),
				"locale":      pres.Locale,
				"revision_id": pres.RevisionId,
			}, nil
		},
	)
}

func (s *Skill) listSlidesTool() skill.Tool {
	return skill.NewTool(
		"list_slides",
		"List all slides in a Google Slides presentation with their titles and element counts",
		map[string]skill.Parameter{
			"presentation_id": {
				Type:        "string",
				Description: "The ID of the Google Slides presentation",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			presentationID, _ := params["presentation_id"].(string)
			if presentationID == "" {
				return nil, fmt.Errorf("presentation_id is required")
			}

			pres, err := s.getPresentation(ctx, presentationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get presentation: %w", err)
			}

			slides := make([]map[string]any, len(pres.Slides))
			for i, slide := range pres.Slides {
				slides[i] = map[string]any{
					"object_id":     slide.ObjectId,
					"index":         i,
					"title":         slidesutil.ExtractSlideTitle(slide),
					"element_count": len(slide.PageElements),
				}
			}

			return map[string]any{"slides": slides}, nil
		},
	)
}

func (s *Skill) getSlideTool() skill.Tool {
	return skill.NewTool(
		"get_slide",
		"Get the content and elements of a specific slide by index or object ID",
		map[string]skill.Parameter{
			"presentation_id": {
				Type:        "string",
				Description: "The ID of the Google Slides presentation",
				Required:    true,
			},
			"slide_index": {
				Type:        "integer",
				Description: "The zero-based index of the slide (mutually exclusive with slide_object_id)",
				Required:    false,
			},
			"slide_object_id": {
				Type:        "string",
				Description: "The object ID of the slide (mutually exclusive with slide_index)",
				Required:    false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			presentationID, _ := params["presentation_id"].(string)
			if presentationID == "" {
				return nil, fmt.Errorf("presentation_id is required")
			}

			pres, err := s.getPresentation(ctx, presentationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get presentation: %w", err)
			}

			// Parse slide_index (may come as float64 from JSON)
			var slideIndex *int
			if idx, ok := params["slide_index"]; ok && idx != nil {
				switch v := idx.(type) {
				case float64:
					i := int(v)
					slideIndex = &i
				case int:
					slideIndex = &v
				}
			}
			slideObjectID, _ := params["slide_object_id"].(string)

			slide, idx, err := slidesutil.FindSlide(pres, slideIndex, slideObjectID)
			if err != nil {
				return nil, err
			}

			textContent := slidesutil.ExtractTextContent(slide)
			images := slidesutil.ExtractImages(slide)

			elements := make([]map[string]any, len(slide.PageElements))
			for i, elem := range slide.PageElements {
				elements[i] = map[string]any{
					"object_id":    elem.ObjectId,
					"element_type": slidesutil.GetElementType(elem),
					"description":  slidesutil.GetElementDescription(elem),
					"image_url":    slidesutil.GetImageURL(elem),
				}
			}

			imageInfos := make([]map[string]any, len(images))
			for i, img := range images {
				imageInfos[i] = map[string]any{
					"object_id":   img.ObjectID,
					"content_url": img.ContentURL,
					"source_url":  img.SourceURL,
					"alt_text":    img.AltText,
				}
			}

			return map[string]any{
				"index":           idx,
				"title":           slidesutil.ExtractSlideTitle(slide),
				"text_content":    textContent,
				"element_summary": elements,
				"images":          imageInfos,
			}, nil
		},
	)
}

func (s *Skill) getSlideNotesTool() skill.Tool {
	return skill.NewTool(
		"get_slide_notes",
		"Get the speaker notes for a specific slide by index or object ID",
		map[string]skill.Parameter{
			"presentation_id": {
				Type:        "string",
				Description: "The ID of the Google Slides presentation",
				Required:    true,
			},
			"slide_index": {
				Type:        "integer",
				Description: "The zero-based index of the slide (mutually exclusive with slide_object_id)",
				Required:    false,
			},
			"slide_object_id": {
				Type:        "string",
				Description: "The object ID of the slide (mutually exclusive with slide_index)",
				Required:    false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			presentationID, _ := params["presentation_id"].(string)
			if presentationID == "" {
				return nil, fmt.Errorf("presentation_id is required")
			}

			pres, err := s.getPresentation(ctx, presentationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get presentation: %w", err)
			}

			var slideIndex *int
			if idx, ok := params["slide_index"]; ok && idx != nil {
				switch v := idx.(type) {
				case float64:
					i := int(v)
					slideIndex = &i
				case int:
					slideIndex = &v
				}
			}
			slideObjectID, _ := params["slide_object_id"].(string)

			slide, _, err := slidesutil.FindSlide(pres, slideIndex, slideObjectID)
			if err != nil {
				return nil, err
			}

			notes := slidesutil.ExtractNotesText(slide)

			return map[string]any{"notes": notes}, nil
		},
	)
}

func (s *Skill) getPresentationContentTool() skill.Tool {
	return skill.NewTool(
		"get_presentation_content",
		"Get all slide content (text and images) in a single call, ideal for AI analysis of the entire presentation",
		map[string]skill.Parameter{
			"presentation_id": {
				Type:        "string",
				Description: "The ID of the Google Slides presentation",
				Required:    true,
			},
			"include_notes": {
				Type:        "boolean",
				Description: "Include speaker notes for each slide",
				Required:    false,
				Default:     false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			presentationID, _ := params["presentation_id"].(string)
			if presentationID == "" {
				return nil, fmt.Errorf("presentation_id is required")
			}

			includeNotes, _ := params["include_notes"].(bool)

			pres, err := s.getPresentation(ctx, presentationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get presentation: %w", err)
			}

			content := slidesutil.ExtractPresentationContent(pres, includeNotes)

			slides := make([]map[string]any, len(content.Slides))
			for i, sc := range content.Slides {
				images := make([]map[string]any, len(sc.Images))
				for j, img := range sc.Images {
					images[j] = map[string]any{
						"object_id":   img.ObjectID,
						"content_url": img.ContentURL,
						"source_url":  img.SourceURL,
						"alt_text":    img.AltText,
					}
				}

				slides[i] = map[string]any{
					"index":        sc.Index,
					"object_id":    sc.ObjectID,
					"title":        sc.Title,
					"text_content": sc.TextContent,
					"images":       images,
				}
				if includeNotes {
					slides[i]["notes"] = sc.Notes
				}
			}

			return map[string]any{
				"title":  content.Title,
				"slides": slides,
			}, nil
		},
	)
}
