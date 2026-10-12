# birdwatcher

[![CI](https://github.com/umutto/birdwatcher/actions/workflows/ci.yml/badge.svg)](https://github.com/umutto/birdwatcher/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Python / uv](https://img.shields.io/badge/uv-Python_3.14+-DE5FE9?logo=astral&logoColor=white)](https://docs.astral.sh/uv/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![Protobuf](https://img.shields.io/badge/Protobuf-Buf_v2-244c5a?logo=buf&logoColor=white)](https://buf.build/)
[![NATS JetStream](https://img.shields.io/badge/NATS-JetStream-27AAE1?logo=nats.io&logoColor=white)](https://nats.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A practice project of a distributed, event-driven web scraper and recognition pipeline built with Go to extract, classify, and index bird species from a given URL.

---

## Data Sources & Attribution

- **Illustrations**: John James Audubon, *The Birds of America* (1827–1838), Public Domain. All bird plates are Courtesy of the John James Audubon Center at Mill Grove, Montgomery County Audubon Collection, and Zebra Publishing.
- **Audio Recordings**: Sourced from [xeno-canto.org](https://xeno-canto.org). Individual clips are licensed under Creative Commons (see `metadata.audio` in `species.json` for specific recordists and licenses).
- **Species Summaries**: Sourced from [Wikipedia](https://wikipedia.org) under CC BY-SA 4.0.
