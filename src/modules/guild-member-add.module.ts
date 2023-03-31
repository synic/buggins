import { DiscordModule } from '@ao/types';
import { Client, Events, GuildMember } from 'discord.js';

export class GuildMemberAddModule extends DiscordModule {
  constructor(client: Client) {
    super(client);

    client.on(Events.GuildMemberAdd, async (member: GuildMember) => {
      console.log(`${member.displayName} has joined`);

      const role = member.guild.roles.cache.find(
        (r) => r.name.toLowerCase() === 'photographers',
      );
      if (!role) {
        console.log('Could not find photographers role, not adding.');
        return;
      }

      try {
        await member.roles.add(role);
      } catch (error) {
        console.log(`error adding user to role: ${error}`);
      }
    });
  }
}
