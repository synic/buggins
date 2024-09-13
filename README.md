Discord bot for Macromania
----

Usage:

bot start -c config.yaml

**Configuration Example**

```yaml
---
token: [your-discord-token]
database_url: db.sqlite  # this is the default location
modules:
  inatobs:
    page_size: 1
    channels:
      - id: 1373934330235659869
        inat_project_id: 146454
        cron_pattern: "0 * * * *"  # every hour is the default
  thisthat:
    channels:
      - id: 1273994333335659869
  featured:
    guilds:
      - name: macromania  # can be any name, not used in the app
        id: 1036730444714232627
        channel_id: 1273394190334659869
        reaction_count: 2  # default is 6
  inatlookup:
    guilds:
      - name: macromania
        id: 1213734464740232627
        command_prefix: ","
        channels:
          - all  # can be a list of channel ids or "all"
```
