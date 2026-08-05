package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SupabaseClient struct {
	baseURL    string
	bucket     string
	serviceKey string
	http       *http.Client
}

func NewSupabaseClient(baseURL, serviceKey, bucket string) *SupabaseClient {
	return &SupabaseClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		bucket:     bucket,
		serviceKey: serviceKey,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SupabaseClient) Upload(ctx context.Context, path, contentType string, body io.Reader) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, normalize(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("supabase upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase upload failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (s *SupabaseClient) Delete(ctx context.Context, path string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, normalize(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("supabase delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase delete failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (s *SupabaseClient) PublicURL(path string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, normalize(path))
}

func normalize(p string) string {
	return strings.TrimPrefix(p, "/")
}
