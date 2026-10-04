#!/usr/bin/env python3
"""
Create release marketing copy and optionally publish to supported channels.

Usage:
  python scripts/release_broadcast.py --tag v0.1.4 --repo Mutasem-mk4/procscope
  python scripts/release_broadcast.py --tag v0.1.4 --repo Mutasem-mk4/procscope --publish
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import pathlib
import sys
import urllib.error
import urllib.request


ROOT = pathlib.Path(__file__).resolve().parents[1]
OUT_DIR = ROOT / "docs" / "marketing" / "autogen"


def _post_json(url: str, payload: dict) -> tuple[bool, str]:
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url=url,
        data=data,
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            return True, f"HTTP {resp.status}"
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        return False, f"HTTPError {exc.code}: {body}"
    except Exception as exc:  # noqa: BLE001
        return False, str(exc)


def _write_marketing_bundle(tag: str, repo: str) -> pathlib.Path:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    stamp = dt.datetime.utcnow().strftime("%Y-%m-%d")
    out = OUT_DIR / f"launch-{tag}.md"
    url = f"https://github.com/{repo}/releases/tag/{tag}"
    body = f"""# Launch Copy Pack ({tag})

Generated: {stamp} (UTC)

## X/Twitter (single post)

{tag} of procscope is out.

Process-scoped eBPF tracing for Linux incident response:
- trace suspicious binaries without ptrace overhead
- generate JSONL + evidence bundles + markdown reports
- ship as a single binary

Release: {url}
Repo: https://github.com/{repo}

#eBPF #Linux #CyberSecurity #OpenSource

## LinkedIn

I just shipped {tag} of procscope.

procscope helps incident responders and security engineers trace what a Linux process actually did, with process-scoped eBPF visibility and low-noise outputs designed for triage.

Highlights:
- process-tree scoped tracing
- event timeline + JSONL + evidence bundle outputs
- easy install from GitHub releases

Release notes: {url}
GitHub: https://github.com/{repo}

## Dev.to / Blog Intro Paragraph

Today I released {tag} of procscope, a process-scoped eBPF tracer built for malware triage and incident response. Instead of collecting host-wide noise, procscope follows the specific process tree you care about and outputs investigation-ready artifacts (timeline, JSONL, bundle, and summary) you can share with your team.

## Manual-only Channels Checklist

- [ ] Hacker News `Show HN` post submitted
- [ ] Reddit post in relevant subreddit submitted
- [ ] Replies monitored during first 24h
"""
    out.write_text(body, encoding="utf-8")
    return out


def _publish_if_configured(tag: str, repo: str) -> list[str]:
    logs: list[str] = []
    release_url = f"https://github.com/{repo}/releases/tag/{tag}"
    text = (
        f"{tag} of procscope is live: {release_url}\n"
        f"Process-scoped eBPF tracing for incident response.\n"
        "Star the repo: https://github.com/{repo}"
    )
    text = text.format(repo=repo)

    discord_webhook = os.getenv("DISCORD_WEBHOOK_URL", "").strip()
    if discord_webhook:
        ok, msg = _post_json(discord_webhook, {"content": text})
        logs.append(f"discord: {'ok' if ok else 'fail'} ({msg})")
    else:
        logs.append("discord: skipped (DISCORD_WEBHOOK_URL not set)")

    slack_webhook = os.getenv("SLACK_WEBHOOK_URL", "").strip()
    if slack_webhook:
        ok, msg = _post_json(slack_webhook, {"text": text})
        logs.append(f"slack: {'ok' if ok else 'fail'} ({msg})")
    else:
        logs.append("slack: skipped (SLACK_WEBHOOK_URL not set)")

    telegram_token = os.getenv("TELEGRAM_BOT_TOKEN", "").strip()
    telegram_chat_id = os.getenv("TELEGRAM_CHAT_ID", "").strip()
    if telegram_token and telegram_chat_id:
        url = f"https://api.telegram.org/bot{telegram_token}/sendMessage"
        ok, msg = _post_json(url, {"chat_id": telegram_chat_id, "text": text})
        logs.append(f"telegram: {'ok' if ok else 'fail'} ({msg})")
    else:
        logs.append("telegram: skipped (TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID not set)")

    return logs


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate release marketing bundle")
    parser.add_argument("--tag", required=True, help="Release tag, e.g. v0.1.4")
    parser.add_argument("--repo", required=True, help="owner/repo, e.g. Mutasem-mk4/procscope")
    parser.add_argument("--publish", action="store_true", help="Publish to configured webhooks")
    args = parser.parse_args()

    out = _write_marketing_bundle(tag=args.tag, repo=args.repo)
    print(f"Generated: {out}")

    if args.publish:
        logs = _publish_if_configured(tag=args.tag, repo=args.repo)
        for line in logs:
            print(line)
    else:
        print("Publish skipped (pass --publish to enable).")

    return 0


if __name__ == "__main__":
    sys.exit(main())
