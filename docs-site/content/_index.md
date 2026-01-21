---
title: CC-Relay
layout: hextra-home
---

{{< hextra/hero-headline >}}
  CC-Relay
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  A blazing-fast, multi-provider LLM proxy&nbsp;<br class="sm:block hidden" />for Claude Code and beyond
{{< /hextra/hero-subtitle >}}

{{< hextra/hero-button text="Get Started" link="/docs/getting-started/" >}}
{{< hextra/hero-button text="GitHub →" link="https://github.com/omarluq/cc-relay" style="secondary" >}}

<div class="mt-6 mb-6">
{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Multi-Provider Support"
    subtitle="Connect to Anthropic, Z.AI, Ollama, AWS Bedrock, Azure, and Vertex AI from a single endpoint"
  >}}
  {{< hextra/feature-card
    title="Rate Limit Pooling"
    subtitle="Intelligently distribute requests across multiple API keys to maximize throughput"
  >}}
  {{< hextra/feature-card
    title="Cost Optimization"
    subtitle="Route requests based on cost, latency, or model availability for optimal efficiency"
  >}}
  {{< hextra/feature-card
    title="Automatic Failover"
    subtitle="Circuit breaker with health tracking ensures high availability across providers"
  >}}
  {{< hextra/feature-card
    title="Real-time Monitoring"
    subtitle="TUI dashboard with live stats, provider health, and request logging"
  >}}
  {{< hextra/feature-card
    title="Hot Reload Config"
    subtitle="Update provider settings and routing strategies without restart"
  >}}
{{< /hextra/feature-grid >}}
</div>

## Why CC-Relay?

**Claude Code** users often hit rate limits. **CC-Relay** solves this by:

- **Smart Routing**: Shuffle, round-robin, failover, cost-based, latency-based, or model-based
- **API Key Pools**: Manage multiple keys per provider with RPM/TPM tracking
- **Cloud Provider Support**: Native integration with Bedrock, Azure Foundry, and Vertex AI
- **Health Tracking**: Automatic circuit breaking and recovery for failed providers
- **SSE Streaming**: Perfect compatibility with Claude Code's real-time streaming
- **Management API**: gRPC interface for stats, config updates, and provider control

## Quick Start

```bash
# Install
go install github.com/omarluq/cc-relay@latest

# Run with example config
cc-relay serve --config config/example.yaml

# Point Claude Code to the proxy
export ANTHROPIC_BASE_URL="http://localhost:8787"
export ANTHROPIC_API_KEY="managed-by-cc-relay"
claude
```

## Architecture

CC-Relay sits between your LLM client and multiple providers, intelligently routing requests based on your configured strategy:

<div class="architecture-diagram">
  <div class="arch-layer">
    <div class="arch-layer-title">Client Layer</div>
    <div class="arch-node arch-node-client">
      Claude Code Client<br/>
      <span style="font-size: 0.875rem; opacity: 0.9;">POST /v1/messages</span>
    </div>
  </div>

  <div class="arch-connector">↓</div>

  <div class="arch-layer">
    <div class="arch-layer-title">Proxy Engine</div>
    <div class="arch-proxy">
      <div class="arch-proxy-component">🔐 Authentication</div>
      <div class="arch-proxy-component">🎯 Smart Router</div>
      <div class="arch-proxy-component">💚 Health Tracker</div>
      <div class="arch-proxy-component">🔑 API Key Pool</div>
    </div>
  </div>

  <div class="arch-connector">↓</div>

  <div class="arch-layer">
    <div class="arch-layer-title">Provider Layer</div>
    <div class="arch-providers">
      <div class="arch-provider anthropic">
        <img src="/logos/anthropic.svg" alt="Anthropic" class="arch-provider-logo" />
        <div class="arch-provider-name">Anthropic</div>
        <div class="arch-provider-desc">Claude Models</div>
      </div>
      <div class="arch-provider bedrock">
        <img src="/logos/aws.svg" alt="AWS Bedrock" class="arch-provider-logo" />
        <div class="arch-provider-name">AWS Bedrock</div>
        <div class="arch-provider-desc">SigV4 Auth</div>
      </div>
      <div class="arch-provider azure">
        <img src="/logos/azure.svg" alt="Azure" class="arch-provider-logo" />
        <div class="arch-provider-name">Azure Foundry</div>
        <div class="arch-provider-desc">Deployments</div>
      </div>
      <div class="arch-provider vertex">
        <img src="/logos/gcp.svg" alt="Vertex AI" class="arch-provider-logo" />
        <div class="arch-provider-name">Vertex AI</div>
        <div class="arch-provider-desc">OAuth</div>
      </div>
      <div class="arch-provider ollama">
        <img src="/logos/ollama.svg" alt="Ollama" class="arch-provider-logo" />
        <div class="arch-provider-name">Ollama</div>
        <div class="arch-provider-desc">Local Models</div>
      </div>
      <div class="arch-provider zai">
        <img src="/logos/openai.svg" alt="Z.AI" class="arch-provider-logo" />
        <div class="arch-provider-name">Z.AI</div>
        <div class="arch-provider-desc">GLM Models</div>
      </div>
    </div>
  </div>
</div>

## Features in Detail

### Rate Limit Management

Automatically track and respect provider rate limits (RPM/TPM):

- Per-key rate tracking with token bucket algorithm
- Automatic key rotation when limits are reached
- Configurable retry strategies and backoff

### Cost-Based Routing

Route requests to the cheapest available provider:

```yaml
routing:
  strategy: cost-based
  max_cost_per_million_tokens: 15.0
```

### Health Tracking

Circuit breaker with three states (CLOSED/OPEN/HALF-OPEN):

- Automatic failure detection (429s, 5xx, timeouts)
- Configurable failure thresholds and cooldown periods
- Health probe recovery after cooldown

### Streaming Support

Perfect SSE event sequence matching Anthropic API:

- `message_start` → `content_block_start` → `content_block_delta` → `content_block_stop` → `message_delta` → `message_stop`
- Preserves `tool_use_id` for Claude Code's parallel tool calls
- Supports extended thinking content blocks

## Use Cases

- **Development Teams**: Share API quota across multiple developers
- **CI/CD Pipelines**: High-throughput testing with rate limit pooling
- **Cost Optimization**: Route to cheapest provider while maintaining quality
- **High Availability**: Automatic failover ensures uptime
- **Multi-Cloud**: Leverage Bedrock, Azure, and Vertex AI simultaneously

## Documentation

<div class="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2">
  {{< hextra/feature-card
    title="Getting Started"
    subtitle="Installation, configuration, and first run"
    link="/docs/getting-started/"
  >}}
  {{< hextra/feature-card
    title="Configuration"
    subtitle="Provider setup, routing strategies, and advanced options"
    link="/docs/configuration/"
  >}}
  {{< hextra/feature-card
    title="Architecture"
    subtitle="System design, components, and API compatibility"
    link="/docs/architecture/"
  >}}
  {{< hextra/feature-card
    title="API Reference"
    subtitle="gRPC management API and REST endpoints"
    link="/docs/api/"
  >}}
</div>

## Contributing

CC-Relay is open source! Contributions are welcome.

- [Report bugs](https://github.com/omarluq/cc-relay/issues)
- [Request features](https://github.com/omarluq/cc-relay/issues)
- [Submit PRs](https://github.com/omarluq/cc-relay/pulls)

## License

MIT License - see [LICENSE](https://github.com/omarluq/cc-relay/blob/main/LICENSE) for details.
