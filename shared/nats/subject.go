package nats

// Pipeline subject definitions for Birdwatcher v1 domain events.
const (
	SubjectPageFetchRequested = "birdwatcher.v1.page.fetch"
	SubjectUrlDiscovered      = "birdwatcher.v1.url.discovered"
	SubjectPageScraped        = "birdwatcher.v1.page.scraped"
	SubjectSightingText       = "birdwatcher.v1.sighting.text"
	SubjectSightingImage      = "birdwatcher.v1.sighting.image"
	SubjectJobCompleted       = "birdwatcher.v1.job.completed"
)
