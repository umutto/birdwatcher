package consumer

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/shared/nats"
	"google.golang.org/protobuf/proto"
)

var (
	testCrawlID = uuid.MustParse("0195c000-0000-7000-8000-000000000001")
)

type mockRepo struct {
	repository.Querier
	err error

	updatedTitleParams      *repository.UpdateCrawlTitleParams
	incrementProgressParams *repository.IncrementCrawlProgressParams
	createdSightingParams   *repository.CreateSightingParams
	updatedStateParams      *repository.UpdateCrawlStateParams
}

func (m *mockRepo) UpdateCrawlTitle(_ context.Context, arg repository.UpdateCrawlTitleParams) error {
	m.updatedTitleParams = &arg
	return m.err
}

func (m *mockRepo) IncrementCrawlProgress(_ context.Context, arg repository.IncrementCrawlProgressParams) error {
	m.incrementProgressParams = &arg
	return m.err
}

func (m *mockRepo) CreateSighting(_ context.Context, arg repository.CreateSightingParams) (repository.Sighting, error) {
	m.createdSightingParams = &arg
	return repository.Sighting{}, m.err
}

func (m *mockRepo) UpdateCrawlState(_ context.Context, arg repository.UpdateCrawlStateParams) (repository.Crawl, error) {
	m.updatedStateParams = &arg
	return repository.Crawl{}, m.err
}

type mockMsg struct {
	jetstream.Msg
	subject string
	data    []byte
	acked   bool
	naked   bool
	termed  bool
}

func (m *mockMsg) Subject() string { return m.subject }
func (m *mockMsg) Data() []byte     { return m.data }
func (m *mockMsg) Ack() error       { m.acked = true; return nil }
func (m *mockMsg) Nak() error       { m.naked = true; return nil }
func (m *mockMsg) Term() error      { m.termed = true; return nil }

