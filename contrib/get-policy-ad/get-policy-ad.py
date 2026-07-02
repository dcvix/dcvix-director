#!/usr/bin/env python3
"""
Extract dcvix policy data from Active Directory user descriptions.

Queries an AD LDAP server for users whose Description field contains a
<dcvix>...</dcvix> tag, decodes the embedded JSON, and writes a
policydb/users.json file consumable by dcvix-director.

Usage:
    cp .env.example .env   # edit credentials
    pip install ldap3 python-dotenv
    python get-policy-ad.py
"""

from __future__ import annotations

import json
import os
import re
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any

import ldap3
from dotenv import load_dotenv

# ---------------------------------------------------------------------------
# Domain types
# ---------------------------------------------------------------------------

DCVIX_TAG_RE = re.compile(r"<dcvix>(.*?)</dcvix>", re.DOTALL)


@dataclass
class ADUser:
    """Minimal AD user attributes relevant to the script."""

    sAMAccountName: str
    description: str


@dataclass
class DCVIXPolicy:
    """JSON payload expected inside the <dcvix> tag."""

    admin: bool = False
    workstations: list[str] = field(default_factory=list)
    pools: list[str] = field(default_factory=list)


@dataclass
class PolicyUser:
    """A single entry in policydb/users.json."""

    id: str  # maps to sAMAccountName
    admin: bool = False
    workstations: list[str] = field(default_factory=list)
    pools: list[str] = field(default_factory=list)


# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------


@dataclass
class Config:
    ldap_server: str
    bind_dn: str
    bind_password: str
    base_dn: str
    output_dir: Path


def load_config() -> Config:
    """Read configuration from environment variables (already loaded by dotenv)."""
    required = {"LDAP_SERVER", "LDAP_BIND_DN", "LDAP_BIND_PASSWORD", "LDAP_BASE_DN"}
    missing = required - set(os.environ)
    if missing:
        raise SystemExit(
            f"Missing required environment variables: {', '.join(sorted(missing))}\n"
            f"Copy .env.example to .env and fill in the values."
        )

    return Config(
        ldap_server=os.environ["LDAP_SERVER"],
        bind_dn=os.environ["LDAP_BIND_DN"],
        bind_password=os.environ["LDAP_BIND_PASSWORD"],
        base_dn=os.environ["LDAP_BASE_DN"],
        output_dir=Path(os.environ.get("OUTPUT_DIR", "./policydb")),
    )


# ---------------------------------------------------------------------------
# LDAP query
# ---------------------------------------------------------------------------


def query_ad_users(cfg: Config) -> list[ADUser]:
    """
    Connect to AD via LDAPS and retrieve users that have a dcvix tag
    in their Description attribute.
    """
    server = ldap3.Server(cfg.ldap_server, use_ssl=True, get_info=ldap3.NONE)
    conn = ldap3.Connection(server, user=cfg.bind_dn, password=cfg.bind_password, auto_bind=True)

    search_filter = "(&(objectClass=user)(objectCategory=person)(description=*<dcvix>*))"
    attributes = ["sAMAccountName", "description"]

    conn.search(
        search_base=cfg.base_dn,
        search_filter=search_filter,
        attributes=attributes,
        paged_size=500,
    )

    users: list[ADUser] = []
    for entry in conn.entries:
        sam = str(getattr(entry, "sAMAccountName", "") or "")
        desc = str(getattr(entry, "description", "") or "")
        if sam and desc:
            users.append(ADUser(sAMAccountName=sam, description=desc))

    conn.unbind()
    return users


# ---------------------------------------------------------------------------
# Description parsing
# ---------------------------------------------------------------------------


def parse_dcvix_tag(description: str) -> DCVIXPolicy | None:
    """
    Extract the first <dcvix>...</dcvix> block from *description*
    and decode its contents as JSON.

    Returns ``None`` if no valid tag or JSON is found.
    """
    match = DCVIX_TAG_RE.search(description)
    if not match:
        return None

    raw = match.group(1).strip()
    try:
        data: dict[str, Any] = json.loads(raw)
    except json.JSONDecodeError:
        return None

    return DCVIXPolicy(
        admin=bool(data.get("admin", False)),
        workstations=list(data.get("Workstations", [])),
        pools=list(data.get("Pools", [])),
    )


# ---------------------------------------------------------------------------
# Policy generation
# ---------------------------------------------------------------------------


def build_policy_users(ad_users: list[ADUser]) -> list[PolicyUser]:
    """Translate AD users with valid dcvix tags into PolicyUser entries."""
    result: list[PolicyUser] = []
    for ad_user in ad_users:
        policy = parse_dcvix_tag(ad_user.description)
        if policy is None:
            continue
        result.append(
            PolicyUser(
                id=ad_user.sAMAccountName,
                admin=policy.admin,
                workstations=policy.workstations,
                pools=policy.pools,
            )
        )
    return result


def write_users_json(policy_users: list[PolicyUser], output_path: Path) -> None:
    """Write a policydb users.json file compatible with dcvix-director."""
    output_path.parent.mkdir(parents=True, exist_ok=True)
    data = [asdict(u) for u in policy_users]
    output_path.write_text(json.dumps(data, indent=4) + "\n")
    print(f"Wrote {len(policy_users)} user(s) to {output_path}")


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------


def main() -> None:
    load_dotenv()
    cfg = load_config()

    print(f"Connecting to {cfg.ldap_server} ...")
    ad_users = query_ad_users(cfg)
    print(f"Found {len(ad_users)} AD user(s) with a <dcvix> tag.")

    policy_users = build_policy_users(ad_users)
    output_path = cfg.output_dir / "users.json"
    write_users_json(policy_users, output_path)


if __name__ == "__main__":
    main()
