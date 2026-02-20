package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/osteele/liquid"
)

// ErrTemplateNotFound is returned when a template does not exist in S3.
var ErrTemplateNotFound = errors.New("template not found in S3")

// s3Getter abstracts S3 GetObject for testability.
type s3Getter interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// TemplateLoader fetches and caches Liquid email templates from S3.
type TemplateLoader struct {
	s3Client     s3Getter
	bucket       string
	prefix       string // e.g. "email-templates/"
	cache        map[string]*cachedTemplate
	cacheMu      sync.RWMutex
	cacheTTL     time.Duration
	liquidEngine *liquid.Engine
}

type cachedTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
	Meta     TemplateMeta
	LoadedAt time.Time
}

// TemplateMeta holds layout/CTA metadata read from meta.json.
type TemplateMeta struct {
	HeaderTitle    string `json:"header_title"`
	HeaderSubtitle string `json:"header_subtitle"`
	Icon           string `json:"icon"`
	CTAText        string `json:"cta_text"` // Liquid template string
	CTAURL         string `json:"cta_url"`  // Liquid template string
}

// NewTemplateLoader creates a loader pointed at the given S3 bucket and prefix.
func NewTemplateLoader(ctx context.Context, bucket, prefix string) (*TemplateLoader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &TemplateLoader{
		s3Client:     s3.NewFromConfig(cfg),
		bucket:       bucket,
		prefix:       prefix,
		cache:        make(map[string]*cachedTemplate),
		cacheTTL:     5 * time.Minute,
		liquidEngine: liquid.NewEngine(),
	}, nil
}

// LoadTemplate fetches a template set from S3 or returns the cached version.
func (tl *TemplateLoader) LoadTemplate(ctx context.Context, templatePath string) (*cachedTemplate, error) {
	// Fast-path: check cache under read lock.
	tl.cacheMu.RLock()
	if ct, ok := tl.cache[templatePath]; ok && time.Since(ct.LoadedAt) < tl.cacheTTL {
		tl.cacheMu.RUnlock()
		return ct, nil
	}
	tl.cacheMu.RUnlock()

	// Slow-path: fetch from S3 under write lock with double-check.
	tl.cacheMu.Lock()
	defer tl.cacheMu.Unlock()

	if ct, ok := tl.cache[templatePath]; ok && time.Since(ct.LoadedAt) < tl.cacheTTL {
		return ct, nil
	}

	base := tl.prefix + templatePath + "/"

	subject, err := tl.fetchS3(ctx, base+"subject.liquid")
	if err != nil {
		return nil, err
	}

	htmlBody, err := tl.fetchS3(ctx, base+"body.html.liquid")
	if err != nil {
		return nil, err
	}

	textBody, _ := tl.fetchS3(ctx, base+"body.txt.liquid") // optional

	meta := TemplateMeta{}
	if raw, metaErr := tl.fetchS3(ctx, base+"meta.json"); metaErr == nil {
		_ = json.Unmarshal([]byte(raw), &meta)
	}

	ct := &cachedTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Meta:     meta,
		LoadedAt: time.Now(),
	}
	tl.cache[templatePath] = ct
	return ct, nil
}

// RenderTemplate loads and renders all template parts with the given bindings.
func (tl *TemplateLoader) RenderTemplate(ctx context.Context, templatePath string, bindings map[string]interface{}) (subject, html, text string, meta TemplateMeta, err error) {
	ct, err := tl.LoadTemplate(ctx, templatePath)
	if err != nil {
		return "", "", "", TemplateMeta{}, err
	}

	subject, err = tl.render(ct.Subject, bindings)
	if err != nil {
		return "", "", "", TemplateMeta{}, fmt.Errorf("render subject: %w", err)
	}

	html, err = tl.render(ct.HTMLBody, bindings)
	if err != nil {
		return "", "", "", TemplateMeta{}, fmt.Errorf("render html: %w", err)
	}

	text, _ = tl.render(ct.TextBody, bindings) // ok if empty

	return subject, html, text, ct.Meta, nil
}

// RenderString renders a single Liquid template string with the given bindings.
func (tl *TemplateLoader) RenderString(source string, bindings map[string]interface{}) string {
	if source == "" {
		return ""
	}
	out, err := tl.render(source, bindings)
	if err != nil {
		log.Printf("WARNING: liquid render failed for %q: %v", source, err)
		return source
	}
	return out
}

func (tl *TemplateLoader) render(source string, bindings map[string]interface{}) (string, error) {
	if source == "" {
		return "", nil
	}
	return tl.liquidEngine.ParseAndRenderString(source, bindings)
}

func (tl *TemplateLoader) fetchS3(ctx context.Context, key string) (string, error) {
	out, err := tl.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(tl.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *s3types.NoSuchKey
		if errors.As(err, &nsk) {
			return "", ErrTemplateNotFound
		}
		var notFound *s3types.NotFound
		if errors.As(err, &notFound) {
			return "", ErrTemplateNotFound
		}
		return "", fmt.Errorf("s3 GetObject %s: %w", key, err)
	}
	defer out.Body.Close()

	b, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("read s3 body %s: %w", key, err)
	}
	return string(b), nil
}
