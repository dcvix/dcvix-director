# get-policy-ad

An **example** script that demonstrates how to use Active Directory as a
data source for the dcvix-director **PolicyDB**.

## Purpose

dcvix-director's policy database (`policydb/users.json` and
`policydb/pools.json`) is a pair of plain JSON files that define which
users can access which DCV servers (directly or through named pools).

The recommended workflow is to generate these files from an external
authoritative source (HR database, CMDB, etc.) rather than editing them
by hand. This script shows one approach using **Active Directory** as
that source.

Add this script to a cron job and issue an HUP signa to the director to reload the policy (`kill -HUP $(pidof dcvix-director)`).

## How it works

1. Connects to an AD LDAP server over LDAPS.
2. Searches for user objects whose **Description** field contains a
   `<dcvix>...</dcvix>` tag.
3. Parses the tagged JSON payload, which can contain:
   ```json
   {
       "admin": true,
       "Workstations": ["ws1.domain.com", "ws2.domain.com"],
       "Pools": ["testing"]
   }
   ```
4. Writes a `policydb/users.json` file compatible with dcvix-director.

### AD Description field example

```
Senior Engineer <dcvix>{"admin":false,"Workstations":["ws1.domain.com","ws2.domain.com"],"Pools":["testing"]}</dcvix>
```

The script extracts the tagged section, decodes the JSON, and creates a
policy entry using `sAMAccountName` as the user ID.

## Usage

2. **Create a virtual env and install dependencies**

   ```bash
   python3 -m venv venv
   source venv/bin/activate
   pip install ldap3 python-dotenv
   ```

2. **Configure**

   ```bash
   cp .env.example .env
   # Edit .env with your AD server details
   ```

3. **Run**

   ```bash
   python get-policy-ad.py
   ```

4. **Reload dcvix-director**

   After the file is written, send a SIGHUP to the director process to
   pick up the changes without restarting:

   ```bash
   kill -HUP $(pidof dcvix-director)
   ```

## Output format

The generated `policydb/users.json` follows the JSON schema expected by
`internal/database/policy.go`:

```json
[
    {
        "id": "jdoe",
        "admin": false,
        "workstations": ["ws1.domain.com", "ws2.domain.com"],
        "pools": ["testing"]
    }
]
```

The admin flag and server lists come from the `<dcvix>` tag embedded in
each AD user's Description attribute.

## Notes

- This is **example code** meant to be adapted to your environment.
- Only users whose Description contains a `<dcvix>` tag are included.
- The LDAP filter `(&(objectClass=user)(objectCategory=person)(description=*<dcvix>*))`
  minimises the result set, but ensure your AD allows this kind of search.
- If you want to use a different attribute from description, adjust the `attributes` list
  and the `search_filter` in the script.
- The `pools.json` file must be managed separately if you reference pools
  in user entries.
