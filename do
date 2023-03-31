#!/usr/bin/env python3

"""Macromania Discord Bot!"""

import os
import subprocess
import sys

if not os.path.isfile("./lib/dockerman/dockerman/__init__.py"):
    print("`dockerman` not found; run `git submodule update --init`")
    sys.exit(1)

sys.path.append("./lib/dockerman")

import dockerman as dm  # noqa: E402

# enable buildkit
os.environ["DOCKER_BUILDKIT"] = "1"
os.environ["COMPOSE_DOCKER_CLI_BUILD"] = "1"


class Containers:
    api = "buggins-bot"
    db = "buggins-db"


cont = Containers()


@dm.command(passthrough=True)
def bash(opts):
    """Bash shell on the api container."""
    dm.crun("bash", opts.args)


@dm.command(passthrough=True)
def shell(opts):
    """Open a python shell on the api container."""
    dm.crun("yarn console", opts.args)


@dm.command()
def start(opts):
    """Start all services."""
    data = subprocess.check_output(["docker", "network", "list", "-f", "name=buggins"])
    if "defiq" not in data.decode():
        dm.run("docker network create buggins")
    dm.run("docker-compose up -d")


@dm.command(passthrough=True)
def logs(opts):
    """Show logs for main api container."""
    dm.run(f"docker logs -f -n 1000 {cont.api}", opts.args)


@dm.command()
def stop(opts):
    """Stop all services."""
    dm.run("docker-compose stop")


@dm.command()
def db(opts):
    """Execute a database shell."""
    dm.crun("psql -U buggins buggins", container=cont.db)


@dm.command()
def debug(opts):
    """Attach to api container for debugging."""
    dm.warning(f"Attaching to `{cont.api}`. Type CTRL-p CTRL-q to exit.")
    dm.warning("CTRL-c will restart the container.")
    dm.run(f"docker attach {cont.api}")


@dm.command(passthrough=True)
def lint(opts):
    """Lint the code."""
    dm.crun("yarn lint", opts.args)


@dm.command(passthrough=True)
def typeorm(opts):
    """Run migration commands."""
    dm.crun("yarn typeorm:cli", opts.args)


@dm.command()
def migrate(opts):
    """Run all migrations."""
    dm.crun("yarn typeorm:cli migration:run", opts.args)


@dm.command(default=True, passthrough=True)
def yarn(opts):
    """Run yarn commands."""
    dm.crun("yarn", opts.args)


@dm.command(passthrough=True)
def manage(opts):
    """Run management commands."""
    dm.crun("yarn manage", opts.args)


@dm.command(dm.option("-n", "--name", help="migration file base name"))
def createmigration(opts):
    """Create a migration with a name."""
    dm.crun(
        f"yarn typeorm:plaincli migration:create "
        f"./src/databases/migrations/default/{opts.name}",
        opts.args,
    )


@dm.command(dm.option("-n", "--name", help="migration file base name"))
def generatemigration(opts):
    """Generate a migration with a name."""
    dm.crun(
        f"yarn typeorm:cli migration:generate "
        f"./src/databases/migrations/default/{opts.name}",
        opts.args,
    )


if __name__ == "__main__":
    module = sys.modules[__name__]
    splash = "\n".join(module.__doc__.split("\n")[1:-1])
    dm.main(default_container=cont.api, splash=splash)