func TestConsumer_HandlePageScraped(t *testing.T) {
	t.Run("successfully updates title and increments progress", func(t *testing.T) {
		evt := &birdwatcherv1.PageScraped{
			Meta:      &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			Title:     "The Birds of America",
			ImageUrls: []string{"img1.jpg", "img2.jpg"},
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectPageScraped, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.acked {
			t.Errorf("expected message to be acked")
		}
		if repo.updatedTitleParams == nil || *repo.updatedTitleParams.Title != "The Birds of America" {
			t.Errorf("expected title %q, got %+v", "The Birds of America", repo.updatedTitleParams)
		}
		if repo.incrementProgressParams == nil || repo.incrementProgressParams.PagesCrawledCount != 1 || repo.incrementProgressParams.ImagesScannedCount != 2 {
			t.Errorf("expected progress (pages=1, images=2), got %+v", repo.incrementProgressParams)
		}
	})

	t.Run("naks when database fails", func(t *testing.T) {
		evt := &birdwatcherv1.PageScraped{
			Meta: &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{err: errors.New("db connection failure")}
		msg := &mockMsg{subject: nats.SubjectPageScraped, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.naked {
			t.Errorf("expected message to be naked on db error")
		}
	})

	t.Run("terms message when protobuf is malformed", func(t *testing.T) {
		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectPageScraped, data: []byte("corrupt-bytes")}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.termed {
			t.Errorf("expected message to be termed on corrupt bytes")
		}
	})

	t.Run("terms message when job id is invalid uuid", func(t *testing.T) {
		evt := &birdwatcherv1.PageScraped{
			Meta: &birdwatcherv1.EventMetadata{JobId: "not-a-uuid"},
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectPageScraped, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.termed {
			t.Errorf("expected message to be termed on invalid uuid")
		}
	})
}

func TestConsumer_HandleSightingText(t *testing.T) {
	t.Run("successfully records text sighting", func(t *testing.T) {
		evt := &birdwatcherv1.SightingText{
			Meta:               &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			PageUrl:            "https://example.com/birds",
			SpeciesPlateNumber: 11,
			Confidence:         0.92,
			Snippet:            "We saw a bald eagle flying near the river.",
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectSightingText, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.acked {
			t.Errorf("expected message to be acked")
		}
		if repo.createdSightingParams == nil {
			t.Fatal("expected CreateSighting to be called")
		}
		if repo.createdSightingParams.SpeciesPlateNumber != 11 {
			t.Errorf("expected plate 11, got %d", repo.createdSightingParams.SpeciesPlateNumber)
		}
		if repo.createdSightingParams.SourceType != repository.SightingSourceTypeText {
			t.Errorf("expected source_type 'text', got %s", repo.createdSightingParams.SourceType)
		}
		if repo.createdSightingParams.Confidence != 0.92 {
			t.Errorf("expected confidence 0.92, got %f", repo.createdSightingParams.Confidence)
		}
	})

	t.Run("naks on database error", func(t *testing.T) {
		evt := &birdwatcherv1.SightingText{
			Meta:               &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			SpeciesPlateNumber: 11,
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{err: errors.New("db error")}
		msg := &mockMsg{subject: nats.SubjectSightingText, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.naked {
			t.Errorf("expected message to be naked on db error")
		}
	})
}

func TestConsumer_HandleSightingImage(t *testing.T) {
	t.Run("successfully records image sighting", func(t *testing.T) {
		evt := &birdwatcherv1.SightingImage{
			Meta:               &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			PageUrl:            "https://example.com/birds",
			SpeciesPlateNumber: 42,
			Confidence:         0.85,
			ImageUrl:           "https://example.com/eagle.jpg",
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectSightingImage, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.acked {
			t.Errorf("expected message to be acked")
		}
		if repo.createdSightingParams == nil {
			t.Fatal("expected CreateSighting to be called")
		}
		if repo.createdSightingParams.SpeciesPlateNumber != 42 {
			t.Errorf("expected plate 42, got %d", repo.createdSightingParams.SpeciesPlateNumber)
		}
		if repo.createdSightingParams.SourceType != repository.SightingSourceTypeImage {
			t.Errorf("expected source_type 'image', got %s", repo.createdSightingParams.SourceType)
		}
		if repo.createdSightingParams.ImageUrl == nil || *repo.createdSightingParams.ImageUrl != "https://example.com/eagle.jpg" {
			t.Errorf("expected image url, got %+v", repo.createdSightingParams.ImageUrl)
		}
	})

	t.Run("naks on database error", func(t *testing.T) {
		evt := &birdwatcherv1.SightingImage{
			Meta:               &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			SpeciesPlateNumber: 42,
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{err: errors.New("db error")}
		msg := &mockMsg{subject: nats.SubjectSightingImage, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.naked {
			t.Errorf("expected message to be naked on db error")
		}
	})
}

func TestConsumer_HandleJobCompleted(t *testing.T) {
	t.Run("successfully transitions crawl state", func(t *testing.T) {
		evt := &birdwatcherv1.JobCompleted{
			Meta:   &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			Status: birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED,
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectJobCompleted, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.acked {
			t.Errorf("expected message to be acked")
		}
		if repo.updatedStateParams == nil || repo.updatedStateParams.State != repository.CrawlStateCompleted {
			t.Errorf("expected state %s, got %+v", repository.CrawlStateCompleted, repo.updatedStateParams)
		}
	})

	t.Run("naks on database error", func(t *testing.T) {
		evt := &birdwatcherv1.JobCompleted{
			Meta:   &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			Status: birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED,
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{err: errors.New("db error")}
		msg := &mockMsg{subject: nats.SubjectJobCompleted, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.naked {
			t.Errorf("expected message to be naked on db error")
		}
	})

	t.Run("terms message on unsupported status", func(t *testing.T) {
		evt := &birdwatcherv1.JobCompleted{
			Meta:   &birdwatcherv1.EventMetadata{JobId: testCrawlID.String()},
			Status: birdwatcherv1.CrawlStatus_CRAWL_STATUS_UNSPECIFIED,
		}
		data, err := proto.Marshal(evt)
		if err != nil {
			t.Fatalf("marshal proto: %v", err)
		}

		repo := &mockRepo{}
		msg := &mockMsg{subject: nats.SubjectJobCompleted, data: data}
		c := New(slog.New(slog.DiscardHandler), repo, nil)

		c.handleMessage(context.Background(), msg)

		if !msg.termed {
			t.Errorf("expected message to be termed on unsupported status")
		}
	})
}

func TestConsumer_HandleMessage_UnknownSubject(t *testing.T) {
	repo := &mockRepo{}
	msg := &mockMsg{subject: "unknown.subject", data: []byte("data")}
	c := New(slog.New(slog.DiscardHandler), repo, nil)

	c.handleMessage(context.Background(), msg)

	if !msg.termed {
		t.Errorf("expected unknown subject to be termed")
	}
}
